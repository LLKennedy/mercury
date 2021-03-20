# HTTP -> gRPC Client

This is a client library for the [mercury](https://github.com/llkennedy/mercury) proxy system.

HTTP Requests and WebSockets wrapped by the clients created by this library will be compliant with the assumptions and patterns laid out in that system.

This is not at all required to use `mercury`, you can implement canonical JSON HTTP requests for gRPC endpoints trivially yourself, and the custom logic surrounding websockets -> gRPC streams isn't terribly complex (basically just "EOF" as a CloseSend equivalent), but it's designed to make life as easy as possible when using `mercury` and web-ui gRPC in general.