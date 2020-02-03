package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"

	"github.com/LLKennedy/httpgrpc"
	"google.golang.org/grpc"
)

// Handle is an example GRPC server for a microservice
type Handle struct {
	server *grpc.Server
	proxy  interface{ Serve(net.Listener) error }
	photos map[string][]byte
}

// New creates a new server
func New() (*Handle, error) {
	s := new(Handle)
	s.server = grpc.NewServer()
	var err error
	s.proxy, err = httpgrpc.NewServer(&UnimplementedExposedServiceServer{}, s, s.server)
	if err != nil {
		return nil, err
	}
	RegisterServiceServer(s.server, s)
	return s, nil
}

// Start starts the server
func (h *Handle) Start() error {
	listener, err := net.Listen("tcp", ":8953")
	if err != nil {
		return err
	}
	go func() {
		h.proxy.Serve(listener)
	}()
	return h.server.Serve(listener)
}

// Stop stops the server
func (h *Handle) Stop() {
	h.server.GracefulStop()
}

// MakeClientConn returns a client connection to this service
func (h *Handle) MakeClientConn() (*grpc.ClientConn, error) {
	return grpc.Dial(":8953", grpc.WithInsecure())
}

// Fibonacci returns the nth number in the Fibonacci sequence. It does not start with an HTTP method and is therefore not exposed
func (h *Handle) Fibonacci(ctx context.Context, in *FibonacciRequest) (*FibonacciResponse, error) {
	n := in.GetN()
	if n == 0 {
		return nil, fmt.Errorf("testservice/fibonacci: n must be greater than zero")
	}
	prev := uint64(0)
	current := uint64(1)
	for i := uint64(1); i < n; i++ {
		new := prev + current
		prev = current
		current = new
	}
	return &FibonacciResponse{
		Number: current,
	}, nil
}

// Random returns a random integer in the desired range. It may be accessed via a Get request to the proxy at, for example, /api/Service/Random
func (h *Handle) Random(ctx context.Context, in *RandomRequest) (*RandomResponse, error) {
	// ISO standard
	return &RandomResponse{Number: 4}, nil
}

// UploadPhoto allows the upload of a photo to some persistence store. It may be accessed via  Post request to the proxy at, for example, /api/Service/UploadPhoto
func (h *Handle) UploadPhoto(ctx context.Context, in *UploadPhotoRequest) (*UploadPhotoResponse, error) {
	if h.photos == nil {
		h.photos = map[string][]byte{}
	}
	hasher := sha256.New()
	_, err := hasher.Write(in.GetData())
	if err != nil {
		return nil, err
	}
	hash := hex.EncodeToString(hasher.Sum(nil))
	_, found := h.photos[hash]
	if found {
		return nil, fmt.Errorf("photo already exists")
	}
	h.photos[hash] = in.GetData()
	return &UploadPhotoResponse{
		Uuid: hash,
	}, nil
}
