#!/bin/bash
set -e
unset GOARCH # fix go install

cd ../../prutalgen/
go install
cd -

prutalgen --proto_path=. --go_out=../ ./testdata.proto
