proto2.pb.go: proto2.proto
	prutalgen --proto_path=. --go_out=. --go_opt=paths=source_relative ./proto2.proto

test: proto2.pb.go
	go build
