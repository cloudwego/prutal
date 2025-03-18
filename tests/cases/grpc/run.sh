#!/bin/bash

set -e

rm -f *.pb.go

prutalgen --go_out=. --go_opt=paths=source_relative echo.proto
protoc --go-grpc_out=. --go-grpc_opt=paths=source_relative echo.proto

go test -v
