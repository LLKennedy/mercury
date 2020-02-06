package proxy

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/LLKennedy/httpgrpc/proto"
	"github.com/stretchr/testify/assert"
)

type emptyThing struct{}

func TestServer_setAPIConfig(t *testing.T) {
	specificExposedThing := &exposedThingA{}
	tests := []struct {
		name        string
		s           *Server
		api         interface{}
		server      interface{}
		result      *Server
		expectedErr string
	}{
		{
			name:   "no functions",
			s:      &Server{},
			api:    &emptyThing{},
			server: &emptyThing{},
			result: &Server{
				api:         map[string]map[string]apiMethod{},
				innerServer: &emptyThing{},
			},
			expectedErr: "",
		},
		{
			name:   "matching",
			s:      &Server{},
			api:    specificExposedThing,
			server: &thingA{},
			result: &Server{
				api: map[string]map[string]apiMethod{
					"POST": {
						"DoThing": apiMethod{pattern: apiMethodPatternStructStruct, reflection: reflect.TypeOf(specificExposedThing).Method(0)},
					},
				},
				innerServer: &thingA{},
			},
			expectedErr: "",
		},
		{
			name:        "mismatch",
			s:           &Server{},
			api:         specificExposedThing,
			server:      &thingB{},
			result:      &Server{},
			expectedErr: "validation of PostDoThing to DoThing mapping: api and server arguments mismatch: int vs string",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.s.setAPIConfig(tt.api, tt.server)
			if tt.expectedErr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.result.grpcServer, tt.s.grpcServer)
				assert.Equal(t, tt.result.innerServer, tt.s.innerServer)
				if assert.Len(t, tt.s.api, len(tt.result.api)) {
					for key, inMap := range tt.s.api {
						for procName, method := range inMap {
							if _, exists := tt.result.api[key]; assert.True(t, exists) {
								if _, procExists := tt.result.api[key][procName]; assert.True(t, procExists) {
									expectedMethod := tt.result.api[key][procName]
									assert.Equal(t, method.reflection.Name, expectedMethod.reflection.Name)
									assert.Equal(t, method.reflection.PkgPath, expectedMethod.reflection.PkgPath)
									assert.Equal(t, method.reflection.Type, expectedMethod.reflection.Type)
									assert.Equal(t, method.reflection.Index, expectedMethod.reflection.Index)
								}
							}
						}
					}
				}
			} else {
				assert.EqualError(t, err, tt.expectedErr)
			}
		})
	}
}

func TestSetExceptionHandler(t *testing.T) {
	t.Run("use default handler", func(t *testing.T) {
		s := &Server{}
		handled, res, err := s.handleExceptions(nil, nil)
		assert.False(t, handled)
		assert.Nil(t, res)
		assert.NoError(t, err)
	})
	t.Run("simple exception handler catching all methods", func(t *testing.T) {
		s := &Server{}
		fixedResponse := &proto.Response{
			StatusCode: 111,
			Payload:    []byte("testdata"),
		}
		s.SetExceptionHandler(func(ctx context.Context, req *proto.Request) (handled bool, res *proto.Response, err error) {
			return true, fixedResponse, fmt.Errorf("some error")
		})
		res, err := s.Proxy(nil, nil)
		assert.Equal(t, fixedResponse, res)
		assert.EqualError(t, err, "some error")
	})
}
