# grpccodec

`grpccodec` provides a gRPC-compatible protobuf codec backed by `prutal`.

The package itself does not depend on `google.golang.org/grpc`. Its exported [`PrutalCodec`](./grpccodec.go) matches the v1 gRPC codec interface shape (`google.golang.org/grpc/encoding#Codec`), so applications can pass it directly to gRPC codec options.

## Usage with gRPC

```go
import (
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/encoding"

    "github.com/cloudwego/prutal/pkg/grpccodec"
)

var _ encoding.Codec = grpccodec.PrutalCodec{}

server := grpc.NewServer(grpc.ForceServerCodec(grpccodec.PrutalCodec{}))

conn, err := grpc.Dial(
    target,
    []grpc.DialOption{
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithDefaultCallOptions(grpc.ForceCodec(grpccodec.PrutalCodec{})),
    }...,
)
```

## Usage with `prutalgen`

`prutalgen` generates protobuf-compatible gRPC stubs. The codec integration is unchanged: use `grpccodec.PrutalCodec` with `grpc.ForceServerCodec(...)` on the server and `grpc.ForceCodec(...)` on the client.

For a complete working example, see [`tests/cases/grpc`](../../tests/cases/grpc).
