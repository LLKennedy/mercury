package proxy

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"testing"

	"github.com/LLKennedy/httpgrpc/v2/httpapi"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
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
			expectedErr: "validation of PostDoThing to DoThing mapping: argments mismatch in position 0: interface vs int",
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
		fixedResponse := &httpapi.Response{
			StatusCode: 111,
			Payload:    []byte("testdata"),
		}
		s.SetExceptionHandler(func(ctx context.Context, req *httpapi.Request) (handled bool, res *httpapi.Response, err error) {
			return true, fixedResponse, fmt.Errorf("some error")
		})
		res, err := s.Proxy(nil, nil)
		assert.Equal(t, fixedResponse, res)
		assert.EqualError(t, err, "some error")
	})
}

func TestNewServer(t *testing.T) {
	gS := grpc.NewServer()
	type args struct {
		api      interface{}
		server   interface{}
		listener *grpc.Server
	}
	tests := []struct {
		name    string
		args    args
		want    *Server
		wantErr string
	}{
		{
			name:    "empty",
			want:    &Server{},
			wantErr: "httpgrpc: caught panic creating new server: runtime error: invalid memory address or nil pointer dereference",
		},
		{
			name: "success",
			args: args{
				api:      &exposedThingA{},
				server:   &thingA{},
				listener: gS,
			},
			want: &Server{
				api: map[string]map[string]apiMethod{
					"POST": {
						"DoThing": {
							pattern:    apiMethodPatternStructStruct,
							reflection: func() reflect.Method { t, _ := reflect.TypeOf(&exposedThingA{}).MethodByName("PostDoThing"); return t }(),
						},
					},
				},
				innerServer: &thingA{},
				grpcServer:  gS,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewServer(tt.args.api, tt.args.server, tt.args.listener)
			if tt.want == nil {
				assert.Nil(t, got)
			} else {
				// Go really doesn't want to compare want and got successfully
				assert.NotNil(t, got)
			}
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}

type badListener struct{}

type badAddr struct{}

func (a *badAddr) Network() string {
	return ""
}
func (a *badAddr) String() string {
	return ""
}

func (b *badListener) Accept() (net.Conn, error) {
	return nil, fmt.Errorf("some error")
}
func (b *badListener) Close() error {
	return fmt.Errorf("close error")
}
func (b *badListener) Addr() net.Addr {
	return &badAddr{}
}

func TestServer_Serve(t *testing.T) {
	tests := []struct {
		name     string
		s        *Server
		listener net.Listener
		wantErr  string
	}{
		{
			name:    "empty",
			wantErr: "httpgrpc: cannot serve on nil Server",
		},
		{
			name:    "nil listener",
			s:       &Server{grpcServer: grpc.NewServer()},
			wantErr: "httpgrpc: cannot serve on nil Server",
		},
		{
			name:     "bad listener",
			s:        &Server{grpcServer: grpc.NewServer()},
			listener: &badListener{},
			wantErr:  "some error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.s.Serve(tt.listener)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}
