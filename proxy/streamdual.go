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

// Stram of structs in, stream of structs out
func (s *Server) handleDualStream(ctx context.Context, procType reflect.Type, caller reflect.Value, srv httpapi.ExposedService_ProxyStreamServer) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = status.Errorf(codes.Internal, "caught panic for dual stream: %v", r)
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
	// All worked as expected and without error, now we start proxying messages in both directions
	send := endpoint.MethodByName("Send")
	recv := endpoint.MethodByName("Recv")
	sendT := send.Type()
	reqT := sendT.In(0).Elem()
	up := make(chan error, 1)
	down := make(chan error, 1)
	go s.up(client, reqT, srv, up)
	go s.down(recv, srv, down)
	select {
	case err = <-up:
		if err != nil {
			return
		}
		err = <-down
	case err = <-down:
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return
		}
		err = <-up
	}
	if err == io.EOF {
		err = nil
	}
	return
}

func (s *Server) up(client grpc.ClientStream, reqT reflect.Type, srv httpapi.ExposedService_ProxyStreamServer, done chan<- error) {
	defer close(done)
	req, err := srv.Recv()
	for err == nil {
		msg := reflect.New(reqT).Interface().(proto.Message)
		err = unmarshaller.Unmarshal(req.GetRequest(), msg)
		if err != nil {
			break
		}
		err = client.SendMsg(msg)
		req, err = srv.Recv()
	}
	if err == io.EOF {
		client.CloseSend()
		err = nil
	}
	done <- err
}

func (s *Server) down(recv reflect.Value, srv httpapi.ExposedService_ProxyStreamServer, done chan<- error) {
	defer close(done)
	res, err := wrapRecv(recv)
	var data []byte
	for err == nil {
		data, err = marshaller.Marshal(res)
		if err != nil {
			break
		}
		err = srv.Send(&httpapi.StreamedResponse{
			Response: data,
		})
		if err != nil {
			break
		}
		res, err = wrapRecv(recv)
	}
	// EOF here means the other end hung up, so we pass it along
	done <- err
}
