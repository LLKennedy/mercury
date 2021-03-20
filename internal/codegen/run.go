package codegen

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/LLKennedy/mercury/internal/version"
	"github.com/LLKennedy/mercury/proxy"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

var protocVersion = "unknown"

var packageReplacement = regexp.MustCompile(`\.([a-zA-Z0-9_]+)\.(.*)`)

// At time of writing, the only feature that can be marked as supported is restoring the "optional" keyword to proto3, still an experimental feature that is out of scope for this project.
var support uint64 = uint64(pluginpb.CodeGeneratorResponse_FEATURE_NONE)

// Run performs code generation on the input data
func Run(request *pluginpb.CodeGeneratorRequest) (response *pluginpb.CodeGeneratorResponse) {
	defer func() {
		if r := recover(); r != nil {
			response = &pluginpb.CodeGeneratorResponse{
				SupportedFeatures: &support,
				Error:             proto.String(fmt.Sprintf("caught panic in protoc-gen-mercury: %v", r)),
			}
		}
	}()
	// Set runtime version of protoc
	protocVersion = version.FormatProtocVersion(request.GetCompilerVersion())
	// Create a basic response with our feature support (none, see above)
	response = &pluginpb.CodeGeneratorResponse{
		SupportedFeatures: &support,
	}
	// Make sure the request actually exists as a safeguard
	if request == nil {
		response.Error = proto.String("cannot generate from nil input")
		return
	}
	// Generate the files (do the thing)
	generatedFiles, err := generateAllFiles(request)
	if err != nil {
		// It didn't work, ignore any data we generated and only return the error
		response.Error = proto.String(fmt.Sprintf("failed to generate files: %v", err))
		return
	}
	// It worked, set the response data
	response.File = generatedFiles
	return
}

// Naive approach to codegen, creates output files for every message/service in every linked file, not just the parts depended on by the "to generate" files
func generateAllFiles(request *pluginpb.CodeGeneratorRequest) (outfiles []*pluginpb.CodeGeneratorResponse_File, err error) {
	var out *pluginpb.CodeGeneratorResponse_File
	var pkgMap map[string]string
	outfiles, pkgMap, err = generatePackages(request)
	if err != nil {
		outfiles = nil
		return
	}
	for _, file := range request.GetProtoFile() {
		out, err = generateFullFile(file, pkgMap)
		if err != nil {
			return
		}
		outfiles = append(outfiles, out)
	}
	return
}

func generatePackages(request *pluginpb.CodeGeneratorRequest) (out []*pluginpb.CodeGeneratorResponse_File, pkgMap map[string]string, err error) {
	pkgMap = make(map[string]string)
	packageNames := map[string][]string{}
	for _, file := range request.GetProtoFile() {
		if file.GetSyntax() != "proto3" {
			err = fmt.Errorf("proto3 is the only syntax supported by protoc-gen-mercury, found %s in %s", file.GetSyntax(), file.GetName())
			return
		}
		pkgName := file.GetPackage()
		if pkgName == "" {
			err = fmt.Errorf("packages are mandatory with protoc-gen-mercury, %s did not have a package", file.GetName())
			return
		}
		if pkgName == "index" {
			// TODO: check mercury here
			err = fmt.Errorf("for JS/TS language reasons, \"index\" is an invalid package name")
		}
		pkgMap[file.GetName()] = pkgName
		list, _ := packageNames[pkgName]
		list = append(list, file.GetName())
		packageNames[pkgName] = list
	}
	indexFile := &pluginpb.CodeGeneratorResponse_File{
		Name: proto.String("__packages__/mercury/index.ts"),
	}
	indexContent := &strings.Builder{}
	for pkgName, importList := range packageNames {
		outFile := &pluginpb.CodeGeneratorResponse_File{
			Name: proto.String(fmt.Sprintf("__packages__/%s_mercury.ts", pkgName)),
		}
		content := &strings.Builder{}
		for _, importFile := range importList {
			parsedName := filenameFromProto(importFile)
			content.WriteString(fmt.Sprintf("export * from \"%s_mercury\";\n", parsedName.fullWithoutExtension))
		}
		outFile.Content = proto.String(content.String())
		out = append(out, outFile)
		indexContent.WriteString(fmt.Sprintf("export * as %s from \"__packages__/%s_mercury\";\n", pkgName, pkgName))
	}
	indexFile.Content = proto.String(indexContent.String())
	out = append(out, indexFile)
	return
}

func generateFullFile(f *descriptorpb.FileDescriptorProto, pkgMap map[string]string) (out *pluginpb.CodeGeneratorResponse_File, err error) {
	if f.GetSyntax() != "proto3" {
		err = fmt.Errorf("proto3 is the only syntax supported by protoc-gen-mercury, found %s in %s", f.GetSyntax(), f.GetName())
		return
	}
	parsedName := filenameFromProto(f.GetName())
	out = &pluginpb.CodeGeneratorResponse_File{
		Name: proto.String(parsedName.fullWithoutExtension + "_mercury.ts"),
	}
	content := &strings.Builder{}
	content.WriteString(getCodeGenmarker(version.GetVersionString(), protocVersion, f.GetName()))
	// Imports
	content.WriteString("import * as packages from \"__packages__\";\n")
	content.WriteString("import * as mercury_packages from \"__packages__/mercury\";\n")
	content.WriteString("import * as mercury from \"@llkennedy/mercury\";\n")
	content.WriteString("\n")
	// Services
	generateServices(f.GetService(), content)
	// Messages
	generateMessages(f.GetMessageType(), content, f.GetPackage())
	// Comments? unclear how to link them back to other elements
	generateComments(f.GetSourceCodeInfo(), content)
	out.Content = proto.String(content.String())
	return
}

func generateMessages(messages []*descriptorpb.DescriptorProto, content *strings.Builder, pkgName string) {
	for _, message := range messages {
		// TODO: get comment data somehow
		comment := "A message"
		generateMessage(message, comment, message.GetName(), pkgName, content)
		for _, nestedType := range message.GetNestedType() {
			if !nestedType.GetOptions().GetMapEntry() {
				// TODO: get comment data somehow
				comment = "A message"
				name := fmt.Sprintf("%s__%s", message.GetName(), nestedType.GetName())
				generateMessage(nestedType, comment, name, pkgName, content)
			}
		}
	}
}

// /** A message */
// export class FibonacciResponse extends packages.service.FibonacciResponse implements mercury.ProtoJSONCompatible {
// 	public ToProtoJSON(): Object {
// 		return {
// 			number: mercury.ToProtoJSON.StringNumber(this.number),
// 		};
// 	}
// 	public static async Parse(data: any): Promise<FibonacciResponse> {
// 		let objData: Object = mercury.AnyToObject(data);
// 		let res = new FibonacciResponse();
// 		res.number = await mercury.Parse.Number(objData, "number", "number");
// 		return res;
// 	}
// }

func generateMessage(msg *descriptorpb.DescriptorProto, comment, name, pkgName string, content *strings.Builder) {
	content.WriteString(fmt.Sprintf("/** %s */\nexport class %s extends packages.%s.%s implements mercury.ProtoJSONCompatible {\n", comment, name, pkgName, name))
	protoJSONContent := &strings.Builder{}
	protoJSONContent.WriteString(`		return {
`)
	for _, field := range msg.GetField() {
		// FIXME: detect repeated/oneof
		switch field.GetType() {
		case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
			protoJSONContent.WriteString(fmt.Sprintf(`			%s: mercury.ToProtoJSON.Bool(this.%s),
`, field.GetJsonName(), field.GetJsonName()))
		case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
			protoJSONContent.WriteString(fmt.Sprintf(`			%s: mercury.ToProtoJSON.Bytes(this.%s),
`, field.GetJsonName(), field.GetJsonName()))
		case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE, descriptorpb.FieldDescriptorProto_TYPE_FLOAT, descriptorpb.FieldDescriptorProto_TYPE_FIXED32, descriptorpb.FieldDescriptorProto_TYPE_INT32, descriptorpb.FieldDescriptorProto_TYPE_SFIXED32, descriptorpb.FieldDescriptorProto_TYPE_SINT32:
			protoJSONContent.WriteString(fmt.Sprintf(`			%s: mercury.ToProtoJSON.Number(this.%s),
`, field.GetJsonName(), field.GetJsonName()))
		case descriptorpb.FieldDescriptorProto_TYPE_FIXED64, descriptorpb.FieldDescriptorProto_TYPE_SFIXED64, descriptorpb.FieldDescriptorProto_TYPE_UINT64, descriptorpb.FieldDescriptorProto_TYPE_SINT64, descriptorpb.FieldDescriptorProto_TYPE_INT64:
			protoJSONContent.WriteString(fmt.Sprintf(`			%s: mercury.ToProtoJSON.StringNumber(this.%s),
`, field.GetJsonName(), field.GetJsonName()))
		case descriptorpb.FieldDescriptorProto_TYPE_STRING:
			protoJSONContent.WriteString(fmt.Sprintf(`			%s: mercury.ToProtoJSON.String(this.%s),
`, field.GetJsonName(), field.GetJsonName()))
		case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
			// TODO
		case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
			// TODO
		}
	}
	protoJSONContent.WriteString(`		};`)
	parseContent := &strings.Builder{}
	parseContent.WriteString(fmt.Sprintf(`		let objData: Object = mercury.AnyToObject(data);
		let res = new %s();
`, name))
	for _, field := range msg.GetField() {
		// FIXME: detect repeated/oneof
		switch field.GetType() {
		case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
			parseContent.WriteString(fmt.Sprintf(`		res.%s = await mercury.Parse.Bool(objData, "%s", "%s");
`, field.GetJsonName(), field.GetJsonName(), field.GetName()))
		case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
			parseContent.WriteString(fmt.Sprintf(`		res.%s = await mercury.Parse.Bytes(objData, "%s", "%s");
`, field.GetJsonName(), field.GetJsonName(), field.GetName()))
		case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE, descriptorpb.FieldDescriptorProto_TYPE_FIXED32, descriptorpb.FieldDescriptorProto_TYPE_FIXED64, descriptorpb.FieldDescriptorProto_TYPE_FLOAT, descriptorpb.FieldDescriptorProto_TYPE_INT32, descriptorpb.FieldDescriptorProto_TYPE_INT64, descriptorpb.FieldDescriptorProto_TYPE_SFIXED32, descriptorpb.FieldDescriptorProto_TYPE_SFIXED64, descriptorpb.FieldDescriptorProto_TYPE_SINT32, descriptorpb.FieldDescriptorProto_TYPE_SINT64, descriptorpb.FieldDescriptorProto_TYPE_UINT32, descriptorpb.FieldDescriptorProto_TYPE_UINT64:
			parseContent.WriteString(fmt.Sprintf(`		res.%s = await mercury.Parse.Number(objData, "%s", "%s");
`, field.GetJsonName(), field.GetJsonName(), field.GetName()))
		case descriptorpb.FieldDescriptorProto_TYPE_STRING:
			parseContent.WriteString(fmt.Sprintf(`		res.%s = await mercury.Parse.String(objData, "%s", "%s");
`, field.GetJsonName(), field.GetJsonName(), field.GetName()))
		case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
			// TODO
		case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
			// TODO
		}
	}
	parseContent.WriteString(`		return res;`)

	content.WriteString(fmt.Sprintf(`	public ToProtoJSON(): Object {
%s
	}
	public static async Parse(data: any): Promise<%s> {
%s
	}
`, protoJSONContent.String(), name, parseContent.String()))
	content.WriteString("}\n\n")
}

func generateServices(services []*descriptorpb.ServiceDescriptorProto, content *strings.Builder) {
SERVICE_LOOP:
	for _, service := range services {
		innerContent := &strings.Builder{}
		// TODO: get comment data somehow
		comment := "A service client"
		innerContent.WriteString(fmt.Sprintf("/** %s */\nexport class %sClient extends mercury.Client {\n", comment, service.GetName()))
		for _, method := range service.GetMethod() {
			methodString, procName, valid := proxy.MatchAndStripMethodName(method.GetName())
			if !valid {
				log.Printf("Method %s on service %s does not match mercury exposed pattern, skipping service generation\n", method.GetName(), service.GetName())
				continue SERVICE_LOOP
			}
			reqMsgType := formatMessageTypeName(method.GetInputType())
			resMsgType := formatMessageTypeName(method.GetOutputType())
			clientStreaming := method.GetClientStreaming()
			serverStreaming := method.GetServerStreaming()
			switch {
			case clientStreaming && serverStreaming:
				// Dual streaming
				innerContent.WriteString(fmt.Sprintf(`	public async %s(): Promise<mercury.IDualStream<%s, %s>> {
		return this.StartDualStream<%s, %s>("/%s", %s.Parse);
	}
`, procName, reqMsgType, resMsgType, reqMsgType, resMsgType, procName, resMsgType))
			case clientStreaming:
				// Client streaming
				innerContent.WriteString(fmt.Sprintf(`	public async %s(): Promise<mercury.IClientStream<%s, %s>> {
		return this.StartClientStream<%s, %s>("/%s", %s.Parse);
	}
`, procName, reqMsgType, resMsgType, reqMsgType, resMsgType, procName, resMsgType))
			case serverStreaming:
				// Server streaming
				innerContent.WriteString(fmt.Sprintf(`	public async %s(req: %s): Promise<mercury.IServerStream<%s>> {
		return this.StartServerStream<%s, %s>("/%s", req, %s.Parse);
	}
`, procName, reqMsgType, resMsgType, reqMsgType, resMsgType, procName, resMsgType))
			default:
				// Unary
				innerContent.WriteString(fmt.Sprintf(`	public async %s(req: %s): Promise<%s> {
		return this.SendUnary<%s, %s>("/%s", mercury.HTTPMethod.%s, req, %s.Parse);
	}
`, procName, reqMsgType, resMsgType, reqMsgType, resMsgType, procName, methodString, resMsgType))

			}
		}
		innerContent.WriteString("}\n\n")
		// Commit
		content.WriteString(innerContent.String())
	}
}

func generateComments(sourceCodeInfo *descriptorpb.SourceCodeInfo, content *strings.Builder) {

}

func formatMessageTypeName(msgType string) string {
	matches := packageReplacement.FindStringSubmatch(msgType)
	if len(matches) != 3 {
		panic(fmt.Errorf("type name did not match any valid pattern: %s, found %d instead of 3: %s", msgType, len(matches), matches))
	}
	pkgSection := fmt.Sprintf("mercury_packages.%s.", matches[1])
	typeSection := strings.ReplaceAll(matches[2], ".", "__")
	return fmt.Sprintf("%s%s", pkgSection, typeSection)
}

func getProtoJSONTypeName(field *descriptorpb.FieldDescriptorProto, nestedTypes []*descriptorpb.DescriptorProto) string {
	panic("not implemented")
}
