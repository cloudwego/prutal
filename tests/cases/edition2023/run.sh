#!/bin/bash

set -e

# gen test
prutalgen --proto_path=. --go_out=. \
		--go_opt=paths=source_relative --go_opt=Mtest_messages_edition2023.proto=./edition2023 \
		--gen_getter=true \
		./test_messages_edition2023.proto

go build

rm test_messages_edition2023.pb.go

protoc --proto_path=. --go_out=. \
    --go_opt=paths=source_relative --go_opt=Mtest_messages_edition2023.proto=./edition2023 \
    ./test_messages_edition2023.proto

go test -v -count=10
