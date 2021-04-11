package convert

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/LLKennedy/mercury/httpapi"
	"github.com/LLKennedy/mercury/logs"
	"golang.org/x/net/websocket"
	"google.golang.org/grpc/status"
)

const (
	defaultBufferSize = 65536
	// EOFMessage is the EOF message websockets must send to imitate gRPC CloseSend()
	EOFMessage = "EOF"
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
		errWriter.writeWsErr("error initialising: ", err)
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
	up := make(chan error, 1)
	down := make(chan error, 1)
	go h.up(c, client, up)
	go h.down(c, client, down)
	select {
	case err := <-up:
		if err != nil && err != io.EOF {
			errWriter.writeWsErr("error on upstream: ", err)
			return
		}
		// Upstream is closed, just wait on downstream
		err = <-down
		if err == io.EOF {
			c.Write([]byte(EOFMessage))
		}
		if err != nil && err != io.EOF {
			errWriter.writeWsErr("error on downstream: ", err)
			return
		}
		c.WriteClose(http.StatusOK)
		return
	case err := <-down:
		if err == io.EOF {
			c.Write([]byte(EOFMessage))
		}
		if err != nil && err != io.EOF {
			errWriter.writeWsErr("error on downstream: ", err)
			return
		}
		// Downstream is closed, just wait on upstream
		err = <-up
		if err != nil && err != io.EOF {
			errWriter.writeWsErr("error on upstream: ", err)
			return
		}
		c.WriteClose(http.StatusOK)
		return
	}
}

// Messages from client to server
func (h *stream) up(c *websocket.Conn, client httpapi.ExposedService_ProxyStreamClient, out chan<- error) {
	defer close(out)
	for {
		bufferSize := h.readBufferSize
		if bufferSize <= 0 {
			bufferSize = defaultBufferSize
		}
		buffer := make([]byte, bufferSize)
		n, err := c.Read(buffer)
		if err != nil {
			out <- fmt.Errorf("reading from websocket: %v", err)
			return
		}
		msg := buffer[:n]
		if string(msg) == EOFMessage {
			client.CloseSend()
			out <- io.EOF
			return
		}
		err = client.Send(&httpapi.StreamedRequest{
			MessageType: &httpapi.StreamedRequest_Request{
				Request: msg,
			},
		})
		if err != nil {
			out <- fmt.Errorf("writing to service: %v", err)
			return
		}
	}
}

// Messages from server to client
func (h *stream) down(c *websocket.Conn, client httpapi.ExposedService_ProxyStreamClient, out chan<- error) {
	defer close(out)
	for {
		// Get a message from the server
		msg, err := client.Recv()
		if err != nil {
			if err == io.EOF {
				out <- err
				return
			}
			c.Write([]byte(EOFMessage))
			out <- fmt.Errorf("reading from service: %v", err)
			return
		}
		// Send the message to the client
		_, err = c.Write(msg.GetResponse())
		if err != nil {
			out <- fmt.Errorf("writing to websocket: %v", err)
			return
		}
	}
}

type errorWriter struct {
	c       *websocket.Conn
	loggers []logs.Writer
	txid    string
}

func (w errorWriter) writeWsErr(extraMessage string, err error) {
	// GRPC call failed, let's log it, process an error status
	for _, logger := range w.loggers {
		logger.LogErrorf(w.txid, "mercury: writing error to websocket: %v", err)
	}
	errStatus, ok := status.FromError(err)
	if !ok {
		// TODO: how to write these in such a way that parsers can catch them?
		// Can't get proper status code, return bad gateway
		w.c.Write([]byte(fmt.Sprintf("%s%v", extraMessage, err)))
		w.c.WriteClose(http.StatusBadGateway)
	} else {
		w.c.Write([]byte(fmt.Sprintf("%s%v", extraMessage, errStatus.Message())))
		w.c.WriteClose(GRPCStatusToHTTPStatusCode(errStatus.Code()))
	}
}
