package proxy

import (
	"context"
	"fmt"
	"reflect"

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

func (s *Server) fetchInnerMethod(methodType httpgrpc.Method, name string) (reflect.Method, error) {
	// innerName, valid := matchAndStrip(name)
	// if !valid {
	// 	return reflect.Method{}, fmt.Errorf("no HTTP method type prepending procedure name")
	// }
	// switch methodType {
	// case httpgrpc.Method_UNKNOWN:
	// 	return reflect.Method{}, fmt.Errorf("httpgrpc: unknown HTTP method")
	// case httpgrpc.Method_GET:
	// case httpgrpc.Method_HEAD:
	// case httpgrpc.Method_POST:
	// case httpgrpc.Method_PUT:
	// case httpgrpc.Method_DELETE:
	// case httpgrpc.Method_CONNECT:
	// case httpgrpc.Method_OPTIONS:
	// case httpgrpc.Method_TRACE:
	// case httpgrpc.Method_PATCH:
	// default:
	return reflect.Method{}, fmt.Errorf("httpgrpc: unknown HTTP method")
	// }
}
