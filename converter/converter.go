package converter

import (
	"net/http"

	"google.golang.org/grpc"
)

// ProxyRequest proxies an HTTP request through a GRPC connection compliant with httpgrpc/proto
func ProxyRequest(w http.ResponseWriter, r *http.Request, conn *grpc.ClientConn) {
	return
}
