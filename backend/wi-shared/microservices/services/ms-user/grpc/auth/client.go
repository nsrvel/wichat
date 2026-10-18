// Package auth is the gRPC client for ms-user AuthService.
package auth

import (
	"context"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	authv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/auth/v1"
	"google.golang.org/grpc"
)

// Client calls ms-user AuthService with shared timeouts, retries, and circuit breaking.
type Client interface {
	SetupStatus(ctx context.Context, in *authv1.SetupStatusRequest) (*authv1.SetupStatusResponse, error)
	Setup(ctx context.Context, in *authv1.SetupRequest) (*authv1.SetupResponse, error)
	Login(ctx context.Context, in *authv1.LoginRequest) (*authv1.LoginResponse, error)
	Refresh(ctx context.Context, in *authv1.RefreshRequest) (*authv1.RefreshResponse, error)
	Logout(ctx context.Context, in *authv1.LogoutRequest) (*authv1.LogoutResponse, error)
	Me(ctx context.Context, in *authv1.MeRequest) (*authv1.MeResponse, error)
	GoogleStart(ctx context.Context, in *authv1.GoogleStartRequest) (*authv1.GoogleStartResponse, error)
	GoogleFinish(ctx context.Context, in *authv1.GoogleFinishRequest) (*authv1.GoogleFinishResponse, error)
}

type client struct {
	stub authv1.AuthServiceClient
	log  logger.Logger
	none utils.CallConfig
	read utils.CallConfig
}

// New wraps an existing gRPC connection (tests or custom dial). Caller owns conn lifetime.
func New(conn *grpc.ClientConn, settings utils.Settings, log logger.Logger) Client {
	breaker := utils.NewBreaker(settings)
	none := utils.CallConfigFromSettings(settings, utils.RetryNone)
	none.Breaker = breaker
	read := utils.CallConfigFromSettings(settings, utils.RetryIdempotent)
	read.Breaker = breaker

	return &client{
		stub: authv1.NewAuthServiceClient(conn),
		log:  log,
		none: none,
		read: read,
	}
}

func (c *client) SetupStatus(ctx context.Context, in *authv1.SetupStatusRequest) (*authv1.SetupStatusResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "auth.AuthService/SetupStatus", func(callCtx context.Context) (*authv1.SetupStatusResponse, error) {
		return c.stub.SetupStatus(callCtx, in)
	})
}

func (c *client) Setup(ctx context.Context, in *authv1.SetupRequest) (*authv1.SetupResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/Setup", func(callCtx context.Context) (*authv1.SetupResponse, error) {
		return c.stub.Setup(callCtx, in)
	})
}

func (c *client) Login(ctx context.Context, in *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/Login", func(callCtx context.Context) (*authv1.LoginResponse, error) {
		return c.stub.Login(callCtx, in)
	})
}

func (c *client) Refresh(ctx context.Context, in *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/Refresh", func(callCtx context.Context) (*authv1.RefreshResponse, error) {
		return c.stub.Refresh(callCtx, in)
	})
}

func (c *client) Logout(ctx context.Context, in *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/Logout", func(callCtx context.Context) (*authv1.LogoutResponse, error) {
		return c.stub.Logout(callCtx, in)
	})
}

func (c *client) Me(ctx context.Context, in *authv1.MeRequest) (*authv1.MeResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "auth.AuthService/Me", func(callCtx context.Context) (*authv1.MeResponse, error) {
		return c.stub.Me(callCtx, in)
	})
}

func (c *client) GoogleStart(ctx context.Context, in *authv1.GoogleStartRequest) (*authv1.GoogleStartResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/GoogleStart", func(callCtx context.Context) (*authv1.GoogleStartResponse, error) {
		return c.stub.GoogleStart(callCtx, in)
	})
}

func (c *client) GoogleFinish(ctx context.Context, in *authv1.GoogleFinishRequest) (*authv1.GoogleFinishResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/GoogleFinish", func(callCtx context.Context) (*authv1.GoogleFinishResponse, error) {
		return c.stub.GoogleFinish(callCtx, in)
	})
}
