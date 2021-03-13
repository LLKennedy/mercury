package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/LLKennedy/httpgrpc/internal/codegen"
	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	// Read input file(s) from stdin
	reqData, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(fmt.Errorf("httpgrpc: reading from stdin: %v", err))
	}

	req := pluginpb.CodeGeneratorRequest{}
	err = proto.Unmarshal(reqData, &req)
	if err != nil {
		panic(fmt.Errorf("httpgrpc: parsing codegeneratorrequest: %v", err))
	}
	// Generate code from input, but safeguard for potential misbehaviour from our plugin code
	res := codegen.Run(&req)
	if res == nil {
		log.Printf("httpgrpc: codegen.Run returned an nil codegeneratorresponse, returning a static error.")
		res = &pluginpb.CodeGeneratorResponse{
			Error: proto.String("tsjson: codegen.Run incorrectly returned a nil response"),
		}
	}
	output, err := proto.Marshal(res)
	if err != nil {
		log.Printf("httpgrpc: codegen.Run returned an invalid codegeneratorresponse, returning a static error. error marshalling was: %v\n", err)
		res = &pluginpb.CodeGeneratorResponse{
			Error: proto.String("httpgrpc: codegen.Run incorrectly returned an invalid codegeneratorresponse"),
		}
		output, _ = proto.Marshal(res)
	}
	// Write results to sdout
	_, err = os.Stdout.Write(output)
	if err != nil {
		panic(fmt.Errorf("httpgrpc: writing output: %v", err))
	}
}
