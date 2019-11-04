package proxy

// Server is an HTTP to GRPC proxy server
type Server struct {
}

// NewServer creates a new HTTP to GRPC proxy server
func NewServer() *Server {
	s := new(Server)
	return s
}
