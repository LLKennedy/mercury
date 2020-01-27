package httpgrpc

import (
	"context"
	"net/http"

	"github.com/LLKennedy/httpgrpc/convert"
	"github.com/LLKennedy/httpgrpc/logs"
	"github.com/LLKennedy/httpgrpc/proxy"
	"google.golang.org/grpc"
)

// ProxyRequest proxies an HTTP request through a GRPC connection compliant with httpgrpc/proto
func ProxyRequest(ctx context.Context, w http.ResponseWriter, r *http.Request, procedure string, conn *grpc.ClientConn, txid string, loggers ...logs.Writer) {
	convert.ProxyRequest(ctx, w, r, procedure, conn, txid, loggers...)
}

// NewServer creates a new server to convert httpgrpc/proto messages to service-specific messages
func NewServer(api interface{}, server interface{}, listener *grpc.Server) (*proxy.Server, error) {
	return proxy.NewServer(api, server, listener)
}
