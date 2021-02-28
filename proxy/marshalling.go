package proxy

import "google.golang.org/protobuf/encoding/protojson"

var unmarshaller = protojson.UnmarshalOptions{
	AllowPartial:   true,
	DiscardUnknown: true,
}
var marshaller = protojson.MarshalOptions{
	AllowPartial:    true,
	EmitUnpopulated: true,
}
