param(
    [string]$os=$(throw "os Parameter missing,such : -os linux") ,
    [string]$arch=$(throw "arch Parameter missing,such : -arch amd64") ,
    [string]$name=$(throw "arch Parameter missing,such : -name davinci")
)

go env -w GOOS=$os
go env -w GOARCH=$arch
go build -ldflags "-w -s" -o bin/$name davinci.go