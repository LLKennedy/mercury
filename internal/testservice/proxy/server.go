package proxy

import (
	"context"
	"fmt"
	"net/http"

	"github.com/LLKennedy/httpgrpc"
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
	fmt.Println("got HTTP request")
	txid := uuid.New().String()
	procedure := ""
	switch r.URL.Path {
	case "something":
		procedure = "something"
	default:
		fmt.Printf("method: %s\n", r.URL.Path)

	}
	httpgrpc.ProxyRequest(context.Background(), w, r, procedure, h.conn, txid)
}
