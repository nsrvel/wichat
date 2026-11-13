// Package user is the gRPC client for ms-user user-domain RPCs on UserService.
package user

import (
	"context"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

// Client calls user-domain UserService RPCs with shared timeouts, retries, and circuit breaking.
type Client interface {
	SetupStatus(ctx context.Context, in *userv1.SetupStatusRequest) (*userv1.SetupStatusResponse, error)
	CompleteInstanceSetup(ctx context.Context, in *userv1.CompleteInstanceSetupRequest) (*userv1.CompleteInstanceSetupResponse, error)
	SearchUsers(ctx context.Context, in *userv1.SearchUsersRequest) (*userv1.SearchUsersResponse, error)
	CheckUsername(ctx context.Context, in *userv1.CheckUsernameRequest) (*userv1.CheckUsernameResponse, error)
}

type client struct {
	stub userv1.UserServiceClient
	log  logger.Logger
	none utils.CallConfig
	read utils.CallConfig
}

// New wraps an existing UserService client stub.
func New(stub userv1.UserServiceClient, settings utils.Settings, log logger.Logger) Client {
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
	return utils.CallUnaryResult(ctx, c.read, c.log, "user.UserService/SetupStatus", func(callCtx context.Context) (*userv1.SetupStatusResponse, error) {
		return c.stub.SetupStatus(callCtx, in)
	})
}

func (c *client) CompleteInstanceSetup(ctx context.Context, in *userv1.CompleteInstanceSetupRequest) (*userv1.CompleteInstanceSetupResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "user.UserService/CompleteInstanceSetup", func(callCtx context.Context) (*userv1.CompleteInstanceSetupResponse, error) {
		return c.stub.CompleteInstanceSetup(callCtx, in)
	})
}

func (c *client) SearchUsers(ctx context.Context, in *userv1.SearchUsersRequest) (*userv1.SearchUsersResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "user.UserService/SearchUsers", func(callCtx context.Context) (*userv1.SearchUsersResponse, error) {
		return c.stub.SearchUsers(callCtx, in)
	})
}

func (c *client) CheckUsername(ctx context.Context, in *userv1.CheckUsernameRequest) (*userv1.CheckUsernameResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "user.UserService/CheckUsername", func(callCtx context.Context) (*userv1.CheckUsernameResponse, error) {
		return c.stub.CheckUsername(callCtx, in)
	})
}
