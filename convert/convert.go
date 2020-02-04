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
	req.Payload = bodyBytes
	// Forward the actual GRPC request
	res, err := proto.NewExposedServiceClient(conn).Proxy(ctx, req)
	if err != nil {
		// GRPC call failed, let's log it, process an error status
		for _, logger := range loggers {
			logger.LogErrorf(txid, "httpgrpc: received error from target service: %v", err)
		}
		errStatus, ok := status.FromError(err)
		if !ok {
			// Can't get proper status code, return bad gateway
			w.WriteHeader(http.StatusBadGateway)
		} else {
			w.WriteHeader(GRPCStatusToHTTPStatusCode(errStatus.Code()))
		}
	} else {
		// No grpc error, get (presumably) success code from response
		w.WriteHeader(int(res.GetStatusCode()))
	}
	// Write response body
	w.Write(res.GetPayload())
	for name, values := range res.GetWriteHeaders() {
		for _, value := range values.GetValues() {
			w.Header().Add(name, value)
		}
	}
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

// GRPCStatusToHTTPStatusCode converts a GRPC status to an HTTP Status Code
func GRPCStatusToHTTPStatusCode(grpcStatus codes.Code) (c int) {
	c = http.StatusInternalServerError // Default to internal error in case something goes wrong
	switch grpcStatus {
	case codes.Aborted:
		c = http.StatusBadGateway
	case codes.AlreadyExists:
		c = http.StatusNotModified
	case codes.Canceled:
		c = http.StatusGatewayTimeout
	case codes.DataLoss:
		c = http.StatusBadGateway
	case codes.DeadlineExceeded:
		c = http.StatusGatewayTimeout
	case codes.FailedPrecondition:
		c = http.StatusPreconditionFailed
	case codes.Internal:
		c = http.StatusBadGateway
	case codes.InvalidArgument:
		c = http.StatusBadRequest
	case codes.NotFound:
		c = http.StatusNotFound
	case codes.OK:
		c = http.StatusOK
	case codes.OutOfRange:
		c = http.StatusBadRequest
	case codes.PermissionDenied:
		c = http.StatusForbidden
	case codes.ResourceExhausted:
		c = http.StatusTooManyRequests
	case codes.Unauthenticated:
		c = http.StatusUnauthorized
	case codes.Unavailable:
		c = http.StatusServiceUnavailable
	case codes.Unimplemented:
		c = http.StatusNotImplemented
	case codes.Unknown:
		c = http.StatusBadGateway
	default:
		c = http.StatusInternalServerError
	}
	return
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
