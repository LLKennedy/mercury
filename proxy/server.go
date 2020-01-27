package proxy

import (
	"reflect"

	"google.golang.org/grpc"
)

// We use defaultServer in the case that s is nil
var defaultServer = &Server{}

// Server is an HTTP to GRPC proxy server
type Server struct {
	grpcServer  *grpc.Server
	api         map[string]map[string]reflect.Method // the api of innerServer
	innerServer interface{}                          // the actual protobuf endpoints we want to use
}

func (s *Server) getGrpcServer() *grpc.Server {
	if s == nil {
		return defaultServer.grpcServer
	}
	return s.grpcServer
}

func (s *Server) setGrpcServer(in *grpc.Server) {
	if s == nil {
		defaultServer.grpcServer = in
	}
	s.grpcServer = in
}

func (s *Server) getAPI() map[string]map[string]reflect.Method {
	if s == nil {
		return defaultServer.api
	}
	return s.api
}

func (s *Server) setAPI(in map[string]map[string]reflect.Method) {
	if s == nil {
		defaultServer.api = in
	}
	s.api = in
}

func (s *Server) getInnerServer() interface{} {
	if s == nil {
		return defaultServer.innerServer
	}
	return s.innerServer
}

func (s *Server) setInnerServer(in interface{}) {
	if s == nil {
		defaultServer.innerServer = in
	}
	s.innerServer = in
}
