package proxy

import (
	"google.golang.org/grpc"
)

// We use defaultServer in the case that s is nil
var defaultServer = &Server{}

// Server is an HTTP to GRPC proxy server
type Server struct {
	grpcServer *grpc.Server
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
