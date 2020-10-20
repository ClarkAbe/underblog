go env -w GO111MODULE=auto;go env -w GOPROXY=https://goproxy.io,direct
go build -o ./demo/underblog main.go && cd ./demo && ./underblog