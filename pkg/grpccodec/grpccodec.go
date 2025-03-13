package grpccodec

import (
	"google.golang.org/grpc/encoding"
	_ "google.golang.org/grpc/encoding/proto" // force the pkg init before this package

	"github.com/cloudwego/prutal"
)

func init() {
	encoding.RegisterCodec(&protoCodec{})
}

// protoCodec implements google.golang.org/grpc/encoding.Codec
type protoCodec struct{}

var _ encoding.Codec = (*protoCodec)(nil)

func (p *protoCodec) Marshal(v any) ([]byte, error) {
	return prutal.MarshalAppend(nil, v)
}

func (p *protoCodec) Unmarshal(data []byte, v any) error {
	return prutal.Unmarshal(data, v)
}

func (p *protoCodec) Name() string {
	// same as https://pkg.go.dev/google.golang.org/grpc/encoding/proto
	// so that we can overwrites the default proto codec
	return "proto"
}
