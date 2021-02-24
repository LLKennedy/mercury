package httpgrpc

import (
	"context"
	"net/http"

	"github.com/LLKennedy/httpgrpc/v2/convert"
	"github.com/LLKennedy/httpgrpc/v2/logs"
	"github.com/LLKennedy/httpgrpc/v2/proxy"
	"google.golang.org/grpc"
)

// ProxyRequest proxies an HTTP request through a GRPC connection compliant with httpgrpc/proto
func ProxyRequest(ctx context.Context, w http.ResponseWriter, r *http.Request, procedure string, conn *grpc.ClientConn, txid string, loggers ...logs.Writer) {
	convert.ProxyRequest(ctx, w, r, procedure, conn, txid, loggers...)
}

// NewServer creates a new server to convert httpgrpc/proto messages to service-specific messages
func NewServer(api, server interface{}, listener *grpc.Server, bypassInterceptors bool) (*proxy.Server, error) {
	s, err := proxy.NewServer(api, server, listener, bypassInterceptors)
	return s, err
}
