go env -w GOOS=windows
go env -w GOARCH=amd64
go build -ldflags "-w -s" -o bin/davinci.exe davinci.go