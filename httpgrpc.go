package httpgrpc

import (
	"net/http"

	"github.com/LLKennedy/httpgrpc/converter"
	"github.com/LLKennedy/httpgrpc/proxy"
	"google.golang.org/grpc"
)

// ProxyRequest proxies an HTTP request through a GRPC connection compliant with httpgrpc/proto
func ProxyRequest(w http.ResponseWriter, r *http.Request, conn *grpc.ClientConn) {
	converter.ProxyRequest(w, r, conn)
}

// NewServer creates a new server to convert httpgrpc/proto messages to service-specific messages
func NewServer(api interface{}, server interface{}, listener *grpc.Server) (*proxy.Server, error) {
	return proxy.NewServer(api, server, listener)
}
