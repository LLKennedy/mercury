package proxy

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/LLKennedy/httpgrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestServer_Proxy(t *testing.T) {
	tests := []struct {
		name        string
		s           *Server
		ctx         context.Context
		req         *httpgrpc.Request
		want        *httpgrpc.Response
		expectedErr string
	}{
		{
			name: "standard grpc call unimplemented",
			s: &Server{
				api: map[string]map[string]reflect.Method{
					"POST": {
						"Proxy": reflect.TypeOf(&httpgrpc.UnimplementedExposedServiceServer{}).Method(0),
					},
				},
				innerServer: &httpgrpc.UnimplementedExposedServiceServer{},
			},
			ctx: new(mockContext),
			req: &httpgrpc.Request{
				Method:    httpgrpc.Method_POST,
				Procedure: "Proxy",
			},
			want:        &httpgrpc.Response{},
			expectedErr: "rpc error: code = Unimplemented desc = method Proxy not implemented",
		},
		{
			name:        "blank request",
			s:           &Server{},
			ctx:         new(mockContext),
			req:         &httpgrpc.Request{},
			want:        &httpgrpc.Response{},
			expectedErr: "httpgrpc: unknown HTTP method",
		},
		{
			name: "unregistered method",
			s:    &Server{},
			ctx:  new(mockContext),
			req: &httpgrpc.Request{
				Method: httpgrpc.Method_POST,
			},
			want:        &httpgrpc.Response{},
			expectedErr: "httpgrpc: no POST methods defined in api",
		},
		{
			name: "unregistered method",
			s: &Server{
				api: map[string]map[string]reflect.Method{
					"POST": {},
				},
			},
			ctx: new(mockContext),
			req: &httpgrpc.Request{
				Method:    httpgrpc.Method_POST,
				Procedure: "DoThing",
			},
			want:        &httpgrpc.Response{},
			expectedErr: "httpgrpc: no procedure DoThing defined for POST method in api",
		},
		{
			name: "non-standard grpc call",
			s: &Server{
				api: map[string]map[string]reflect.Method{
					"POST": {
						"DoThing": reflect.TypeOf(&thingA{}).Method(0),
					},
				},
				innerServer: &thingA{},
			},
			ctx: new(mockContext),
			req: &httpgrpc.Request{
				Method:    httpgrpc.Method_POST,
				Procedure: "DoThing",
			},
			want:        &httpgrpc.Response{},
			expectedErr: "httpgrpc: nonstandard grpc signature not implemented",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.Proxy(tt.ctx, tt.req)
			if tt.expectedErr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			} else {
				assert.EqualError(t, err, tt.expectedErr)
			}
		})
	}
}

// mockContext is an autogenerated mock type for the mockContext type
type mockContext struct {
	mock.Mock
}

// Deadline provides a mock function with given fields:
func (_m *mockContext) Deadline() (time.Time, bool) {
	ret := _m.Called()

	var r0 time.Time
	if rf, ok := ret.Get(0).(func() time.Time); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(time.Time)
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func() bool); ok {
		r1 = rf()
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

// Done provides a mock function with given fields:
func (_m *mockContext) Done() <-chan struct{} {
	ret := _m.Called()

	var r0 <-chan struct{}
	if rf, ok := ret.Get(0).(func() <-chan struct{}); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan struct{})
		}
	}

	return r0
}

// Err provides a mock function with given fields:
func (_m *mockContext) Err() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Value provides a mock function with given fields: key
func (_m *mockContext) Value(key interface{}) interface{} {
	ret := _m.Called(key)

	var r0 interface{}
	if rf, ok := ret.Get(0).(func(interface{}) interface{}); ok {
		r0 = rf(key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	return r0
}

func Test_methodToString(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name    string
		in      httpgrpc.Method
		wantOut string
		wantErr string
	}{
		{
			name:    "invalid",
			in:      httpgrpc.Method(100),
			wantOut: "",
			wantErr: "unknown HTTP method",
		},
		{
			name:    "unknown",
			in:      httpgrpc.Method_UNKNOWN,
			wantOut: "",
			wantErr: "unknown HTTP method",
		},
		{
			name:    "CONNECT",
			in:      httpgrpc.Method_CONNECT,
			wantOut: "CONNECT",
		},
		{
			name:    "DELETE",
			in:      httpgrpc.Method_DELETE,
			wantOut: "DELETE",
		},
		{
			name:    "GET",
			in:      httpgrpc.Method_GET,
			wantOut: "GET",
		},
		{
			name:    "HEAD",
			in:      httpgrpc.Method_HEAD,
			wantOut: "HEAD",
		},
		{
			name:    "OPTIONS",
			in:      httpgrpc.Method_OPTIONS,
			wantOut: "OPTIONS",
		},
		{
			name:    "PATCH",
			in:      httpgrpc.Method_PATCH,
			wantOut: "PATCH",
		},
		{
			name:    "POST",
			in:      httpgrpc.Method_POST,
			wantOut: "POST",
		},
		{
			name:    "PUT",
			in:      httpgrpc.Method_PUT,
			wantOut: "PUT",
		},
		{
			name:    "TRACE",
			in:      httpgrpc.Method_TRACE,
			wantOut: "TRACE",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOut, err := methodToString(tt.in)
			if tt.wantErr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantOut, gotOut)
			} else {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}
