package proxy

import (
	"context"
	"reflect"

	"github.com/LLKennedy/httpgrpc/proto"
	"google.golang.org/grpc"
)

// We use defaultServer in the case that s is nil
var defaultServer = &Server{}

// ExceptionHandler is an exception handler function
type ExceptionHandler func(ctx context.Context, req *proto.Request) (handled bool, res *proto.Response, err error)

// Server is an HTTP to GRPC proxy server
type Server struct {
	grpcServer       *grpc.Server
	api              map[string]map[string]reflect.Method // the api of innerServer
	innerServer      interface{}                          // the actual protobuf endpoints we want to use
	exceptionHandler ExceptionHandler
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
		return
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
		return
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
		return
	}
	s.innerServer = in
}

func (s *Server) handleExceptions(ctx context.Context, req *proto.Request) (handled bool, res *proto.Response, err error) {
	if s == nil || s.exceptionHandler == nil {
		handled = false
		return
	}
	return s.exceptionHandler(ctx, req)
}
