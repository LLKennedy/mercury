package proxy

import (
	"fmt"
	"reflect"
	"runtime/debug"

	"github.com/LLKennedy/mercury/httpapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// ProxyStream streams requests and responses in both directions in any order
func (s *Server) ProxyStream(srv httpapi.ExposedService_ProxyStreamServer) (err error) {
	wrapErr := func(code codes.Code, err error) error {
		if err == nil {
			return nil
		}
		return status.Error(code, fmt.Sprintf("mercury: %v", err))
	}
	defer func() {
		if r := recover(); r != nil {
			err = wrapErr(codes.Internal, fmt.Errorf("caught panic %v", r))
			fmt.Printf("%s\n", debug.Stack())
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
	switch pattern {
	case apiMethodPatternStreamStream:
		err = s.handleDualStream(ctx, procType, caller, srv)
	case apiMethodPatternStreamStruct:
		err = s.handleClientStream(ctx, procType, caller, srv)
	case apiMethodPatternStructStream:
		err = s.handleServerStream(ctx, procType, caller, srv)
	case apiMethodPatternStructStruct:
		err = wrapErr(codes.Unimplemented, fmt.Errorf("ProxyStream called for non-stream RPC"))
	case apiMethodPatternUnknown:
		fallthrough
	default:
		err = wrapErr(codes.Unimplemented, fmt.Errorf("nonstandard grpc signature not implemented"))
	}
	return err
}

func wrapRecv(recv reflect.Value) (proto.Message, error) {
	resVals := recv.Call(nil)
	errVal := resVals[1].Interface()
	if errVal != nil {
		return nil, errVal.(error)
	}
	return resVals[0].Interface().(proto.Message), nil
}
