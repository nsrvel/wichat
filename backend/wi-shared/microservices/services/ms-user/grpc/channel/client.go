// Package channel is the gRPC client for ms-user channel-domain RPCs on AuthService.
package channel

import (
	"context"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

// Client calls channel-domain AuthService RPCs with shared timeouts, retries, and circuit breaking.
type Client interface {
	ListGroupChannels(ctx context.Context, in *userv1.ListGroupChannelsRequest) (*userv1.ListGroupChannelsResponse, error)
	AssertGroupChannelAccess(ctx context.Context, in *userv1.AssertGroupChannelAccessRequest) (*userv1.AssertGroupChannelAccessResponse, error)
}

type client struct {
	stub userv1.AuthServiceClient
	log  logger.Logger
	read utils.CallConfig
}

// New wraps an existing AuthService client stub.
func New(stub userv1.AuthServiceClient, settings utils.Settings, log logger.Logger) Client {
	breaker := utils.NewBreaker(settings)
	read := utils.CallConfigFromSettings(settings, utils.RetryIdempotent)
	read.Breaker = breaker

	return &client{
		stub: stub,
		log:  log,
		read: read,
	}
}

func (c *client) ListGroupChannels(ctx context.Context, in *userv1.ListGroupChannelsRequest) (*userv1.ListGroupChannelsResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "auth.AuthService/ListGroupChannels", func(callCtx context.Context) (*userv1.ListGroupChannelsResponse, error) {
		return c.stub.ListGroupChannels(callCtx, in)
	})
}

func (c *client) AssertGroupChannelAccess(ctx context.Context, in *userv1.AssertGroupChannelAccessRequest) (*userv1.AssertGroupChannelAccessResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "auth.AuthService/AssertGroupChannelAccess", func(callCtx context.Context) (*userv1.AssertGroupChannelAccessResponse, error) {
		return c.stub.AssertGroupChannelAccess(callCtx, in)
	})
}
