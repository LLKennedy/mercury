package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/LLKennedy/httpgrpc"
	"github.com/peterbourgon/mergemap"
)

// Proxy proxies connections through the server
func (s *Server) Proxy(ctx context.Context, req *httpgrpc.Request) (res *httpgrpc.Response, err error) {
	wrapErr := func(err error) error {
		return fmt.Errorf("httpgrpc: %v", err)
	}
	defer func() {
		if r := recover(); r != nil {
			err = wrapErr(fmt.Errorf("caught panic %v", r))
		}
	}()
	procType, caller, err := s.findProc(req.GetMethod(), req.GetProcedure())
	if err != nil {
		return &httpgrpc.Response{
			StatusCode: 404,
		}, wrapErr(err)
	}
	if procType.NumIn() == 3 && // Check this matches standard GRPC method
		!procType.IsVariadic() && // inputs should be context.Context, *struct
		procType.In(1).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) &&
		procType.In(2).Kind() == reflect.Ptr &&
		procType.In(2).Elem().Kind() == reflect.Struct &&
		procType.NumOut() == 2 && // outputs should be *struct, error
		procType.Out(0).Kind() == reflect.Ptr &&
		procType.Out(0).Elem().Kind() == reflect.Struct &&
		procType.Out(1).Kind() == reflect.Interface &&
		procType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		// This is a normal grpc rpc definition
		return s.callProc(ctx, req, procType, caller)
	}
	// Currently we require that the methods match a standard auto-generated GRPC server method with no plugins that impact function signature.
	return &httpgrpc.Response{
		StatusCode: 501, //Unimplemented
	}, wrapErr(fmt.Errorf("nonstandard grpc signature not implemented"))
}

func (s *Server) callProc(ctx context.Context, req *httpgrpc.Request, procType reflect.Type, caller reflect.Value) (*httpgrpc.Response, error) {
	// Create new instance of struct argument to pass into real implementation
	builtRequest := reflect.New(procType.In(2).Elem())
	builtRequestPtr := builtRequest.Interface()
	// We're going to rely on JSON unmarshalling logic to get data from the request to the new struct
	// Maybe one day there will be custom marshalling/unmarshalling capability here with tag parsing and method analysis for custom functions, but probably not
	// JSON tags are pretty much fit for purpose
	// We just have to make sure we have json to work with first

	// First we convert query parameters to a map
	queryMap := map[string]interface{}{}
	for _, pair := range req.GetParams() {
		// TODO: don't ignore/overwrite duplicate keys here
		queryMap[pair.GetKey()] = pair.GetValue()
	}
	var finalJSON []byte
	var err error
	switch req.GetMethod() {
	case httpgrpc.Method_CONNECT, httpgrpc.Method_GET, httpgrpc.Method_HEAD, httpgrpc.Method_OPTIONS, httpgrpc.Method_TRACE:
		// No request body, only query params are possible
		if len(queryMap) > 0 {
			finalJSON, err = json.Marshal(queryMap)
			if err != nil {
				err = fmt.Errorf("failed to marshal query parameters to JSON: %v", err)
			}
		}
	case httpgrpc.Method_DELETE, httpgrpc.Method_PATCH, httpgrpc.Method_POST, httpgrpc.Method_PUT:
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
	if err != nil {
		return &httpgrpc.Response{
			StatusCode: 500,
		}, err
	}
	if finalJSON != nil {
		err = json.Unmarshal(finalJSON, builtRequestPtr)
		if err != nil {
			return &httpgrpc.Response{
				StatusCode: 400,
			}, fmt.Errorf("failed to marshal request data to procedure argument: %v", err)
		}
	}
	// Actually call the inner procedure
	returnValues := caller.Call([]reflect.Value{reflect.ValueOf(ctx), builtRequest})
	var outJSON []byte
	if returnValues[0].CanInterface() {
		outJSON, _ = json.Marshal(returnValues[0].Interface())
	}
	if returnValues[1].CanInterface() {
		err = returnValues[1].Interface().(error)
	}
	res := &httpgrpc.Response{
		Payload: outJSON,
	}
	if err == nil {
		res.StatusCode = 200 // TODO: parse status code specifically from outJSON
	} else {
		res.StatusCode = 500
	}
	return res, err
}

func (s *Server) findProc(httpMethod httpgrpc.Method, procName string) (procType reflect.Type, caller reflect.Value, err error) {
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
	var apiProc reflect.Method
	apiProc, found = methodMap[procName]
	if !found {
		err = fmt.Errorf("no procedure %s defined for %s method in api", procName, httpMethod)
		return
	}
	procType = apiProc.Type
	caller = reflect.ValueOf(s.getInnerServer()).MethodByName(procName)
	return
}

func methodToString(in httpgrpc.Method) (out string, err error) {
	switch in {
	case httpgrpc.Method_GET:
		out = "GET"
	case httpgrpc.Method_HEAD:
		out = "HEAD"
	case httpgrpc.Method_POST:
		out = "POST"
	case httpgrpc.Method_PUT:
		out = "PUT"
	case httpgrpc.Method_DELETE:
		out = "DELETE"
	case httpgrpc.Method_CONNECT:
		out = "CONNECT"
	case httpgrpc.Method_OPTIONS:
		out = "OPTIONS"
	case httpgrpc.Method_TRACE:
		out = "TRACE"
	case httpgrpc.Method_PATCH:
		out = "PATCH"
	}
	if out == "" {
		err = fmt.Errorf("unknown HTTP method")
	}
	return
}
