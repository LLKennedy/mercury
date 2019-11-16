package proxy

import (
	"reflect"
	"testing"

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
				api:         map[string]map[string]reflect.Method{},
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
				api: map[string]map[string]reflect.Method{
					"POST": {
						"DoThing": reflect.TypeOf(specificExposedThing).Method(0),
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
			expectedErr: "httpgrpc: validation of PostDoThing to DoThing mapping: api and server arguments mismatch: int vs string",
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
							expectedMethod := tt.result.api[key][procName]
							assert.Equal(t, method.Name, expectedMethod.Name)
							assert.Equal(t, method.PkgPath, expectedMethod.PkgPath)
							assert.Equal(t, method.Type, expectedMethod.Type)
							assert.Equal(t, method.Index, expectedMethod.Index)
						}
					}
				}
			} else {
				assert.EqualError(t, err, tt.expectedErr)
			}
		})
	}
}
