package proxy

import (
	"context"
	"fmt"

	"github.com/LLKennedy/httpgrpc"
)

// Proxy proxies connections through the server
func (s *Server) Proxy(ctx context.Context, req *httpgrpc.Request) (*httpgrpc.Response, error) {
	wrapErr := func(err error) error {
		return fmt.Errorf("httpgrpc: %v", err)
	}
	methodString, err := methodToString(req.GetMethod())
	if err != nil {
		return nil, wrapErr(err)
	}
	methodMap, found := s.getAPI()[methodString]
	if !found {
		return nil, wrapErr(fmt.Errorf("no %s methods defined in api", methodString))
	}
	method, found := methodMap[req.GetProcedure()]
	if !found {
		return nil, wrapErr(fmt.Errorf("no procedure %s defined for %s method in api", req.GetProcedure(), req.GetMethod()))
	}
	_ = method
	return (&httpgrpc.UnimplementedExposedServiceServer{}).Proxy(ctx, req)
}

func methodToString(in httpgrpc.Method) (out string, err error) {
	switch in {
	case httpgrpc.Method_GET:
		out = "GET"
	case httpgrpc.Method_HEAD:
		out = "HEAD"
	case httpgrpc.Method_POST:
		out = "POST"
	case httpgrpc.Method_PUT:
		out = "PUT"
	case httpgrpc.Method_DELETE:
		out = "DELETE"
	case httpgrpc.Method_CONNECT:
		out = "CONNECT"
	case httpgrpc.Method_OPTIONS:
		out = "OPTIONS"
	case httpgrpc.Method_TRACE:
		out = "TRACE"
	case httpgrpc.Method_PATCH:
		out = "PATCH"
	}
	if out == "" {
		err = fmt.Errorf("unknown HTTP method")
	}
	return
}
