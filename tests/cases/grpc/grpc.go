package echo

import (
	context "context"
	"sync/atomic"

	"github.com/cloudwego/prutal/pkg/grpccodec"
	"google.golang.org/grpc/encoding"
)

type EchoServer struct {
	UnimplementedEchoServiceServer
}

var _ encoding.Codec = grpccodec.PrutalCodec{}

// CountingCodec verifies that gRPC actually routes message encoding through
// grpccodec.PrutalCodec instead of the default protobuf codec.
type CountingCodec struct {
	grpccodec.PrutalCodec
	marshalCount   atomic.Int32
	unmarshalCount atomic.Int32
}

func (c *CountingCodec) Marshal(v any) ([]byte, error) {
	c.marshalCount.Add(1)
	return c.PrutalCodec.Marshal(v)
}

func (c *CountingCodec) Unmarshal(data []byte, v any) error {
	c.unmarshalCount.Add(1)
	return c.PrutalCodec.Unmarshal(data, v)
}

func (s *EchoServer) Echo(ctx context.Context, req *EchoRequest) (*EchoResponse, error) {
	return &EchoResponse{Message: req.Message}, nil
}
