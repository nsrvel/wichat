package grpc_test

import (
	"context"
	"net"
	"testing"

	grpcutil "github.com/wichat/wichat/backend/wi-shared/microservices/utils/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestDialRejectsEmptyTarget(t *testing.T) {
	t.Parallel()

	if _, err := grpcutil.Dial("", grpcutil.DefaultDialOptions); err == nil {
		t.Fatal("expected error")
	}
}

func TestDialBufconn(t *testing.T) {
	t.Parallel()

	lis := bufconn.Listen(1 << 20)
	srv := grpc.NewServer()
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(func() { srv.Stop() })

	dialer := func(context.Context, string) (net.Conn, error) { return lis.Dial() }
	conn, err := grpcutil.Dial(
		"passthrough:///bufnet",
		grpcutil.DefaultDialOptions,
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
}
