package codegen

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/LLKennedy/httpgrpc/internal/version"
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
				Error:             proto.String(fmt.Sprintf("caught panic in protoc-gen-httpgrpc: %v", r)),
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
			err = fmt.Errorf("proto3 is the only syntax supported by protoc-gen-httpgrpc, found %s in %s", file.GetSyntax(), file.GetName())
			return
		}
		pkgName := file.GetPackage()
		if pkgName == "" {
			err = fmt.Errorf("packages are mandatory with protoc-gen-httpgrpc, %s did not have a package", file.GetName())
			return
		}
		if pkgName == "index" {
			// TODO: check httpgrpc here
			err = fmt.Errorf("for JS/TS language reasons, \"index\" is an invalid package name")
		}
		pkgMap[file.GetName()] = pkgName
		list, _ := packageNames[pkgName]
		list = append(list, file.GetName())
		packageNames[pkgName] = list
	}
	indexFile := &pluginpb.CodeGeneratorResponse_File{
		Name: proto.String("__packages__/httpgrpc/index.ts"),
	}
	indexContent := &strings.Builder{}
	for pkgName, importList := range packageNames {
		outFile := &pluginpb.CodeGeneratorResponse_File{
			Name: proto.String(fmt.Sprintf("__packages__/%s_httpgrpc.ts", pkgName)),
		}
		content := &strings.Builder{}
		for _, importFile := range importList {
			parsedName := filenameFromProto(importFile)
			content.WriteString(fmt.Sprintf("export * from \"%s_httpgrpc\";\n", parsedName.fullWithoutExtension))
		}
		outFile.Content = proto.String(content.String())
		out = append(out, outFile)
		indexContent.WriteString(fmt.Sprintf("export * as %s from \"__packages__/%s_httpgrpc\";\n", pkgName, pkgName))
	}
	indexFile.Content = proto.String(indexContent.String())
	out = append(out, indexFile)
	return
}

func generateFullFile(f *descriptorpb.FileDescriptorProto, pkgMap map[string]string) (out *pluginpb.CodeGeneratorResponse_File, err error) {
	if f.GetSyntax() != "proto3" {
		err = fmt.Errorf("proto3 is the only syntax supported by protoc-gen-httpgrpc, found %s in %s", f.GetSyntax(), f.GetName())
		return
	}
	parsedName := filenameFromProto(f.GetName())
	out = &pluginpb.CodeGeneratorResponse_File{
		Name: proto.String(parsedName.fullWithoutExtension + "_httpgrpc.ts"),
	}
	content := &strings.Builder{}
	content.WriteString(getCodeGenmarker(version.GetVersionString(), protocVersion, f.GetName()))
	// Imports
	content.WriteString("import * as packages from \"__packages__\";\n\n")
	content.WriteString("import * as httpgrpc_packages from \"__packages__/httpgrpc\";\n\n")
	// Enums
	generateEnums(f.GetEnumType(), content)
	// Messages
	generateMessages(f.GetMessageType(), content, f.GetPackage())
	// Services
	generateServices(f.GetService(), content)
	// Comments? unclear how to link them back to other elements
	generateComments(f.GetSourceCodeInfo(), content)
	out.Content = proto.String(content.String())
	return
}

func generateEnums(enums []*descriptorpb.EnumDescriptorProto, content *strings.Builder) {
	for _, enum := range enums {
		// TODO: get comment data somehow
		comment := "An enum"
		content.WriteString(fmt.Sprintf("/** %s */\nexport enum %s {\n", comment, enum.GetName()))
		for _, value := range enum.GetValue() {
			// We don't bother stripping the trailing comma on the last enum element because Typescript doesn't care
			// TODO: get comment data somehow
			comment = "An enum value"
			if value.GetNumber() == 0 {
				// Special case for 0, as it doesn't get written by protojson since it's the default value
				content.WriteString(fmt.Sprintf("	/** %s */\n	%s = \"\",\n", comment, value.GetName()))
			} else {
				content.WriteString(fmt.Sprintf("	/** %s */\n	%s = \"%s\",\n", comment, value.GetName(), value.GetName()))
			}
		}
		content.WriteString("}\n\n")
	}
}

func generateMessages(messages []*descriptorpb.DescriptorProto, content *strings.Builder, pkgName string) {
	for _, message := range messages {
		// TODO: get comment data somehow
		comment := "A message"
		content.WriteString(fmt.Sprintf("/** %s */\nexport class %s extends Object {\n", comment, message.GetName()))
		for _, field := range message.GetField() {
			tsType := getNativeTypeName(field, message, pkgName)
			// TODO: get comment data somehow
			comment = "A field"
			content.WriteString(fmt.Sprintf("	/** %s */\n	public %s?: %s;\n", comment, field.GetName(), tsType))
		}
		content.WriteString("}\n\n")

		for _, nestedType := range message.GetNestedType() {
			if !nestedType.GetOptions().GetMapEntry() {
				// TODO: get comment data somehow
				comment = "A message"
				content.WriteString(fmt.Sprintf("/** %s */\nexport class %s__%s extends Object {\n", comment, message.GetName(), nestedType.GetName()))
				for _, nestedField := range nestedType.GetField() {
					tsType := getNativeTypeName(nestedField, nestedType, pkgName)
					// TODO: get comment data somehow
					comment = "A field"
					content.WriteString(fmt.Sprintf("	/** %s */\n	public %s?: %s;\n", comment, nestedField.GetName(), tsType))
				}
				content.WriteString("}\n\n")
			}
		}
	}
}

func generateServices(services []*descriptorpb.ServiceDescriptorProto, content *strings.Builder) {

}

func generateComments(sourceCodeInfo *descriptorpb.SourceCodeInfo, content *strings.Builder) {

}

func getNativeTypeName(field *descriptorpb.FieldDescriptorProto, message *descriptorpb.DescriptorProto, pkgName string) string {
	switch field.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE,
		descriptorpb.FieldDescriptorProto_TYPE_FLOAT,
		descriptorpb.FieldDescriptorProto_TYPE_INT64,
		descriptorpb.FieldDescriptorProto_TYPE_UINT64,
		descriptorpb.FieldDescriptorProto_TYPE_INT32,
		descriptorpb.FieldDescriptorProto_TYPE_FIXED64,
		descriptorpb.FieldDescriptorProto_TYPE_FIXED32,
		descriptorpb.FieldDescriptorProto_TYPE_UINT32,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED32,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED64,
		descriptorpb.FieldDescriptorProto_TYPE_SINT32,
		descriptorpb.FieldDescriptorProto_TYPE_SINT64:
		// Javascript only has one number format
		return "number"
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		return "boolean"
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		return "string"
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		return "Uint8Array"
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		// TODO: this lookup is not efficient, but it'll do for now. building a map of known types by name as we go would be good
		for _, nestedMessage := range message.GetNestedType() {
			// FIXME: it is possible for this to misfire at least sometimes, though we'll see if it particularly matters
			if nestedMessage.GetOptions().GetMapEntry() && strings.Contains(field.GetTypeName(), nestedMessage.GetName()) {
				keyType := getNativeTypeName(nestedMessage.GetField()[0], nil, pkgName)
				valType := getNativeTypeName(nestedMessage.GetField()[1], nil, pkgName)
				return fmt.Sprintf("Map<%s, %s>", keyType, valType)
			}
		}
		// Not a map
		fallthrough
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		typeName := field.GetTypeName()
		matches := packageReplacement.FindStringSubmatch(typeName)
		if len(matches) != 3 {
			panic(fmt.Errorf("type name did not match any valid pattern: %s, found %d instead of 3: %s", typeName, len(matches), matches))
		}
		pkgSection := fmt.Sprintf("packages.%s.", matches[1])
		typeSection := strings.ReplaceAll(matches[2], ".", "__")
		return fmt.Sprintf("%s%s", pkgSection, typeSection)
	default:
		panic(fmt.Errorf("unknown field type: %s", field))
	}
}

func getProtoJSONTypeName(field *descriptorpb.FieldDescriptorProto, nestedTypes []*descriptorpb.DescriptorProto) string {
	panic("not implemented")
}
