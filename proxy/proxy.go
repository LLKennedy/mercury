package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/LLKennedy/httpgrpc/proto"
	"github.com/peterbourgon/mergemap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Proxy proxies connections through the server
func (s *Server) Proxy(ctx context.Context, req *proto.Request) (res *proto.Response, err error) {
	wrapErr := func(code codes.Code, err error) error {
		return status.Error(code, fmt.Sprintf("httpgrpc: %v", err))
	}
	defer func() {
		if r := recover(); r != nil {
			err = wrapErr(codes.Internal, fmt.Errorf("caught panic %v", r))
		}
	}()
	procType, caller, pattern, err := s.findProc(req.GetMethod(), req.GetProcedure())
	if err != nil {
		return &proto.Response{}, wrapErr(codes.Unimplemented, err)
	}
	if pattern == apiMethodPatternUnknown {
		return &proto.Response{}, wrapErr(codes.Unimplemented, fmt.Errorf("nonstandard grpc signature not implemented"))
	}
	return s.callProc(ctx, req, procType, caller, pattern)
}

func (s *Server) callProc(ctx context.Context, req *proto.Request, procType reflect.Type, caller reflect.Value, pattern apiMethodPattern) (*proto.Response, error) {
	// Create new instance of struct argument to pass into real implementation
	builtRequest := reflect.New(procType.In(2).Elem())
	builtRequestPtr := builtRequest.Interface()
	// We're going to rely on JSON unmarshalling logic to get data from the request to the new struct
	// Maybe one day there will be custom marshalling/unmarshalling capability here with tag parsing and method analysis for custom functions, but probably not
	// JSON tags are pretty much fit for purpose
	// We just have to make sure we have json to work with first
	inputJSON, err := parseRequest(req)
	if err != nil {
		return &proto.Response{
			StatusCode: 500,
		}, err
	}
	if inputJSON != nil {
		err = json.Unmarshal(inputJSON, builtRequestPtr)
		if err != nil {
			return &proto.Response{
				StatusCode: 400,
			}, fmt.Errorf("failed to marshal request data to procedure argument: %v", err)
		}
	}
	// Actually call the inner procedure
	returnValues := caller.Call([]reflect.Value{reflect.ValueOf(ctx), builtRequest})
	var outJSON []byte
	if returnValues[0].CanInterface() {
		outJSON, _ = json.Marshal(returnValues[0].Interface())
		if string(outJSON) == "null" {
			outJSON = nil
		}
	}
	if returnValues[1].CanInterface() {
		err, _ = returnValues[1].Interface().(error)
	}
	res := &proto.Response{
		Payload: outJSON,
	}
	if err == nil {
		res.StatusCode = 200 // TODO: parse status code specifically from outJSON
	} else {
		res.StatusCode = 500
	}
	return res, err
}

func parseRequest(req *proto.Request) (finalJSON []byte, err error) {
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
	case proto.Method_CONNECT, proto.Method_GET, proto.Method_HEAD, proto.Method_OPTIONS, proto.Method_TRACE:
		// No request body, only query params are possible
		if len(queryMap) > 0 {
			finalJSON, err = json.Marshal(queryMap)
			if err != nil {
				err = fmt.Errorf("failed to marshal query parameters to JSON: %v", err)
			}
		}
	case proto.Method_DELETE, proto.Method_PATCH, proto.Method_POST, proto.Method_PUT:
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

func (s *Server) findProc(httpMethod proto.Method, procName string) (procType reflect.Type, caller reflect.Value, pattern apiMethodPattern, err error) {
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
	caller = reflect.ValueOf(s.getInnerServer()).MethodByName(procName)
	return
}

func methodToString(in proto.Method) (out string, err error) {
	switch in {
	case proto.Method_GET:
		out = "GET"
	case proto.Method_HEAD:
		out = "HEAD"
	case proto.Method_POST:
		out = "POST"
	case proto.Method_PUT:
		out = "PUT"
	case proto.Method_DELETE:
		out = "DELETE"
	case proto.Method_CONNECT:
		out = "CONNECT"
	case proto.Method_OPTIONS:
		out = "OPTIONS"
	case proto.Method_TRACE:
		out = "TRACE"
	case proto.Method_PATCH:
		out = "PATCH"
	}
	if out == "" {
		err = fmt.Errorf("unknown HTTP method")
	}
	return
}
