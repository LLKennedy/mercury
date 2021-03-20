# Mercury

[![GoDoc](https://godoc.org/github.com/LLKennedy/mercury?status.svg)](https://godoc.org/github.com/LLKennedy/mercury)
[![Build Status](https://travis-ci.org/disintegration/imaging.svg?branch=master)](https://travis-ci.org/LLKennedy/mercury)
![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/LLKennedy/mercury.svg)
[![Coverage Status](https://coveralls.io/repos/github/LLKennedy/mercury/badge.svg?branch=master)](https://coveralls.io/github/LLKennedy/mercury?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/LLKennedy/mercury)](https://goreportcard.com/report/github.com/LLKennedy/mercury)
[![Maintainability](https://api.codeclimate.com/v1/badges/07f4a4d2a6a69c182e6c/maintainability)](https://codeclimate.com/github/LLKennedy/mercury/maintainability)
[![GitHub](https://img.shields.io/github/license/LLKennedy/mercury.svg)](https://github.com/LLKennedy/mercury/blob/master/LICENSE)

## HTTP -> gRPC = H -> g = Hg = Mercury

A Golang API and Client + Typescript Browser Client to standardise conversion of HTTP requests to Google Remote Procedure Calls (GRPC). Allows a generic web proxy to talk to GRPC services via a standard message while still allowing each service to maintain its API using GRPC and protocol buffers.

There are multiple implemenations that follow this basic intent already (HTTP+JSON reverse proxied to GRPC) but assume each service is directly handling external HTTP traffic, rather than sitting behind load-balanced webservers in a DMZ somewhere separate to your nice safe application servers. For example, [here](https://github.com/grpc-ecosystem/grpc-gateway) and [here](https://github.com/weaveworks/common/tree/master/mercury).

This differs, in that your GRPC services are only expected to be handling GRPC. The logic used by the reverse proxy to determine where to send the message is up to you, this library does not cover service discovery or decoding the URL to extract data such as procedure or service names.

In order to expose a method to HTTP, the protobuf should define a service in which each method is prefixed by an HTTP method. The service application must then implement a version of that method without the method name. The easiest way to do this is to specify two service definitions, one with only exposed methods matching the internal methods, as below:

```proto
service App {
    rpc ListThings (ListThingsRequest) returns (ListThingsResponse) {}
    rpc RemoveThing (ThingID) returns (Thing) {}
    rpc AddThing (Thing) returns (ThingID) {}
}

service ExposedApp {
    rpc GetListThings (ListThingsRequest) returns (ListThingsResponse) {}
}
```

In this example, App would provide three GRPC endpoints, but only one would be exposed for HTTP methods from the proxy. Important to note is that you don't need to actually implement GetListThings in this example, simply defining it will allow mercury to lookup the ListThings method on your real server.

### Limitation: Streamed RPC HTTP Methods

Due to the usage of WebSockets as the underlying implementation of all kinds of streamed endpoints, all RPCs with "stream" request or response messages **must** be exposed via the "Get" HTTP method, or not exposed at all. This is because the WebSocket handshake always begins with a GET request and then upgrades out of standard HTTP traffic, so there is no possibility of routing on any other method.

## Installation

As a Go dependency:

`go get "github.com/LLKennedy/mercury"`

As a JS/TS dependency:

`npm install @llkennedy/mercury`

## Basic Usage

### In your Typescript UI

#### Protobuf to TS types

<!-- TODO -->
##### Installation of protoc-gen-mercury
`go install github.com/LLKennedy/mercury/cmd/protoc-gen-mercury`

### In Your Web Proxy

```golang
import "github.com/LLKennedy/mercury"

...

func (ws *WebServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Handle other endpoints before this, ProxyRequest will always write to w
    // Your URL decoding and service discovery goes here, filling the two variables below.
    var procedure string
    var clientConn *grpc.ClientConn
    // Optionally you can also provide a transaction ID, a unique string that identifies this request. Empty string can be provided.
    var string txid
    // This will automatically convert the request into a proxy message and send it to the client, which is assumed to implement mercury/proto.ExposedServiceServer
    // After converting and sending the message, it will parsse the response and write to w
    return mercury.ProxyRequest(context.Background(), w, r, procedure, clientConn, txid)
}
```

### In Your Application Service

```golang
import "github.com/LLKennedy/mercury"

...

// Handle is an example GRPC server for a microservice matching the example App protobuf from earlier - actual GRPC methods not defined in this example.
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
    // This is the magic bit - &UnimplementedExposed<Service>Server{} defines the API, then we pass it the handler (s) and a GRPC server.
    // The GRPC server can be the same one your app server is using, or a different one if you want to run the exposed endpoints on a different port or with different TLS settings.
    s.proxy, err = mercury.NewServer(&UnimplementedExposedAppServer{}, s, s.server)
    if err != nil {
        return nil, err
    }
    RegisterAppServer(s.server, s)
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
```

## Testing

On windows, the simplest way to test is to use the powershell script.

`./test.ps1`

To emulate the testing which occurs in build pipelines for linux and mac, run the following:

`go test ./... -race`
