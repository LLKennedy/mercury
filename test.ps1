gofmt -s -w .
go test ./... -race -cover -coverprofile="coverage.out"; 
if ($LastExitCode -eq 0) {
	go tool cover -html="coverage.out";
	go tool cover -func="coverage.out";
}
Remove-Item coverage.out;