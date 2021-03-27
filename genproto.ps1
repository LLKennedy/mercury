go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.25.0
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.0.0
go get google.golang.org/grpc@v1.34.0
go install ./cmd/protoc-gen-mercury
$Directory = "./httpapi"
$IncludeRule = "*.proto"
$ExcludeRUle = [Regex]'.*google.*|.*audit/.*'
$PBPath = "./httpapi"
$ProtoFiles = Get-ChildItem -path $Directory -Recurse -Include $IncludeRule | Where-Object FullName -NotMatch $ExcludeRUle
foreach ($file in $ProtoFiles) {
	protoc --proto_path="$($file.DirectoryName)" --go_out=paths=source_relative:$PBPath --go-grpc_out=paths=source_relative:$PBPath $file.FullName
}
$Directory = "./internal/testservice/service"
$IncludeRule = "*.proto"
$ExcludeRUle = [Regex]'.*google.*|.*audit/.*|.*node_modules.*'
$PBPath = "./internal/testservice/service"
$ProtoFiles = Get-ChildItem -path $Directory -Recurse -Include $IncludeRule | Where-Object FullName -NotMatch $ExcludeRUle
foreach ($file in $ProtoFiles) {
	protoc --proto_path="$($file.DirectoryName)" --go_out=paths=source_relative:$PBPath --go-grpc_out=paths=source_relative:$PBPath --tsjson_out=$PBPath --mercury_out=$PBPath $file.FullName 
}
go build $PBPath
go mod tidy