go env -w GOOS=linux
go env -w GOARCH=amd64
go build -ldflags "-w -s" -o bin/davinci davinci.go