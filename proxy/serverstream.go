package proxy

import (
	"context"
	"fmt"
	"reflect"

	"github.com/LLKennedy/httpgrpc/httpapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// One struct in, stream of structs out
func (s *Server) callStructStream(ctx context.Context, inputJSON []byte, procType reflect.Type, caller reflect.Value, srv httpapi.ExposedService_ProxyStreamServer) (err error) {
	return status.Error(codes.Unimplemented, fmt.Sprintf("httpgrpc: Struct In, Stream Out is not yet supported, please manually implement exceptions for endpoint %s", procType))
}
