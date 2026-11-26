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
	CreateGroupChannel(ctx context.Context, in *userv1.CreateGroupChannelRequest) (*userv1.CreateGroupChannelResponse, error)
	UpdateGroupChannel(ctx context.Context, in *userv1.UpdateGroupChannelRequest) (*userv1.UpdateGroupChannelResponse, error)
	DeleteGroupChannel(ctx context.Context, in *userv1.DeleteGroupChannelRequest) (*userv1.DeleteGroupChannelResponse, error)
	ListChannelPermissions(ctx context.Context, in *userv1.ListChannelPermissionsRequest) (*userv1.ListChannelPermissionsResponse, error)
	UpsertChannelPermission(ctx context.Context, in *userv1.UpsertChannelPermissionRequest) (*userv1.UpsertChannelPermissionResponse, error)
	DeleteChannelPermission(ctx context.Context, in *userv1.DeleteChannelPermissionRequest) (*userv1.DeleteChannelPermissionResponse, error)
	SyncChannelPermissions(ctx context.Context, in *userv1.SyncChannelPermissionsRequest) (*userv1.SyncChannelPermissionsResponse, error)
	ReorderGroupChannels(ctx context.Context, in *userv1.ReorderGroupChannelsRequest) (*userv1.ReorderGroupChannelsResponse, error)
	AssertGroupChannelAccess(ctx context.Context, in *userv1.AssertGroupChannelAccessRequest) (*userv1.AssertGroupChannelAccessResponse, error)
	EnsureGroupDmChannel(ctx context.Context, in *userv1.EnsureGroupDmChannelRequest) (*userv1.EnsureGroupDmChannelResponse, error)
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

func (c *client) CreateGroupChannel(ctx context.Context, in *userv1.CreateGroupChannelRequest) (*userv1.CreateGroupChannelResponse, error) {
	return utils.CallUnaryResult(ctx, c.write, c.log, "channel.ChannelService/CreateGroupChannel", func(callCtx context.Context) (*userv1.CreateGroupChannelResponse, error) {
		return c.stub.CreateGroupChannel(callCtx, in)
	})
}

func (c *client) UpdateGroupChannel(ctx context.Context, in *userv1.UpdateGroupChannelRequest) (*userv1.UpdateGroupChannelResponse, error) {
	return utils.CallUnaryResult(ctx, c.write, c.log, "channel.ChannelService/UpdateGroupChannel", func(callCtx context.Context) (*userv1.UpdateGroupChannelResponse, error) {
		return c.stub.UpdateGroupChannel(callCtx, in)
	})
}

func (c *client) DeleteGroupChannel(ctx context.Context, in *userv1.DeleteGroupChannelRequest) (*userv1.DeleteGroupChannelResponse, error) {
	return utils.CallUnaryResult(ctx, c.write, c.log, "channel.ChannelService/DeleteGroupChannel", func(callCtx context.Context) (*userv1.DeleteGroupChannelResponse, error) {
		return c.stub.DeleteGroupChannel(callCtx, in)
	})
}

func (c *client) ListChannelPermissions(ctx context.Context, in *userv1.ListChannelPermissionsRequest) (*userv1.ListChannelPermissionsResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "channel.ChannelService/ListChannelPermissions", func(callCtx context.Context) (*userv1.ListChannelPermissionsResponse, error) {
		return c.stub.ListChannelPermissions(callCtx, in)
	})
}

func (c *client) UpsertChannelPermission(ctx context.Context, in *userv1.UpsertChannelPermissionRequest) (*userv1.UpsertChannelPermissionResponse, error) {
	return utils.CallUnaryResult(ctx, c.write, c.log, "channel.ChannelService/UpsertChannelPermission", func(callCtx context.Context) (*userv1.UpsertChannelPermissionResponse, error) {
		return c.stub.UpsertChannelPermission(callCtx, in)
	})
}

func (c *client) DeleteChannelPermission(ctx context.Context, in *userv1.DeleteChannelPermissionRequest) (*userv1.DeleteChannelPermissionResponse, error) {
	return utils.CallUnaryResult(ctx, c.write, c.log, "channel.ChannelService/DeleteChannelPermission", func(callCtx context.Context) (*userv1.DeleteChannelPermissionResponse, error) {
		return c.stub.DeleteChannelPermission(callCtx, in)
	})
}

func (c *client) SyncChannelPermissions(ctx context.Context, in *userv1.SyncChannelPermissionsRequest) (*userv1.SyncChannelPermissionsResponse, error) {
	return utils.CallUnaryResult(ctx, c.write, c.log, "channel.ChannelService/SyncChannelPermissions", func(callCtx context.Context) (*userv1.SyncChannelPermissionsResponse, error) {
		return c.stub.SyncChannelPermissions(callCtx, in)
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

func (c *client) EnsureGroupDmChannel(ctx context.Context, in *userv1.EnsureGroupDmChannelRequest) (*userv1.EnsureGroupDmChannelResponse, error) {
	return utils.CallUnaryResult(ctx, c.write, c.log, "channel.ChannelService/EnsureGroupDmChannel", func(callCtx context.Context) (*userv1.EnsureGroupDmChannelResponse, error) {
		return c.stub.EnsureGroupDmChannel(callCtx, in)
	})
}
