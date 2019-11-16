package proxy

import (
	"fmt"
	"net"
	"reflect"

	"github.com/LLKennedy/httpgrpc"
	"google.golang.org/grpc"
)

// NewServer creates a proxy from HTTP(S) traffic to server using the methods defined by api
// api should be the Unimplemented<ServiceName> struct compiled by the protobuf. All methods defined on api MUST start with an HTTP method name
// server MUST implement the same methods as api without the prepended method names, though it may have others without exposing them to HTTP(S) traffic
func NewServer(api, server interface{}, listener *grpc.Server) (*Server, error) {
	s := new(Server)
	s.register(listener)
	return s, s.setAPIConfig(api, server)
}

// register registers the server
func (s *Server) register(listener *grpc.Server) {
	s.setGrpcServer(listener)
	httpgrpc.RegisterExposedServiceServer(s.getGrpcServer(), s)
}

// setAPIConfig validates and sets the inner api and endpoint config
func (s *Server) setAPIConfig(api, server interface{}) (err error) {
	wrapErr := func(in error) error {
		return fmt.Errorf("httpgrpc: %v", in)
	}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("httpgrpc: caught panic %v", r)
		}
	}()
	apiType := reflect.TypeOf(api)
	serverType := reflect.TypeOf(server)
	apiMethods := map[string]map[string]reflect.Method{}
	// Check every function defined on api
	for i := 0; i < apiType.NumMethod(); i++ {
		// Each function in api must map exactly to an equivalent on server with the HTTP method stripped off
		apiMethod := apiType.Method(i)
		methodString, procedureName, err := validateMethod(apiMethod, serverType)
		if err != nil {
			// one of the functions didn't match
			return wrapErr(err)
		}
		if _, exists := apiMethods[methodString]; !exists {
			apiMethods[methodString] = map[string]reflect.Method{}
		}
		apiMethods[methodString][procedureName] = apiMethod
	}
	// We know all api functions map to server functions, now hold onto the method list and server pointer for later
	s.setAPI(apiMethods)
	s.setInnerServer(server)
	return nil
}

// Serve starts the server
func (s *Server) Serve(listener net.Listener) error {
	return s.grpcServer.Serve(listener)
}
