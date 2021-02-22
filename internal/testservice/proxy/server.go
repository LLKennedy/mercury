package proxy

import (
	"context"
	"net/http"
	"strings"

	"github.com/LLKennedy/httpgrpc/v2"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

// New creates a new proxy handle
func New(conn *grpc.ClientConn) *Handle {
	return &Handle{
		conn: conn,
	}
}

// Handle is a proxy handle
type Handle struct {
	conn *grpc.ClientConn
}

// Start starts the proxy
func (h *Handle) Start() error {
	return http.ListenAndServe("localhost:4848", h)
}

// ServeHTTP serves HTTP requests and proxies them to the service
func (h *Handle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	txid := uuid.New().String()
	procedure := strings.TrimLeft(r.URL.Path, "/")
	httpgrpc.ProxyRequest(context.Background(), w, r, procedure, h.conn, txid)
}
