$Directory = "."
$ProtoFiles = Get-ChildItem -path $Directory -Recurse -Include *.proto 
foreach ($file in $ProtoFiles) {
    protoc --proto_path="$($file.DirectoryName)" --go_out=plugins=grpc:"$($file.DirectoryName)" $file.FullName
}