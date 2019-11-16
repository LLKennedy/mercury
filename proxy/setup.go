package proxy

import (
	"fmt"
	"net"
	"reflect"

	"github.com/LLKennedy/httpgrpc"
	"github.com/LLKennedy/httpgrpc/internal/methods/get"
	"google.golang.org/grpc"
)

// NewServer creates a proxy from HTTP(S) traffic to server using the methods defined by api
// api should be the Unimplemented<ServiceName> struct compiled by the protobuf. All methods defined on api MUST start with an HTTP method name
// server MUST implement the same methods as api, though it may have others without exposing them to HTTP(S) traffic
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
	apiMethods := make([]string, apiType.NumMethod())
	for i := range apiMethods {
		name := apiType.Method(i).Name
		switch {
		case get.Match(name):
			//good
		default:
			//bad
		}
		apiMethods[i] = name
	}
	serverType := reflect.TypeOf(server)
}

// Serve starts the server
func (s *Server) Serve(listener net.Listener) error {
	return s.grpcServer.Serve(listener)
}
