// Package user is the gRPC client for ms-user user-domain RPCs on AuthService.
package user

import (
	"context"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

// Client calls user-domain AuthService RPCs with shared timeouts, retries, and circuit breaking.
type Client interface {
	SetupStatus(ctx context.Context, in *userv1.SetupStatusRequest) (*userv1.SetupStatusResponse, error)
	CompleteInstanceSetup(ctx context.Context, in *userv1.CompleteInstanceSetupRequest) (*userv1.CompleteInstanceSetupResponse, error)
	RegisterGuest(ctx context.Context, in *userv1.RegisterGuestRequest) (*userv1.RegisterGuestResponse, error)
}

type client struct {
	stub userv1.AuthServiceClient
	log  logger.Logger
	none utils.CallConfig
	read utils.CallConfig
}

// New wraps an existing AuthService client stub.
func New(stub userv1.AuthServiceClient, settings utils.Settings, log logger.Logger) Client {
	breaker := utils.NewBreaker(settings)
	none := utils.CallConfigFromSettings(settings, utils.RetryNone)
	none.Breaker = breaker
	read := utils.CallConfigFromSettings(settings, utils.RetryIdempotent)
	read.Breaker = breaker

	return &client{
		stub: stub,
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

func (c *client) CompleteInstanceSetup(ctx context.Context, in *userv1.CompleteInstanceSetupRequest) (*userv1.CompleteInstanceSetupResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/CompleteInstanceSetup", func(callCtx context.Context) (*userv1.CompleteInstanceSetupResponse, error) {
		return c.stub.CompleteInstanceSetup(callCtx, in)
	})
}

func (c *client) RegisterGuest(ctx context.Context, in *userv1.RegisterGuestRequest) (*userv1.RegisterGuestResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/RegisterGuest", func(callCtx context.Context) (*userv1.RegisterGuestResponse, error) {
		return c.stub.RegisterGuest(callCtx, in)
	})
}
