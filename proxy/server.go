package proxy

import (
	"context"

	"github.com/LLKennedy/httpgrpc"
	"google.golang.org/grpc"
)

// Server is an HTTP to GRPC proxy server
type Server struct {
	grpcServer *grpc.Server
}

// NewServer creates a new HTTP to GRPC proxy server
func NewServer() *Server {
	s := new(Server)
	_ = httpgrpc.RegisterExposedServiceServer
	return s
}

// Register registers the server
func (s *Server) Register(opt ...grpc.ServerOption) {
	if s != nil {
		s.grpcServer = grpc.NewServer(opt...)
		go httpgrpc.RegisterExposedServiceServer(s.grpcServer, s)
	}
}

// Proxy proxies connections through the server
func (s *Server) Proxy(ctx context.Context, req *httpgrpc.Request) (*httpgrpc.Response, error) {
	return (&httpgrpc.UnimplementedExposedServiceServer{}).Proxy(ctx, req)
}
