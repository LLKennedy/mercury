package proxy

import (
	"context"
	"fmt"
	"reflect"

	"github.com/LLKennedy/httpgrpc/httpapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProxyStream streams requests and responses in both directions in any order
func (s *Server) ProxyStream(srv httpapi.ExposedService_ProxyStreamServer) (err error) {
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
	initMsg, err := srv.Recv()
	if err != nil {
		return wrapErr(codes.InvalidArgument, fmt.Errorf("could not receive initial routing message in ProxyClientStream: %v", err))
	}
	switch initMsg.GetMessageType().(type) {
	case *httpapi.StreamedRequest_Init:
		break
	case *httpapi.StreamedRequest_Request:
		return wrapErr(codes.InvalidArgument, fmt.Errorf("first message in ProxyClientStream must be Init"))
	}
	msg := initMsg.GetMessageType().(*httpapi.StreamedRequest_Init).Init
	procType, caller, pattern, err := s.findProc(msg.GetMethod(), msg.GetProcedure())
	if err != nil {
		return wrapErr(codes.Unimplemented, err)
	}
	if pattern == apiMethodPatternUnknown {
		return wrapErr(codes.Unimplemented, fmt.Errorf("nonstandard grpc signature not implemented"))
	}
	if pattern != apiMethodPatternStreamStream {
		return wrapErr(codes.InvalidArgument, fmt.Errorf("ProxyDualStream called for non-dual-stream RPC"))
	}
	err = s.callStreamStream(ctx, procType, caller, srv)
	return err
}

// Stram of structs in, stream of structs out
func (s *Server) callStreamStream(ctx context.Context, procType reflect.Type, caller reflect.Value, srv httpapi.ExposedService_ProxyStreamServer) (err error) {
	return status.Error(codes.Unimplemented, fmt.Sprintf("httpgrpc: Stream In, Stream Out is not yet supported, please manually implement exceptions for endpoint %s", procType))
}
