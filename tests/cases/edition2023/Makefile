test_messages_edition2023.pb.go: test_messages_edition2023.proto
	prutalgen --proto_path=. --go_out=. \
		--go_opt=paths=source_relative --go_opt=Mtest_messages_edition2023.proto=./edition2023 \
		--gen_getter=true \
		./test_messages_edition2023.proto

test: test_messages_edition2023.pb.go
	go build
