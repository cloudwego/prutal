#!/bin/bash

set -e

protoc --proto_path=. --go_out=.  --go_opt=paths=source_relative ./benchmark.proto
GOMAXPROCS=1 go test -v -bench=. -benchmem
