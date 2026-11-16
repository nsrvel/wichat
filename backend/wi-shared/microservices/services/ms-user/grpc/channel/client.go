// Package channel is the gRPC client for ms-user channel-domain RPCs on ChannelService.
package channel

import (
	"context"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

// Client calls channel-domain ChannelService RPCs with shared timeouts, retries, and circuit breaking.
type Client interface {
	ListGroupChannels(ctx context.Context, in *userv1.ListGroupChannelsRequest) (*userv1.ListGroupChannelsResponse, error)
	CreateChannelCategory(ctx context.Context, in *userv1.CreateChannelCategoryRequest) (*userv1.CreateChannelCategoryResponse, error)
	CreateGroupChannel(ctx context.Context, in *userv1.CreateGroupChannelRequest) (*userv1.CreateGroupChannelResponse, error)
	ReorderChannelCategories(ctx context.Context, in *userv1.ReorderChannelCategoriesRequest) (*userv1.ReorderChannelCategoriesResponse, error)
	ReorderGroupChannels(ctx context.Context, in *userv1.ReorderGroupChannelsRequest) (*userv1.ReorderGroupChannelsResponse, error)
	AssertGroupChannelAccess(ctx context.Context, in *userv1.AssertGroupChannelAccessRequest) (*userv1.AssertGroupChannelAccessResponse, error)
}

type client struct {
	stub  userv1.ChannelServiceClient
	log   logger.Logger
	read  utils.CallConfig
	write utils.CallConfig
}

// New wraps an existing ChannelService client stub.
func New(stub userv1.ChannelServiceClient, settings utils.Settings, log logger.Logger) Client {
	breaker := utils.NewBreaker(settings)
	read := utils.CallConfigFromSettings(settings, utils.RetryIdempotent)
	read.Breaker = breaker
	write := utils.CallConfigFromSettings(settings, utils.RetryNone)
	write.Breaker = breaker

	return &client{
		stub:  stub,
		log:   log,
		read:  read,
		write: write,
	}
}

func (c *client) ListGroupChannels(ctx context.Context, in *userv1.ListGroupChannelsRequest) (*userv1.ListGroupChannelsResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "channel.ChannelService/ListGroupChannels", func(callCtx context.Context) (*userv1.ListGroupChannelsResponse, error) {
		return c.stub.ListGroupChannels(callCtx, in)
	})
}

func (c *client) CreateChannelCategory(ctx context.Context, in *userv1.CreateChannelCategoryRequest) (*userv1.CreateChannelCategoryResponse, error) {
	return utils.CallUnaryResult(ctx, c.write, c.log, "channel.ChannelService/CreateChannelCategory", func(callCtx context.Context) (*userv1.CreateChannelCategoryResponse, error) {
		return c.stub.CreateChannelCategory(callCtx, in)
	})
}

func (c *client) CreateGroupChannel(ctx context.Context, in *userv1.CreateGroupChannelRequest) (*userv1.CreateGroupChannelResponse, error) {
	return utils.CallUnaryResult(ctx, c.write, c.log, "channel.ChannelService/CreateGroupChannel", func(callCtx context.Context) (*userv1.CreateGroupChannelResponse, error) {
		return c.stub.CreateGroupChannel(callCtx, in)
	})
}

func (c *client) ReorderChannelCategories(ctx context.Context, in *userv1.ReorderChannelCategoriesRequest) (*userv1.ReorderChannelCategoriesResponse, error) {
	return utils.CallUnaryResult(ctx, c.write, c.log, "channel.ChannelService/ReorderChannelCategories", func(callCtx context.Context) (*userv1.ReorderChannelCategoriesResponse, error) {
		return c.stub.ReorderChannelCategories(callCtx, in)
	})
}

func (c *client) ReorderGroupChannels(ctx context.Context, in *userv1.ReorderGroupChannelsRequest) (*userv1.ReorderGroupChannelsResponse, error) {
	return utils.CallUnaryResult(ctx, c.write, c.log, "channel.ChannelService/ReorderGroupChannels", func(callCtx context.Context) (*userv1.ReorderGroupChannelsResponse, error) {
		return c.stub.ReorderGroupChannels(callCtx, in)
	})
}

func (c *client) AssertGroupChannelAccess(ctx context.Context, in *userv1.AssertGroupChannelAccessRequest) (*userv1.AssertGroupChannelAccessResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "channel.ChannelService/AssertGroupChannelAccess", func(callCtx context.Context) (*userv1.AssertGroupChannelAccessResponse, error) {
		return c.stub.AssertGroupChannelAccess(callCtx, in)
	})
}
