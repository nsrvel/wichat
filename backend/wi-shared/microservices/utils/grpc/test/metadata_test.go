package grpc_test

import (
	"context"
	"testing"

	"github.com/wichat/wichat/backend/wi-shared/microservices/utils/correlation"
	grpcutil "github.com/wichat/wichat/backend/wi-shared/microservices/utils/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestUnaryClientRequestID(t *testing.T) {
	t.Parallel()

	inter := grpcutil.UnaryClientRequestID()
	ctx := correlation.WithRequestID(context.Background(), "rid-client")
	var got string
	err := inter(ctx, "/m", nil, nil, nil, func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			return nil
		}
		vals := md.Get("x-request-id")
		if len(vals) > 0 {
			got = vals[0]
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != "rid-client" {
		t.Fatalf("got %q", got)
	}
}

func TestUnaryServerRequestID(t *testing.T) {
	t.Parallel()

	inter := grpcutil.UnaryServerRequestID()
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-request-id", "rid-in"))
	var seen string
	_, err := inter(ctx, struct{}{}, &grpc.UnaryServerInfo{FullMethod: "/test/M"}, func(ctx context.Context, req any) (any, error) {
		seen = correlation.RequestIDFromContext(ctx)

		return struct{}{}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if seen != "rid-in" {
		t.Fatalf("got %q", seen)
	}
}
