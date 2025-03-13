#!/bin/bash
set -e
prutalgen --proto_path=. --go_out=. --go_opt=paths=source_relative ./oneof.proto
go build
