#!/bin/bash

set -e

# gen test
prutalgen --proto_path=. --go_out=. --go_opt=paths=source_relative --gen_getter=true ./nested.proto
go build

# use protoc
rm nested.pb.go
protoc --proto_path=. --go_out=. --go_opt=paths=source_relative ./nested.proto

go test -v
