package proxy

import (
	"context"
	"io"
	"reflect"

	"github.com/LLKennedy/httpgrpc/httpapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Stream of structs in, one struct out
func (s *Server) callStreamStruct(ctx context.Context, procType reflect.Type, caller reflect.Value, srv httpapi.ExposedService_ProxyStreamServer) (err error) {
	defer func() {
		r := recover()
		if r != nil {

		}
	}()
	// Client streaming always starts by passing the context and nothing else to receive a stream + error
	returnValues := caller.Call([]reflect.Value{reflect.ValueOf(ctx)})
	// Parse our return values
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
	// All worked as expected and without error, now we start proxying request messages
	send := returnValues[0].MethodByName("Send")
	t := send.Type().NumIn
	_ = t
	var req *httpapi.StreamedRequest
	req, err = srv.Recv()
	for err == nil {
		// FIXME: marshal payload to real request here
		err = client.SendMsg(req)
		if err != nil {
			break
		}
		req, err = srv.Recv()
	}
	if err == io.EOF {
		err = client.CloseSend()
		if err != nil {
			return
		}
		recvAndCloseMethod := returnValues[0].MethodByName("CloseAndRecv")
		_ = recvAndCloseMethod
		// recvAndCloseMethod.Call()
		// err = client.RecvMsg()
	}
	return
}
