package proxy

import (
	"context"
	"fmt"
	"io"
	"reflect"

	"github.com/LLKennedy/httpgrpc/httpapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProxyClientStream streams requests from the client and terminates in a single server response
func (s *Server) ProxyClientStream(srv httpapi.ExposedService_ProxyClientStreamServer) (err error) {
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
	if pattern != apiMethodPatternStreamStruct {
		return wrapErr(codes.InvalidArgument, fmt.Errorf("ProxyClientStream called for non-client-stream RPC"))
	}
	err = s.callStreamStruct(ctx, procType, caller, srv)
	return err
}

// Stream of structs in, one struct out
func (s *Server) callStreamStruct(ctx context.Context, procType reflect.Type, caller reflect.Value, srv httpapi.ExposedService_ProxyClientStreamServer) (err error) {
	returnValues := caller.Call([]reflect.Value{reflect.ValueOf(ctx)})
	var clientErr error
	var client grpc.ClientStream
	if returnValues[0].CanInterface() {
		var ok bool
		client, ok = (returnValues[0].Interface()).(grpc.ClientStream)
		if !ok {
			clientErr = status.Errorf(codes.Internal, "response message could not be converted to grpc.ClientStream interface")
		}
	}
	if returnValues[1].CanInterface() {
		err, _ = returnValues[1].Interface().(error)
	}
	if err != nil {
		_, ok := status.FromError(err)
		if !ok {
			err = status.Errorf(codes.Internal, "non-gRPC error returned when initiating stream: %v", err)
		}
		return
	}
	if clientErr != nil {
		err = clientErr
		return
	}
	var req *httpapi.StreamedRequest
	for req, err = srv.Recv(); err == nil; req, err = srv.Recv() {
		// FIXME: marshal payload to real request here
		err = client.SendMsg(req)
		if err != nil {
			break
		}
	}
	if err == io.EOF {
		err = client.CloseSend()
		if err != nil {
			return
		}
		recvAndCloseMethod := returnValues[0].MethodByName("CloseAndRecv")
		recvAndCloseMethod.Call()
		err = client.RecvMsg()
	}
	return
}
