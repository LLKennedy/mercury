package main

import (
	context "context"
	"net"

	grpc "google.golang.org/grpc"
)

// Server is an example GRPC server for a microservice
type Server struct {
	server  *grpc.Server
	handler *Handler
}

// NewServer creates a new server
func NewServer() *Server {
	s := new(Server)
	s.server = grpc.NewServer()
	s.handler = new(Handler)
	RegisterServiceServer(s.server, s.handler)
	return s
}

// Start starts the server
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", ":8953")
	if err != nil {
		return err
	}
	return s.server.Serve(listener)
}

// Handler handles GRPC method calls
type Handler struct{}

// Fibonacci returns the nth number in the Fibonacci sequence. It does not start with an HTTP method and is therefore not exposed
func (h *Handler) Fibonacci(ctx context.Context, in *FibonacciRequest) (*FibonacciResponse, error) {
	return (&UnimplementedServiceServer{}).Fibonacci(ctx, in)
}

// GetRandom returns a random integer in the desired range. It may be accessed via a Get request to the proxy at, for example, /api/Service/Random
func (h *Handler) GetRandom(ctx context.Context, in *RandomRequest) (*RandomResponse, error) {
	return (&UnimplementedServiceServer{}).GetRandom(ctx, in)
}

// PostUploadPhoto allows the upload of a photo to some persistence store. It may be accessed via  Post request to the proxy at, for example, /api/Service/UploadPhoto
func (h *Handler) PostUploadPhoto(ctx context.Context, in *UploadPhotoRequest) (*UploadPhotoResponse, error) {
	return (&UnimplementedServiceServer{}).PostUploadPhoto(ctx, in)
}
