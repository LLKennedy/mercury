package proxy

import (
	"context"
	"fmt"
	"reflect"

	"github.com/LLKennedy/httpgrpc/httpapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProxyServerStream takes a single client request then streams responses from the server
func (s *Server) ProxyServerStream(req *httpapi.Request, srv httpapi.ExposedService_ProxyServerStreamServer) (err error) {
	wrapErr := func(code codes.Code, err error) error {
		if err == nil {
			return nil
		}
		return status.Error(code, fmt.Sprintf("httpgrpc: %v", err))
	}
	defer func() {
		if r := recover(); r != nil {
			err = wrapErr(codes.Internal, fmt.Errorf("caught panic %v", r))
		}
	}()
	ctx := srv.Context()
	procType, caller, pattern, err := s.findProc(req.GetMethod(), req.GetProcedure())
	if err != nil {
		return wrapErr(codes.Unimplemented, err)
	}
	if pattern == apiMethodPatternUnknown {
		return wrapErr(codes.Unimplemented, fmt.Errorf("nonstandard grpc signature not implemented"))
	}
	if pattern != apiMethodPatternStructStream {
		return wrapErr(codes.InvalidArgument, fmt.Errorf("ProxyServerStream called for non-server-stream RPC"))
	}
	var inputJSON []byte
	inputJSON, err = parseRequest(req)
	if err != nil {
		return wrapErr(codes.Internal, err)
	}
	err = s.callStructStream(ctx, inputJSON, procType, caller, srv)
	return err
}

// One struct in, stream of structs out
func (s *Server) callStructStream(ctx context.Context, inputJSON []byte, procType reflect.Type, caller reflect.Value, srv httpapi.ExposedService_ProxyServerStreamServer) (err error) {
	return status.Error(codes.Unimplemented, fmt.Sprintf("httpgrpc: Struct In, Stream Out is not yet supported, please manually implement exceptions for endpoint %s", procType))
}
