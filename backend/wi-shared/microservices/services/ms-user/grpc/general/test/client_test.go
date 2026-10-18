package general_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/general"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type healthStub struct {
	userv1.UnimplementedGeneralServiceServer
}

func (s *healthStub) Health(context.Context, *userv1.HealthRequest) (*userv1.HealthResponse, error) {
	return &userv1.HealthResponse{
		Status:  userv1.ServingStatus_SERVING_STATUS_SERVING,
		Service: "ms-user",
	}, nil
}

func TestHealth(t *testing.T) {
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	userv1.RegisterGeneralServiceServer(srv, &healthStub{})
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
		MaxRetries:              0,
		BreakerFailureThreshold: 0,
	}
	grpcClient := general.New(conn, settings, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	out, err := grpcClient.Health(ctx, &userv1.HealthRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if out.GetStatus() != userv1.ServingStatus_SERVING_STATUS_SERVING || out.GetService() != "ms-user" {
		t.Fatalf("unexpected response: %+v", out)
	}
}
