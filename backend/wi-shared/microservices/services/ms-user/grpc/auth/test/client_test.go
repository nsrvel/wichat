package auth_test

import (
	"context"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/auth"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	authv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/auth/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type setupStatusServer struct {
	authv1.UnimplementedAuthServiceServer
	calls atomic.Int32
}

func (s *setupStatusServer) SetupStatus(context.Context, *authv1.SetupStatusRequest) (*authv1.SetupStatusResponse, error) {
	if s.calls.Add(1) == 1 {
		return nil, status.Error(codes.Unavailable, "try again")
	}

	return &authv1.SetupStatusResponse{NeedsSetup: true}, nil
}

func TestSetupStatusRetriesIdempotent(t *testing.T) {
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	authv1.RegisterAuthServiceServer(srv, &setupStatusServer{})
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(func() { srv.Stop() })

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	settings := utils.Settings{
		AttemptTimeoutMs:        500,
		OverallTimeoutMs:        2000,
		MaxRetries:              2,
		RetryBackoffMs:          1,
		BreakerFailureThreshold: 0,
	}
	grpcClient := auth.New(conn, settings, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	out, err := grpcClient.SetupStatus(ctx, &authv1.SetupStatusRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if !out.GetNeedsSetup() {
		t.Fatalf("unexpected response: %+v", out)
	}
}
