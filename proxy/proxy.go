package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/LLKennedy/httpgrpc/v2/convert"
	"github.com/LLKennedy/httpgrpc/v2/httpapi"
	"github.com/peterbourgon/mergemap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ProxyUnary proxies connections through the server
func (s *Server) ProxyUnary(ctx context.Context, req *httpapi.Request) (res *httpapi.Response, err error) {
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

// ProxyDualStream streams requests and responses in both directions in any order
func (s *Server) ProxyDualStream(srv httpapi.ExposedService_ProxyDualStreamServer) (err error) {
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
		return &httpapi.Response{}, status.Error(codes.InvalidArgument, "httpgrpc: cannot convert json data to non-proto message using protojson")
	}
	if inputJSON == nil {
		inputJSON = []byte("{}")
	}
	unmarshaller := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: true,
	}
	marshaller := protojson.MarshalOptions{
		AllowPartial:    true,
		EmitUnpopulated: true,
	}
	err = unmarshaller.Unmarshal(inputJSON, builtRequestMessage)
	if err != nil {
		return &httpapi.Response{}, status.Error(codes.InvalidArgument, fmt.Sprintf("httpgrpc: %v", err))
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
		err = status.Errorf(codes.Internal, "httpgrpc: response error was not an error message?")
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
			res.Payload = []byte(fmt.Sprintf("httpgrpc: received non-gRPC error from endpoint: %v", err))
		} else {
			res.StatusCode = uint32(convert.GRPCStatusToHTTPStatusCode(sErr.Code()))
			res.Payload = []byte(sErr.Message())
		}
	}
	return
}

// One struct in, stream of structs out
func (s *Server) callStructStream(ctx context.Context, inputJSON []byte, procType reflect.Type, caller reflect.Value, srv httpapi.ExposedService_ProxyServerStreamServer) (err error) {
	return status.Error(codes.Unimplemented, fmt.Sprintf("httpgrpc: Struct In, Stream Out is not yet supported, please manually implement exceptions for endpoint %s", procType))
}

// Stream of structs in, one struct out
func (s *Server) callStreamStruct(ctx context.Context, procType reflect.Type, caller reflect.Value, srv httpapi.ExposedService_ProxyClientStreamServer) (err error) {
	return status.Error(codes.Unimplemented, fmt.Sprintf("httpgrpc: Stream In, Struct Out is not yet supported, please manually implement exceptions for endpoint %s", procType))
}

// Stram of structs in, stream of structs out
func (s *Server) callStreamStream(ctx context.Context, procType reflect.Type, caller reflect.Value, srv httpapi.ExposedService_ProxyDualStreamServer) (err error) {
	return status.Error(codes.Unimplemented, fmt.Sprintf("httpgrpc: Stream In, Stream Out is not yet supported, please manually implement exceptions for endpoint %s", procType))
}

func parseRequest(req *httpapi.Request) (finalJSON []byte, err error) {
	// First we convert query parameters to a map
	queryMap := map[string]interface{}{}
	for key, value := range req.GetParams() {
		// TODO: don't ignore/overwrite duplicate keys here
		passedValue := ""
		for _, passedValue = range value.GetValues() {
		}
		queryMap[key] = passedValue
	}
	switch req.GetMethod() {
	case httpapi.Method_CONNECT, httpapi.Method_GET, httpapi.Method_HEAD, httpapi.Method_OPTIONS, httpapi.Method_TRACE:
		// No request body, only query params are possible
		if len(queryMap) > 0 {
			finalJSON, err = json.Marshal(queryMap)
			if err != nil {
				err = fmt.Errorf("failed to marshal query parameters to JSON: %v", err)
			}
		}
	case httpapi.Method_DELETE, httpapi.Method_PATCH, httpapi.Method_POST, httpapi.Method_PUT:
		// Merge request body with query params
		bodyJSON := req.GetPayload()
		if bodyJSON != nil && len(queryMap) > 0 {
			var bodyMap map[string]interface{}
			err = json.Unmarshal(bodyJSON, &bodyMap)
			if err != nil {
				err = fmt.Errorf("failed to unmarshall request body JSON: %v", err)
				break
			}
			// Merge both maps, using request body's values on conflict
			mergedMaps := mergemap.Merge(queryMap, bodyMap)
			finalJSON, err = json.Marshal(mergedMaps)
		} else if bodyJSON != nil {
			finalJSON = bodyJSON
		} else if len(queryMap) > 0 {
			finalJSON, err = json.Marshal(queryMap)
		}
	default:
		// Invalid http method
		// It shouldn't be possible to hit this normally, we do validation before we reach this point
		err = fmt.Errorf("invalid http method")
	}
	return
}

func methodToString(in httpapi.Method) (out string, err error) {
	switch in {
	case httpapi.Method_GET:
		out = "GET"
	case httpapi.Method_HEAD:
		out = "HEAD"
	case httpapi.Method_POST:
		out = "POST"
	case httpapi.Method_PUT:
		out = "PUT"
	case httpapi.Method_DELETE:
		out = "DELETE"
	case httpapi.Method_CONNECT:
		out = "CONNECT"
	case httpapi.Method_OPTIONS:
		out = "OPTIONS"
	case httpapi.Method_TRACE:
		out = "TRACE"
	case httpapi.Method_PATCH:
		out = "PATCH"
	}
	if out == "" {
		err = fmt.Errorf("unknown HTTP method")
	}
	return
}
