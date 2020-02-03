package convert

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/LLKennedy/httpgrpc/logs"
	"github.com/LLKennedy/httpgrpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProxyRequest proxies an HTTP request through a GRPC connection compliant with httpgrpc/proto
func ProxyRequest(ctx context.Context, w http.ResponseWriter, r *http.Request, procedure string, conn *grpc.ClientConn, txid string, loggers ...logs.Writer) {
	req := RequestFromRequest(r)
	req.Procedure = procedure
	bodyBytes, err := ioutil.ReadAll(r.Body)
	for _, logger := range loggers {
		logger.LogErrorf(txid, "httpgrpc: failed to read body from HTTP request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	req.Payload = bodyBytes
	res, err := proto.NewExposedServiceClient(conn).Proxy(ctx, req)
	if err != nil {
		for _, logger := range loggers {
			logger.LogErrorf(txid, "httpgrpc: received error from target service: %v", err)
		}
		errStatus, ok := status.FromError(err)
		if !ok {
			// Can't get proper status code, return bad gateway
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		status.New(codes.FailedPrecondition, "test")
		statusCode := http.StatusInternalServerError // Default to 500 if we can't parse it here
		errStatus.Proto().GetCode()
		switch errStatus.Code() {
		case codes.Aborted:
		case codes.AlreadyExists:
		case codes.Canceled:
		case codes.DataLoss:
		case codes.DeadlineExceeded:
		case codes.FailedPrecondition:
		case codes.Internal:
		case codes.InvalidArgument:
		case codes.NotFound:
		case codes.OK:
		case codes.OutOfRange:
		case codes.PermissionDenied:
		case codes.ResourceExhausted:
		case codes.Unauthenticated:
		case codes.Unavailable:
		case codes.Unimplemented:
		case codes.Unknown:
			fallthrough
		default:
		}
		w.WriteHeader(statusCode)

	}
	for name, values := range res.GetWriteHeaders() {
		for _, value := range values.GetValues() {
			w.Header().Add(name, value)
		}
	}
	w.Write(res.GetPayload())
	w.WriteHeader(int(res.GetStatusCode()))
	return
}

// RequestFromRequest creates a *proto.Request from *http.Request filling all values except body, which could error
func RequestFromRequest(r *http.Request) *proto.Request {
	req := &proto.Request{}
	req.Headers = map[string]*proto.MultiVal{}
	for name, values := range r.Header {
		newHeader := &proto.MultiVal{}
		newHeader.Values = values
		req.Headers[name] = newHeader
	}
	req.Method = MethodFromString(r.Method)
	req.Params = map[string]*proto.MultiVal{}
	for name, values := range r.URL.Query() {
		newParam := &proto.MultiVal{}
		newParam.Values = values
		req.Params[name] = newParam
	}
	return req
}

// MethodFromString converts a method string to a proto.Method
func MethodFromString(methodString string) proto.Method {
	switch strings.ToUpper(methodString) {
	case "GET":
		return proto.Method_GET
	case "HEAD":
		return proto.Method_HEAD
	case "POST":
		return proto.Method_POST
	case "PUT":
		return proto.Method_PUT
	case "DELETE":
		return proto.Method_DELETE
	case "CONNECT":
		return proto.Method_CONNECT
	case "OPTIONS":
		return proto.Method_OPTIONS
	case "TRACE":
		return proto.Method_TRACE
	case "PATCH":
		return proto.Method_PATCH
	default:
		return proto.Method_UNKNOWN

	}
}
