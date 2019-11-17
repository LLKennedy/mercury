package proxy

import (
	"context"
	"fmt"
	"reflect"

	"github.com/LLKennedy/httpgrpc"
	"google.golang.org/grpc"
)

// Proxy proxies connections through the server
func (s *Server) Proxy(ctx context.Context, req *httpgrpc.Request) (*httpgrpc.Response, error) {
	wrapErr := func(err error) error {
		return fmt.Errorf("httpgrpc: %v", err)
	}
	procedure, err := s.findProc(req.GetMethod(), req.GetProcedure())
	if err != nil {
		return &httpgrpc.Response{
			StatusCode: 404,
		}, wrapErr(err)
	}
	procType := procedure.Type
	if procType.NumIn() == 3 && // Check this matches standard GRPC method
		procType.IsVariadic() && // inputs should be *struct, grpc.CallOption...
		procType.In(1).Kind() == reflect.Ptr &&
		procType.In(1).Elem().Kind() == reflect.Struct &&
		procType.In(2).Implements(reflect.TypeOf((*grpc.CallOption)(nil)).Elem()) &&
		procType.NumOut() == 2 && // outputs should be *struct, error
		procType.Out(0).Kind() == reflect.Ptr &&
		procType.Out(0).Elem().Kind() == reflect.Struct &&
		procType.Out(1).Kind() == reflect.Interface &&
		procType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		// This is a normal grpc rpc definition
		return (&httpgrpc.UnimplementedExposedServiceServer{}).Proxy(ctx, req)
	}
	return &httpgrpc.Response{
		StatusCode: 501, //Unimplemented
	}, wrapErr(fmt.Errorf("nonstandard grpc signature not implemented"))
}

func (s *Server) findProc(httpMethod httpgrpc.Method, procName string) (reflect.Method, error) {
	methodString, err := methodToString(httpMethod)
	if err != nil {
		return reflect.Method{}, err
	}
	methodMap, found := s.getAPI()[methodString]
	if !found {
		return reflect.Method{}, fmt.Errorf("no %s methods defined in api", methodString)
	}
	procedure, found := methodMap[procName]
	if !found {
		return reflect.Method{}, fmt.Errorf("no procedure %s defined for %s method in api", procName, httpMethod)
	}
	return procedure, nil
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
