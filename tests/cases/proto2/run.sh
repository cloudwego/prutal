#!/bin/bash
set -e
prutalgen --proto_path=. --go_out=. --go_opt=paths=source_relative ./proto2.proto
go build
