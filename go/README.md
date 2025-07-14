How to run

`go run .`

$env:GOOS="windows"; $env:GOARCH="amd64"; go run . 

$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o remote