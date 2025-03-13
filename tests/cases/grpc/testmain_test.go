package echo

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "github.com/cloudwego/prutal/pkg/grpccodec"
)

type server struct {
	UnimplementedEchoServiceServer
}

func (s *server) Echo(ctx context.Context, req *EchoRequest) (*EchoResponse, error) {
	return &EchoResponse{Message: req.Message}, nil
}

func startServer(t *testing.T) net.Listener {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	RegisterEchoServiceServer(s, &server{})
	go s.Serve(ln)
	return ln
}

func TestGRPC(t *testing.T) {
	ln := startServer(t)
	defer ln.Close()

	conn, err := grpc.Dial(
		ln.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
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
}
