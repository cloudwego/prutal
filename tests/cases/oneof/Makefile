oneof.pb.go: oneof.proto
	prutalgen --proto_path=. --go_out=. --go_opt=paths=source_relative ./oneof.proto

test: oneof.pb.go
	go build
