package proxy

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"runtime/debug"

	"github.com/LLKennedy/httpgrpc/httpapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Stream of structs in, one struct out
func (s *Server) callStreamStruct(ctx context.Context, procType reflect.Type, caller reflect.Value, srv httpapi.ExposedService_ProxyStreamServer) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = status.Errorf(codes.Internal, "caught panic: %v", r)
			fmt.Printf("%s\n", debug.Stack())
		}
	}()
	// Client streaming always starts by passing the context and nothing else to receive a stream + error
	returnValues := caller.Call([]reflect.Value{reflect.ValueOf(context.Background())})
	// Parse our return values
	var clientErr error
	var client grpc.ClientStream
	endpoint := returnValues[0]
	if endpoint.CanInterface() {
		var ok bool
		client, ok = (endpoint.Interface()).(grpc.ClientStream)
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
	send := endpoint.MethodByName("Send")
	recv := endpoint.MethodByName("CloseAndRecv")
	sendT := send.Type()
	reqT := sendT.In(0)
	var req *httpapi.StreamedRequest
	req, err = srv.Recv()
	for err == nil {
		msg := reflect.New(reqT.Elem()).Interface().(proto.Message)
		err = unmarshaller.Unmarshal(req.GetRequest(), msg)
		if err != nil {
			break
		}
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
		resVals := recv.Call(nil)
		res := resVals[0].Interface().(proto.Message)
		err = resVals[1].Interface().(error)
		if err != nil {
			return
		}
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
	}
	return
}
