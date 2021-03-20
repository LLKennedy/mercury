package proxy

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/LLKennedy/mercury/convert"
	"github.com/LLKennedy/mercury/httpapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// ProxyUnary proxies connections through the server
func (s *Server) ProxyUnary(ctx context.Context, req *httpapi.Request) (res *httpapi.Response, err error) {
	wrapErr := func(code codes.Code, err error) error {
		if err == nil {
			return nil
		}
		return status.Error(code, fmt.Sprintf("mercury: %v", err))
	}
	defer func() {
		if r := recover(); r != nil {
			err = wrapErr(codes.Internal, fmt.Errorf("caught panic %v", r))
		}
	}()
	var handled bool
	handled, res, err = s.handleExceptions(ctx, req)
	if handled {
		// The user-defined exception handler already processes this request, we don't have to deal with it
		return
	}
	procType, caller, pattern, err := s.findProc(req.GetMethod(), req.GetProcedure())
	if err != nil {
		return &httpapi.Response{}, wrapErr(codes.Unimplemented, err)
	}
	if pattern == apiMethodPatternUnknown {
		return &httpapi.Response{}, wrapErr(codes.Unimplemented, fmt.Errorf("nonstandard grpc signature not implemented"))
	}
	if pattern != apiMethodPatternStructStruct {
		return &httpapi.Response{}, wrapErr(codes.InvalidArgument, fmt.Errorf("ProxyUnary called for non-unary RPC"))
	}
	var inputJSON []byte
	inputJSON, err = parseRequest(req)
	if err != nil {
		return &httpapi.Response{}, wrapErr(codes.Internal, err)
	}
	res, err = s.callStructStruct(ctx, inputJSON, procType, caller)
	return res, err
}

func (s *Server) findProc(httpMethod httpapi.Method, procName string) (procType reflect.Type, caller reflect.Value, pattern apiMethodPattern, err error) {
	var methodString string
	methodString, err = methodToString(httpMethod)
	if err != nil {
		return
	}
	methodMap, found := s.getAPI()[methodString]
	if !found {
		err = fmt.Errorf("no %s methods defined in api", methodString)
		return
	}
	var apiProc apiMethod
	apiProc, found = methodMap[procName]
	if !found {
		err = fmt.Errorf("no procedure %s defined for %s method in api", procName, httpMethod)
		return
	}
	procType = apiProc.reflection.Type
	pattern = apiProc.pattern
	caller = apiProc.value
	return
}

// One struct in, one struct out
func (s *Server) callStructStruct(ctx context.Context, inputJSON []byte, procType reflect.Type, caller reflect.Value) (res *httpapi.Response, err error) {
	// Create new instance of struct argument to pass into real implementation
	builtRequest := reflect.New(procType.In(2).Elem())
	builtRequestPtr := builtRequest.Interface()
	builtRequestMessage, ok := builtRequestPtr.(proto.Message)
	if !ok {
		return &httpapi.Response{}, status.Error(codes.InvalidArgument, "mercury: cannot convert json data to non-proto message using protojson")
	}
	if inputJSON == nil {
		inputJSON = []byte("{}")
	}
	err = unmarshaller.Unmarshal(inputJSON, builtRequestMessage)
	if err != nil {
		return &httpapi.Response{}, status.Error(codes.InvalidArgument, fmt.Sprintf("mercury: %v", err))
	}
	if !s.getSkipForwardingMetadata() {
		incoming, ok := metadata.FromIncomingContext(ctx)
		if ok {
			ctx = metadata.NewOutgoingContext(ctx, incoming)
		}
	}
	var outJSON []byte
	var jsonErr error
	returnValues := caller.Call([]reflect.Value{reflect.ValueOf(ctx), builtRequest})
	if returnValues[0].CanInterface() {
		outMessage, ok := (returnValues[0].Interface()).(proto.Message)
		if ok {
			outJSON, jsonErr = marshaller.Marshal(outMessage)
		} else {
			jsonErr = status.Errorf(codes.Internal, "response message could not be converted to protMessage interface")
		}
	}
	if returnValues[1].CanInterface() {
		err, _ = returnValues[1].Interface().(error)
	} else {
		err = status.Errorf(codes.Internal, "mercury: response error was not an error message?")
	}
	// TODO: this error gets swallowed if the actual endpoint also returned an error, we should fix this somehow
	if jsonErr != nil && err == nil {
		outJSON = nil
		err = status.Errorf(codes.Internal, "could not marshal response message to JSON: %v", jsonErr)
	} else if jsonErr != nil || outJSON != nil && (len(outJSON) == 0 || string(outJSON) == "null" || string(outJSON) == "{}") {
		outJSON = nil
	}
	res = &httpapi.Response{
		Payload: outJSON,
	}
	if err == nil {
		res.StatusCode = http.StatusOK
	} else {
		sErr, ok := status.FromError(err)
		if !ok {
			res.StatusCode = http.StatusInternalServerError
			res.Payload = []byte(fmt.Sprintf("mercury: received non-gRPC error from endpoint: %v", err))
		} else {
			res.StatusCode = uint32(convert.GRPCStatusToHTTPStatusCode(sErr.Code()))
			res.Payload = []byte(sErr.Message())
		}
	}
	return
}
