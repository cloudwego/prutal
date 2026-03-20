package echo

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding"
)

func startServer(t *testing.T, codec encoding.Codec) net.Listener {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(grpc.ForceServerCodec(codec))
	RegisterEchoServiceServer(s, &EchoServer{})
	go s.Serve(ln)
	return ln
}

func TestGRPC(t *testing.T) {
	serverCodec := &CountingCodec{}
	clientCodec := &CountingCodec{}

	ln := startServer(t, serverCodec)
	defer ln.Close()

	conn, err := grpc.Dial(
		ln.Addr().String(),
		[]grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultCallOptions(grpc.ForceCodec(clientCodec)),
		}...,
	)
	if err != nil {
		t.Fatal("grpc.Dial err", err)
	}
	defer conn.Close()
	c := NewEchoServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	msg := "hello"
	resp, err := c.Echo(ctx, &EchoRequest{Message: msg})
	if err != nil {
		t.Fatal("Echo err", err)
	}
	if resp.Message != msg {
		t.Fatal("got", resp.Message)
	}
	if clientCodec.marshalCount.Load() == 0 {
		t.Fatal("client codec Marshal was not used")
	}
	if clientCodec.unmarshalCount.Load() == 0 {
		t.Fatal("client codec Unmarshal was not used")
	}
	if serverCodec.marshalCount.Load() == 0 {
		t.Fatal("server codec Marshal was not used")
	}
	if serverCodec.unmarshalCount.Load() == 0 {
		t.Fatal("server codec Unmarshal was not used")
	}
}
