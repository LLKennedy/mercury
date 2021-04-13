package codegen

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/LLKennedy/mercury/internal/version"
	"github.com/LLKennedy/mercury/proxy"
	"github.com/LLKennedy/protoc-gen-tsjson/tsjsonpb"
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
				Error:             proto.String(fmt.Sprintf("caught panic in protoc-gen-tsjson: %v", r)),
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

const googlePrefix = "google."
const googleProtobufPrefix = "google.protobuf"

type exportDetails struct {
	npmPackage   string
	importPath   string
	protoPackage string
}

// Naive approach to codegen, creates output files for every message/service in every linked file, not just the parts depended on by the "to generate" files
func generateAllFiles(request *pluginpb.CodeGeneratorRequest) (outfiles []*pluginpb.CodeGeneratorResponse_File, err error) {
	var out *pluginpb.CodeGeneratorResponse_File
	var impexp importsExports
	impexp, err = buildImportsAndTypes(request.GetProtoFile())
	if err != nil {
		return nil, err
	}
	for _, file := range request.GetProtoFile() {
		if len(file.GetService()) <= 0 {
			continue
		}
		for _, toGen := range request.GetFileToGenerate() {
			if file.GetName() == toGen {
				out, err = generateFullFile(file, impexp)
				if err != nil {
					return
				}
				outfiles = append(outfiles, out)
				break
			}
		}
	}
	return
}

type importsExports struct {
	exportMap   map[string]exportDetails
	typeMap     map[string]map[string]exportDetails
	fileTypeMap map[string][]string
}

func buildImportsAndTypes(files []*descriptorpb.FileDescriptorProto) (impexp importsExports, err error) {
	// Map of file names to input paths
	impexp.exportMap = make(map[string]exportDetails, len(files))
	// Map of package names to type names to import details
	impexp.typeMap = make(map[string]map[string]exportDetails, len(files)) // Length here is just a starting value, not expected to be accurate
	// Map of file names to type names
	impexp.fileTypeMap = make(map[string][]string, len(files))
	// Map of
	// Check all files except google ones have both npm_package and import_path options set
	for _, file := range files {
		fileName := file.GetName()
		pkgName := file.GetPackage()
		if len(pkgName) >= len(googlePrefix) && pkgName[:len(googlePrefix)] == googlePrefix {
			// Google files are allowed to not have the options, we handle them differently
			continue
		}
		npmPackage, ok := proto.GetExtension(file.GetOptions(), tsjsonpb.E_NpmPackage).(string)
		if !ok || npmPackage == "" {
			err = fmt.Errorf("all non-google imports must specify the option (tsjson.npm_package), file %s did not", fileName)
			return
		}
		importPath, _ := proto.GetExtension(file.GetOptions(), tsjsonpb.E_ImportPath).(string)
		pkg := file.GetPackage()
		details := exportDetails{
			npmPackage:   npmPackage,
			importPath:   importPath,
			protoPackage: pkg,
		}
		impexp.exportMap[fileName] = details
		pkgTypes, ok := impexp.typeMap[pkg]
		if !ok {
			pkgTypes = make(map[string]exportDetails, len(file.GetEnumType())+len(file.GetMessageType()))
			impexp.typeMap[pkg] = pkgTypes
		}
		// Map out type defintions to packages for lookup later
		for _, enum := range file.GetEnumType() {
			parsedName := strings.ReplaceAll(enum.GetName(), ".", "__")
			pkgTypes[parsedName] = details
			impexp.fileTypeMap[fileName] = append(impexp.fileTypeMap[fileName], parsedName)
		}
		for _, msg := range file.GetMessageType() {
			parsedName := strings.ReplaceAll(msg.GetName(), ".", "__")
			pkgTypes[parsedName] = details
			impexp.fileTypeMap[fileName] = append(impexp.fileTypeMap[fileName], parsedName)
			for _, innerMsg := range msg.GetNestedType() {
				innerName := fmt.Sprintf("%s__%s", parsedName, strings.ReplaceAll(innerMsg.GetName(), ".", "__"))
				pkgTypes[innerName] = details
				impexp.fileTypeMap[fileName] = append(impexp.fileTypeMap[fileName], innerName)
			}
		}
	}
	return impexp, nil
}

func generateFullFile(f *descriptorpb.FileDescriptorProto, impexp importsExports) (out *pluginpb.CodeGeneratorResponse_File, err error) {
	fileName := f.GetName()
	if f.GetSyntax() != "proto3" {
		err = fmt.Errorf("proto3 is the only syntax supported by protoc-gen-tsjson, found %s in %s", f.GetSyntax(), fileName)
		return
	}
	parsedName := filenameFromProto(fileName)
	details, _ := impexp.exportMap[fileName]
	outName := details.importPath
	if outName == "" {
		outName = parsedName.fullWithoutExtension
	}
	out = &pluginpb.CodeGeneratorResponse_File{
		Name: proto.String(outName + "_mercury.ts"),
	}
	content := &strings.Builder{}
	content.WriteString(getCodeGenmarker(version.GetVersionString(), protocVersion, fileName))

	parsedMethods := map[string]map[string]parsedMethod{}
SERVICE_CHECK_LOOP:
	for _, service := range f.GetService() {
		for _, method := range service.GetMethod() {
			httpMethod, proc, valid := proxy.MatchAndStripMethodName(method.GetName())
			if !valid {
				log.Printf("Service %s did not match exposed patterns, skipping client generation for this service.\n", service.GetName())
				continue SERVICE_CHECK_LOOP
			}
			srvMethods, exists := parsedMethods[service.GetName()]
			if !exists {
				srvMethods = make(map[string]parsedMethod)
			}
			srvMethods[method.GetName()] = parsedMethod{
				method: httpMethod,
				proc:   proc,
			}
			parsedMethods[service.GetName()] = srvMethods
		}
	}
	// Imports
	generateImports(f, content, impexp, parsedMethods)
	// Services
	generateServices(f, content, impexp, parsedMethods)
	// Comments? unclear how to link them back to other elements
	generateComments(f.GetSourceCodeInfo(), content)
	out.Content = proto.String(content.String())
	return
}

func generateImports(f *descriptorpb.FileDescriptorProto, content *strings.Builder, impexp importsExports, parsedMethods map[string]map[string]parsedMethod) {
	if len(f.GetService()) > 0 {
		// All messages need the common imports
		content.WriteString("import * as mercury from \"@llkennedy/mercury\";\n")
	}
	importMap := make(map[string][]string)
	useGoogle := false
	for _, service := range f.GetService() {
		_, ok := parsedMethods[service.GetName()]
		if !ok {
			continue
		}
		for _, method := range service.GetMethod() {
			useGoogle = generateImportsForMethod(method, f.GetPackage(), f.GetName(), importMap, content, impexp) || useGoogle
		}
	}
	if useGoogle {
		content.WriteString("import { google } from \"@llkennedy/protoc-gen-tsjson\";\n")
	}
	for importPath, imports := range importMap {
		fullImportList := &strings.Builder{}
		for i, imp := range imports {
			if i != 0 {
				fullImportList.WriteString(",")
			}
			fullImportList.WriteString("\n	")
			fullImportList.WriteString(imp)
		}
		content.WriteString(fmt.Sprintf("import { %s\n} from \"%s\";\n", fullImportList.String(), importPath))
	}
	content.WriteString("\n")
}

func generateServices(f *descriptorpb.FileDescriptorProto, content *strings.Builder, impexp importsExports, parsedMethods map[string]map[string]parsedMethod) {
	for _, service := range f.GetService() {
		srvMethods, ok := parsedMethods[service.GetName()]
		if ok {
			generateService(f, service, content, impexp, srvMethods)
		}
	}
}

// import * as mercury from "@llkennedy/mercury";
// import { RandomRequest, RandomResponse, UploadPhotoRequest, UploadPhotoResponse, FeedData, FeedResponse, BroadcastRequest, BroadcastData, ConvertInput, ConvertOutput } from "service";

// export class ExposedAppClient extends mercury.Client {
// 	constructor(basePath: string | undefined = "localhost/api/ExposedApp", useTLS: boolean | undefined = true, client: mercury.AxiosInstance | undefined = undefined) {
// 		super(basePath, useTLS, client);
// 	}
// 	public async Random(req: RandomRequest): Promise<RandomResponse> {
// 		return this.SendUnary("Random", mercury.HTTPMethod.GET, req, RandomResponse.Parse);
// 	}
// 	public async UploadPhoto(req: UploadPhotoRequest): Promise<UploadPhotoResponse> {
// 		return this.SendUnary("UploadPhoto", mercury.HTTPMethod.POST, req, UploadPhotoResponse.Parse);
// 	}
// 	public async Feed(): Promise<mercury.ClientStream<FeedData, FeedResponse>> {
// 		return this.StartClientStream<FeedData, FeedResponse>("Feed", FeedResponse.Parse);
// 	}
// 	public async Broadcast(req: BroadcastRequest): Promise<mercury.ServerStream<BroadcastRequest, BroadcastData>> {
// 		return this.StartServerStream<BroadcastRequest, BroadcastData>("Broadcast", req, BroadcastData.Parse);
// 	}
// 	public async ConvertToString(): Promise<mercury.DualStream<ConvertInput, ConvertOutput>> {
// 		return this.StartDualStream<ConvertInput, ConvertOutput>("ConvertToString", ConvertOutput.Parse);
// 	}
// }

type parsedMethod struct {
	method string
	proc   string
}

func generateService(f *descriptorpb.FileDescriptorProto, service *descriptorpb.ServiceDescriptorProto, content *strings.Builder, impexp importsExports, parsedMethods map[string]parsedMethod) {
	name := service.GetName()
	// Create class definition and constructor
	content.WriteString(fmt.Sprintf(`export class %sClient extends mercury.Client {
	constructor(basePath: string | undefined = "localhost/api/%s", useTLS: boolean | undefined = true, client: mercury.AxiosInstance | undefined = undefined) {
		super(basePath, useTLS, client);
	}
`, name, name))
	for _, method := range service.GetMethod() {
		parseMethodTypeName := func(name string) string {
			typeName := strings.TrimLeft(name, ".")
			if len(typeName) >= len(googleProtobufPrefix) && typeName[:len(googleProtobufPrefix)] == googleProtobufPrefix {
				// This is a google well-known type
				return typeName
			}
			parts := strings.Split(name, ".")
			pkg := ""
			if len(parts) > 0 {
				pkg = strings.Join(parts[1:len(parts)-1], "__") + "__"
			}
			return pkg + parts[len(parts)-1]
		}
		reqT := parseMethodTypeName(method.GetInputType())
		resT := parseMethodTypeName(method.GetOutputType())
		methodName := method.GetName()
		clientStreaming := method.GetClientStreaming()
		serverStreaming := method.GetServerStreaming()
		parsed := parsedMethods[methodName]
		switch {
		case clientStreaming && serverStreaming:
			// Dual stream
			content.WriteString(fmt.Sprintf(`	public async %s(): Promise<mercury.DualStream<%s, %s>> {
		return this.StartDualStream<%s, %s>("%s", %s.Parse);
	}
`, parsed.proc, reqT, resT, reqT, resT, parsed.proc, resT))
		case clientStreaming:
			// Client stream
			content.WriteString(fmt.Sprintf(`	public async %s(): Promise<mercury.ClientStream<%s, %s>> {
		return this.StartClientStream<%s, %s>("%s", %s.Parse);
	}
`, parsed.proc, reqT, resT, reqT, resT, parsed.proc, resT))
		case serverStreaming:
			// Server stream
			content.WriteString(fmt.Sprintf(`	public async %s(req: %s): Promise<mercury.ServerStream<%s, %s>> {
		return this.StartServerStream<%s, %s>("%s", req, %s.Parse);
	}
`, parsed.proc, reqT, reqT, resT, reqT, resT, parsed.proc, resT))
		default:
			// Unary stream
			content.WriteString(fmt.Sprintf(`	public async %s(req: %s): Promise<%s> {
		return this.SendUnary("%s", mercury.HTTPMethod.%s, req, %s.Parse);
	}
`, parsed.proc, reqT, resT, parsed.proc, strings.ToUpper(parsed.method), resT))
		}
	}
	content.WriteString(`}
`)
}

func generateImportsForMethod(f *descriptorpb.MethodDescriptorProto, ownPkg string, fileName string, importMap map[string][]string, content *strings.Builder, impexp importsExports) (useGoogle bool) {
	for _, typeName := range []string{f.GetInputType(), f.GetOutputType()} {
		if typeName == "" {
			continue
		}
		typeName = strings.TrimLeft(typeName, ".")
		typeNameParts := strings.Split(typeName, ".")
		trueName := typeNameParts[len(typeNameParts)-1]
		pkgName := strings.TrimSuffix(typeName, "."+trueName)
		var importPath string
		if len(pkgName) >= len(ownPkg) && pkgName[:len(ownPkg)] == ownPkg {
			pkgName = ownPkg
			pkg, ok := impexp.typeMap[ownPkg]
			if !ok {
				panic(fmt.Sprintf("failed to find own package %s in imports for file %s", ownPkg, fileName))
			}
			trueName = typeName[len(ownPkg)+1:]
			parsedName := strings.ReplaceAll(trueName, ".", "__")
			// Exclude local messages/enums from import
			details, ok := pkg[parsedName]
			if !ok {
				panic(fmt.Sprintf("failed to find type %s in exports for package %s in file %s", trueName, pkgName, fileName))
			}
			importPath = details.importPath
		} else if pkgName == googleProtobufPrefix {
			useGoogle = true
			continue
		} else {
			pkg, ok := impexp.typeMap[pkgName]
			if !ok {
				panic(fmt.Sprintf("failed to find package %s in imports for file %s", pkgName, fileName))
			}
			details, ok := pkg[trueName]
			if !ok {
				panic(fmt.Sprintf("failed to find type %s in exports for package %s in file %s", trueName, pkgName, fileName))
			}
			importPath = fmt.Sprintf("%s/%s", details.npmPackage, details.importPath)
		}
		imports, _ := importMap[importPath]
		uniqueImports := map[string]struct{}{}
		for _, anImport := range imports {
			uniqueImports[anImport] = struct{}{}
		}
		uniqueImports[fmt.Sprintf("%s as %s__%s", trueName, pkgName, trueName)] = struct{}{}
		imports = []string{}
		for anImport := range uniqueImports {
			imports = append(imports, anImport)
		}
		importMap[importPath] = imports
	}
	return
}

type mapTypeData struct {
	toProtoJSON string
	parse       string
	keyIsString bool
}

func generateComments(sourceCodeInfo *descriptorpb.SourceCodeInfo, content *strings.Builder) {

}

func getNativeTypeName(field *descriptorpb.FieldDescriptorProto, message *descriptorpb.DescriptorProto, pkgName string, fileExports []string) string {
	repeatedStr := ""
	if field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED {
		repeatedStr = "[]"
	}
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
		return "number" + repeatedStr
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		return "boolean" + repeatedStr
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		return "string" + repeatedStr
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		return "Uint8Array" + repeatedStr
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		// TODO: this lookup is not efficient, but it'll do for now. building a map of known types by name as we go would be good
		for _, nestedMessage := range message.GetNestedType() {
			// FIXME: it is possible for this to misfire at least sometimes, though we'll see if it particularly matters
			if nestedMessage.GetOptions().GetMapEntry() && strings.Contains(field.GetTypeName(), nestedMessage.GetName()) {
				keyType := getNativeTypeName(nestedMessage.GetField()[0], nil, pkgName, fileExports)
				valType := getNativeTypeName(nestedMessage.GetField()[1], nil, pkgName, fileExports)
				return fmt.Sprintf("ReadonlyMap<%s, %s | null>", keyType, valType)
			}
		}
		fieldTypeName := strings.TrimLeft(field.GetTypeName(), ".")
		if len(fieldTypeName) >= len(googleProtobufPrefix) && fieldTypeName[:len(googleProtobufPrefix)] == googleProtobufPrefix {
			// This is a google well-known type
			return fieldTypeName + repeatedStr
		}
		// Not a map, not a google type
		fallthrough
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		typeName := field.GetTypeName()
		matches := packageReplacement.FindStringSubmatch(typeName)
		if len(matches) != 3 {
			panic(fmt.Errorf("type name did not match any valid pattern: %s, found %d instead of 3: %s", typeName, len(matches), matches))
		}
		pkgSection := fmt.Sprintf("%s__", matches[1])
		typeSection := strings.ReplaceAll(matches[2], ".", "__")
		fullTypeSection := typeSection + repeatedStr
		for _, exp := range fileExports {
			if exp == typeSection {
				return fullTypeSection
			}
		}
		return fmt.Sprintf("%s%s", pkgSection, fullTypeSection)
	default:
		panic(fmt.Errorf("unknown field type: %s", field))
	}
}

func getProtoJSONTypeName(field *descriptorpb.FieldDescriptorProto, nestedTypes []*descriptorpb.DescriptorProto) string {
	panic("not implemented")
}
