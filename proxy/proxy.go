package proxy

import (
	"context"

	"github.com/LLKennedy/httpgrpc"
)

// Proxy proxies connections through the server
func (s *Server) Proxy(ctx context.Context, req *httpgrpc.Request) (*httpgrpc.Response, error) {
	innerMethod, err := s.fetchInnerMethod(req.GetMethod(), req.GetProcedure())
	if err != nil {
		return nil, err
	}
	_ = innerMethod // use reflection to call this
	return (&httpgrpc.UnimplementedExposedServiceServer{}).Proxy(ctx, req)
}
