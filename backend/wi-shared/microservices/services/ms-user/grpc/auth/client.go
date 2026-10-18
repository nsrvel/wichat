// Package auth is the gRPC client for ms-user AuthService.
package auth

import (
	"context"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"google.golang.org/grpc"
)

// Client calls ms-user AuthService with shared timeouts, retries, and circuit breaking.
type Client interface {
	SetupStatus(ctx context.Context, in *userv1.SetupStatusRequest) (*userv1.SetupStatusResponse, error)
	Setup(ctx context.Context, in *userv1.SetupRequest) (*userv1.SetupResponse, error)
	Login(ctx context.Context, in *userv1.LoginRequest) (*userv1.LoginResponse, error)
	Refresh(ctx context.Context, in *userv1.RefreshRequest) (*userv1.RefreshResponse, error)
	Logout(ctx context.Context, in *userv1.LogoutRequest) (*userv1.LogoutResponse, error)
	Me(ctx context.Context, in *userv1.MeRequest) (*userv1.MeResponse, error)
	GoogleStart(ctx context.Context, in *userv1.GoogleStartRequest) (*userv1.GoogleStartResponse, error)
	GoogleFinish(ctx context.Context, in *userv1.GoogleFinishRequest) (*userv1.GoogleFinishResponse, error)
}

type client struct {
	stub userv1.AuthServiceClient
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
		stub: userv1.NewAuthServiceClient(conn),
		log:  log,
		none: none,
		read: read,
	}
}

func (c *client) SetupStatus(ctx context.Context, in *userv1.SetupStatusRequest) (*userv1.SetupStatusResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "auth.AuthService/SetupStatus", func(callCtx context.Context) (*userv1.SetupStatusResponse, error) {
		return c.stub.SetupStatus(callCtx, in)
	})
}

func (c *client) Setup(ctx context.Context, in *userv1.SetupRequest) (*userv1.SetupResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/Setup", func(callCtx context.Context) (*userv1.SetupResponse, error) {
		return c.stub.Setup(callCtx, in)
	})
}

func (c *client) Login(ctx context.Context, in *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/Login", func(callCtx context.Context) (*userv1.LoginResponse, error) {
		return c.stub.Login(callCtx, in)
	})
}

func (c *client) Refresh(ctx context.Context, in *userv1.RefreshRequest) (*userv1.RefreshResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/Refresh", func(callCtx context.Context) (*userv1.RefreshResponse, error) {
		return c.stub.Refresh(callCtx, in)
	})
}

func (c *client) Logout(ctx context.Context, in *userv1.LogoutRequest) (*userv1.LogoutResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/Logout", func(callCtx context.Context) (*userv1.LogoutResponse, error) {
		return c.stub.Logout(callCtx, in)
	})
}

func (c *client) Me(ctx context.Context, in *userv1.MeRequest) (*userv1.MeResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "auth.AuthService/Me", func(callCtx context.Context) (*userv1.MeResponse, error) {
		return c.stub.Me(callCtx, in)
	})
}

func (c *client) GoogleStart(ctx context.Context, in *userv1.GoogleStartRequest) (*userv1.GoogleStartResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/GoogleStart", func(callCtx context.Context) (*userv1.GoogleStartResponse, error) {
		return c.stub.GoogleStart(callCtx, in)
	})
}

func (c *client) GoogleFinish(ctx context.Context, in *userv1.GoogleFinishRequest) (*userv1.GoogleFinishResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/GoogleFinish", func(callCtx context.Context) (*userv1.GoogleFinishResponse, error) {
		return c.stub.GoogleFinish(callCtx, in)
	})
}
