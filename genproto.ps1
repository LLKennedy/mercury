protoc --go_out=plugins=grpc:. *.proto
Set-Location ./internal/testservice
protoc --go_out=plugins=grpc:. *.proto
Set-Location ../..