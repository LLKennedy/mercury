package convert

import (
	"context"
	"fmt"
	"net/http"

	"github.com/LLKennedy/httpgrpc/httpapi"
	"github.com/LLKennedy/httpgrpc/logs"
	"golang.org/x/net/websocket"
	"google.golang.org/grpc/status"
)

const (
	defaultBufferSize = 65536
)

type stream struct {
	ctx            context.Context
	remote         httpapi.ExposedServiceClient
	loggers        []logs.Writer
	procedure      string
	headers        http.Header
	readBufferSize int
	txid           string
}

func (h stream) Serve(c *websocket.Conn) {
	errWriter := errorWriter{
		c:       c,
		loggers: h.loggers,
		txid:    h.txid,
	}
	client, err := h.remote.ProxyStream(h.ctx)
	if err != nil {
		errWriter.writeWsErr(fmt.Errorf("error initialising: %v", err))
		return
	}
	routingInfo := &httpapi.RoutingInformation{
		Method:    httpapi.Method_GET,
		Procedure: h.procedure,
	}
	routingInfo.Headers = map[string]*httpapi.MultiVal{}
	for name, values := range h.headers {
		newHeader := &httpapi.MultiVal{}
		newHeader.Values = values
		routingInfo.Headers[name] = newHeader
	}
	err = client.Send(&httpapi.StreamedRequest{
		MessageType: &httpapi.StreamedRequest_Init{
			Init: routingInfo,
		},
	})
	for {
		up := make(chan error, 1)
		down := make(chan error, 1)
		go h.up(c, client, up)
		go h.down(c, client, down)
		select {
		case err := <-up:
			if err != nil {
				errWriter.writeWsErr(fmt.Errorf("error on upstream: %v", err))
				return
			}
		case err := <-down:
			if err != nil {
				errWriter.writeWsErr(fmt.Errorf("error on downstream: %v", err))
				return
			}
		}
	}
}

// Messages from client to server
func (h *stream) up(c *websocket.Conn, client httpapi.ExposedService_ProxyStreamClient, out chan<- error) {
	defer close(out)
	bufferSize := h.readBufferSize
	if bufferSize <= 0 {
		bufferSize = defaultBufferSize
	}
	buffer := make([]byte, bufferSize)
	n, err := c.Read(buffer)
	if err != nil {
		out <- err
		return
	}
	msg := buffer[:n]
	err = client.Send(&httpapi.StreamedRequest{
		MessageType: &httpapi.StreamedRequest_Request{
			Request: msg,
		},
	})
	if err != nil {
		out <- err
		return
	}
}

// Messages from server to client
func (h *stream) down(c *websocket.Conn, client httpapi.ExposedService_ProxyStreamClient, out chan<- error) {
	defer close(out)
	// Get a message from the server
	msg, err := client.Recv()
	if err != nil {
		// FIXME: handle io.EOF here
		out <- err
		return
	}
	// Send the message to the client
	_, err = c.Write(msg.GetResponse())
	if err != nil {
		out <- err
	}
}

type errorWriter struct {
	c       *websocket.Conn
	loggers []logs.Writer
	txid    string
}

func (w errorWriter) writeWsErr(err error) {
	// GRPC call failed, let's log it, process an error status
	for _, logger := range w.loggers {
		logger.LogErrorf(w.txid, "httpgrpc: writing error to websocket: %v", err)
	}
	errStatus, ok := status.FromError(err)
	if !ok {
		// Can't get proper status code, return bad gateway
		w.c.WriteClose(http.StatusBadGateway)
	} else {
		w.c.WriteClose(GRPCStatusToHTTPStatusCode(errStatus.Code()))
	}
}
