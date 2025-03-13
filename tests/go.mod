module github.com/cloudwego/prutal/tests

go 1.23.7

replace github.com/cloudwego/prutal => ../

replace github.com/cloudwego/prutal/pkg/grpccodec => ../pkg/grpccodec

require (
	github.com/cloudwego/prutal/pkg/grpccodec v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.71.0
)

require (
	github.com/cloudwego/prutal v0.0.0-20250312062053-d17030f08590 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250115164207-1a7da9e5054f // indirect
	google.golang.org/protobuf v1.36.4 // indirect
)
