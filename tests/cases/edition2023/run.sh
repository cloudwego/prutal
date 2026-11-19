#!/bin/bash

set -e

rm -rf prutalout *.pb.go
mkdir -p prutalout
prutalgen --proto_path=. --go_out=prutalout \
    --go_opt=paths=source_relative \
    --go_opt=Mtest_messages_edition2023.proto=./prutalout \
    --go_opt=Mfeatures_interop.proto=./prutalout \
    --gen_getter=true \
    ./test_messages_edition2023.proto ./features_interop.proto

protoc --proto_path=. --go_out=. \
    --go_opt=paths=source_relative \
    --go_opt=Mtest_messages_edition2023.proto=./edition2023 \
    --go_opt=Mfeatures_interop.proto=./edition2023 \
    ./test_messages_edition2023.proto ./features_interop.proto

go test -v -count=10
