package mercury

import (
	"context"
	"net/http"

	"github.com/LLKennedy/mercury/convert"
	"github.com/LLKennedy/mercury/logs"
	"github.com/LLKennedy/mercury/proxy"
	"google.golang.org/grpc"
)

// ProxyRequest proxies an HTTP request through a GRPC connection compliant with mercury/proto
func ProxyRequest(ctx context.Context, w http.ResponseWriter, r *http.Request, procedure string, conn *grpc.ClientConn, txid string, loggers ...logs.Writer) {
	convert.ProxyRequest(ctx, w, r, procedure, conn, txid, loggers...)
}

// NewServer creates a new server to convert mercury/proto messages to service-specific messages
func NewServer(api interface{}, listener *grpc.Server) (*proxy.Server, error) {
	s, err := proxy.NewServer(api, listener)
	return s, err
}
