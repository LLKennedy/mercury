# HTTP -> gRPC Client

This is a client library for the [httpgrpc](https://github.com/llkennedy/httpgrpc) proxy system.

HTTP Requests and WebSockets wrapped by the clients created by this library will be compliant with the assumptions and patterns laid out in that system.

This is not at all required to use `httpgrpc`, you can implement canonical JSON HTTP requests for gRPC endpoints trivially yourself, and the custom logic surrounding websockets -> gRPC streams isn't terribly complex (basically just "EOF" as a CloseSend equivalent), but it's designed to make life as easy as possible when using `httpgrpc` and web-ui gRPC in general.