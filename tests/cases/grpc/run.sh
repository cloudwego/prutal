#!/bin/bash

rm -f *.pb.go
rm -f go.mod
rm -f go.sum

prutalgen --go_out=. --go_opt=paths=source_relative echo.proto
protoc --go-grpc_out=. --go-grpc_opt=paths=source_relative echo.proto

#go mod init echo
#go mod edit -replace=github.com/cloudwego/prutal=../../../
#go mod edit -replace=github.com/cloudwego/prutal/pkg/grpccodec=../../../pkg/grpccodec
#go mod tidy

go test -v
