package proxy

import (
	"fmt"
	"net"
	"reflect"

	"github.com/LLKennedy/httpgrpc/v2/httpapi"
	"google.golang.org/grpc"
)

// NewServer creates a proxy from HTTP(S) traffic to server using the methods defined by api
// api should be the Unimplemented<ServiceName> struct compiled by the protobuf. All methods defined on api MUST start with an HTTP method name
// server MUST implement the same methods as api without the prepended method names, though it may have others without exposing them to HTTP(S) traffic
func NewServer(api, server interface{}, listener *grpc.Server) (s *Server, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("httpgrpc: caught panic creating new server: %v", r)
		}
	}()
	wrapErr := func(in error) error {
		if in == nil {
			return nil
		}
		return fmt.Errorf("httpgrpc: %v", in)
	}
	s = new(Server)
	s.register(listener)
	err = wrapErr(s.setAPIConfig(api, server))
	return s, err
}

// SetExceptionHandler sets a function which is called before auto-proxying a request.
// This function may return handled = false to indicate it did not handle the request and it should be auto-proxied as usual.
// If handled is returned true, the proxy will assume the request has been handled already and will immediately return res and error
func (s *Server) SetExceptionHandler(handler ExceptionHandler) {
	s.exceptionHandler = handler
}

// register registers the server
func (s *Server) register(listener *grpc.Server) {
	s.setGrpcServer(listener)
	httpapi.RegisterExposedServiceServer(s.getGrpcServer(), s)
}

// setAPIConfig validates and sets the inner api and endpoint config
func (s *Server) setAPIConfig(api, server interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("caught panic %v", r)
		}
	}()
	apiType := reflect.TypeOf(api)
	serverType := reflect.TypeOf(server)
	apiMethods := map[string]map[string]apiMethod{}
	// Check every function defined on api
	for i := 0; i < apiType.NumMethod(); i++ {
		// Each function in api must map exactly to an equivalent on server with the HTTP method stripped off
		apiMethodReflection := apiType.Method(i)
		methodString, procedureName, pattern, err := validateMethod(apiMethodReflection, serverType)
		if err != nil {
			// one of the functions didn't match
			return err
		}
		if _, exists := apiMethods[methodString]; !exists {
			apiMethods[methodString] = map[string]apiMethod{}
		}
		apiMethods[methodString][procedureName] = apiMethod{pattern: pattern, reflection: apiMethodReflection}
	}
	// We know all api functions map to server functions, now hold onto the method list and server pointer for later
	s.setAPI(apiMethods)
	s.setInnerServer(server)
	return nil
}

// Serve starts the server
func (s *Server) Serve(listener net.Listener) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("httpgrpc: cannot serve on nil Server")
		}
	}()
	return s.grpcServer.Serve(listener)
}
