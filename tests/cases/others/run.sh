#!/bin/bash

set -e

# gen test
prutalgen --proto_path=. --go_out=. --go_opt=paths=source_relative ./others.proto
go build

# use protoc
rm others.pb.go
protoc --proto_path=. --go_out=. --go_opt=paths=source_relative ./others.proto

go test -v
