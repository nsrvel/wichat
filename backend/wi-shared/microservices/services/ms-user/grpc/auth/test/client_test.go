package auth_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/auth"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type authMFAStub struct {
	userv1.UnimplementedAuthServiceServer
}

func (s *authMFAStub) VerifyLoginMFA(_ context.Context, _ *userv1.VerifyLoginMFARequest) (*userv1.VerifyLoginMFAResponse, error) {
	return &userv1.VerifyLoginMFAResponse{
		Session: &userv1.Session{AccessToken: "acc-1", ExpiresIn: 900, User: &userv1.User{Id: "u-1"}},
	}, nil
}

func (s *authMFAStub) BeginLoginWebAuthn(_ context.Context, _ *userv1.BeginLoginWebAuthnRequest) (*userv1.BeginLoginWebAuthnResponse, error) {
	return &userv1.BeginLoginWebAuthnResponse{WebauthnOptionsJson: `{"publicKey":{}}`}, nil
}

func TestAuthMFACalls(t *testing.T) {
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	userv1.RegisterAuthServiceServer(srv, &authMFAStub{})
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
		RetryBackoffMs:          1,
		BreakerFailureThreshold: 0,
	}
	c := auth.New(userv1.NewAuthServiceClient(conn), settings, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	beginOut, err := c.BeginLoginWebAuthn(ctx, &userv1.BeginLoginWebAuthnRequest{ChallengeId: "ch-1"})
	if err != nil || beginOut.GetWebauthnOptionsJson() == "" {
		t.Fatalf("unexpected begin login webauthn: %+v %v", beginOut, err)
	}

	verOut, err := c.VerifyLoginMFA(ctx, &userv1.VerifyLoginMFARequest{ChallengeId: "ch-1", Method: "totp", Code: "123456"})
	if err != nil || verOut.GetSession().GetAccessToken() != "acc-1" {
		t.Fatalf("unexpected verify login mfa: %+v %v", verOut, err)
	}
}
