package main

import fmt "fmt"

// Server is an example GRPC server for a microservice
type Server struct {
}

// NewServer creates a new server
func NewServer() *Server {
	s := new(Server)
	return s
}

// Start starts the server
func (s *Server) Start() error {
	return fmt.Errorf("not implemented")
}
