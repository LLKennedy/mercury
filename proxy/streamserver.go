package proxy

import (
	"context"
	"fmt"
	"reflect"
	"runtime/debug"

	"github.com/LLKennedy/httpgrpc/httpapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// One struct in, stream of structs out
func (s *Server) handleServerStream(ctx context.Context, procType reflect.Type, caller reflect.Value, srv httpapi.ExposedService_ProxyStreamServer) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = status.Errorf(codes.Internal, "caught panic for server stream: %v", r)
			fmt.Printf("%s\n", debug.Stack())
		}
	}()
	var onlyUpMsg *httpapi.StreamedRequest
	onlyUpMsg, err = srv.Recv()
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "initial request message could not be received: %v", err)
	}
	onlyUpData := onlyUpMsg.GetRequest()
	onlyUpParsed := reflect.New(procType.In(1).Elem()).Interface().(proto.Message)
	err = unmarshaller.Unmarshal(onlyUpData, onlyUpParsed)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "could not parse input data to request message: %v", err)
	}
	// Client streaming always starts by passing the context and nothing else to receive a stream + error
	returnValues := caller.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(onlyUpParsed)})
	// Parse our return values
	var clientErr error
	endpoint := returnValues[0]
	if endpoint.CanInterface() {
		var ok bool
		_, ok = (endpoint.Interface()).(grpc.ClientStream)
		if !ok {
			clientErr = status.Errorf(codes.Internal, "response message could not be converted to grpc.ServerStream interface")
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
	// All worked as expected and without error, now we start proxying response messages
	recv := endpoint.MethodByName("Recv")
	var res proto.Message
	res, err = wrapRecv(recv)
	for err == nil {
		var data []byte
		data, err = marshaller.Marshal(res)
		if err != nil {
			return
		}
		err = srv.Send(&httpapi.StreamedResponse{
			Response: data,
		})
		if err != nil {
			return
		}
		res, err = wrapRecv(recv)
	}
	return
}
