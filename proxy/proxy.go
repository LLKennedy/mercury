package proxy

import (
	"context"

	"github.com/LLKennedy/httpgrpc"
)

// Proxy proxies connections through the server
func (s *Server) Proxy(ctx context.Context, req *httpgrpc.Request) (*httpgrpc.Response, error) {
	return (&httpgrpc.UnimplementedExposedServiceServer{}).Proxy(ctx, req)
}
