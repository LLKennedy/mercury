package proxy

import (
	"fmt"
	"net"
	"reflect"

	"github.com/LLKennedy/httpgrpc"
	"github.com/LLKennedy/httpgrpc/internal/methods"
	"google.golang.org/grpc"
)

// NewServer creates a proxy from HTTP(S) traffic to server using the methods defined by api
// api should be the Unimplemented<ServiceName> struct compiled by the protobuf. All methods defined on api MUST start with an HTTP method name
// server MUST implement the same methods as api without the prepended method names, though it may have others without exposing them to HTTP(S) traffic
func NewServer(api, server interface{}, opt ...grpc.ServerOption) (*Server, error) {
	s := new(Server)
	s.register(opt...)
	return s, s.setAPI(api, server)
}

// register registers the server
func (s *Server) register(opt ...grpc.ServerOption) {
	s.setGrpcServer(grpc.NewServer(opt...))
	httpgrpc.RegisterExposedServiceServer(s.grpcServer, s)
}

// setAPI sets the server's
func (s *Server) setAPI(api, server interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("httpgrpc: caught panic %v", r)
		}
	}()
	apiType := reflect.TypeOf(api)
	serverType := reflect.TypeOf(server)
	apiMethods := make([]string, apiType.NumMethod())
	for i := range apiMethods {
		apiMethod := apiType.Method(i)
		name := apiMethod.Name
		trueName, valid := methods.MatchAndStrip(name)
		if !valid {
			err = fmt.Errorf("httpgrpc: %s does not begin with a valid HTTP method", name)
			break
		}
		serverMethod, found := serverType.MethodByName(trueName)
		if !found {
			err = fmt.Errorf("httpgrpc: server is missing method %s", trueName)
			break
		}
		expectedType := apiMethod.Type
		foundType := serverMethod.Type
		matching, matchErr := argsMatch(expectedType, foundType)
		if !matching {
			err = fmt.Errorf("httpgrpc: api/server arguments do not match for method (%s/%s): %v", name, trueName, matchErr)
			break
		}
	}
	if err == nil {
		// Set api details for this server
	}
	return
}

// Serve starts the server
func (s *Server) Serve(listener net.Listener) error {
	return s.grpcServer.Serve(listener)
}
