$Directory = "."
$IncludeRule = "*.proto"
$ExcludeRUle = [Regex]'.*google.*'
$ProtoFiles = Get-ChildItem -path $Directory -Recurse -Include $IncludeRule | Where-Object FullName -NotMatch $ExcludeRUle
foreach ($file in $ProtoFiles) {
    protoc --proto_path="$($file.DirectoryName)" -I "./proto"  --go_out=plugins=grpc:"$($file.DirectoryName)" $file.FullName
}