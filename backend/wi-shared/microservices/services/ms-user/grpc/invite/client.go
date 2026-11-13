// Package invite is the gRPC client for ms-user invite domain RPCs on InviteService.
package invite

import (
	"context"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

// Client calls invite-domain InviteService RPCs with shared timeouts, retries, and circuit breaking.
type Client interface {
	GetGroupInviteLink(ctx context.Context, in *userv1.GetGroupInviteLinkRequest) (*userv1.GetGroupInviteLinkResponse, error)
	SendGroupInvites(ctx context.Context, in *userv1.SendGroupInvitesRequest) (*userv1.SendGroupInvitesResponse, error)
	PreviewGroupInvite(ctx context.Context, in *userv1.PreviewGroupInviteRequest) (*userv1.PreviewGroupInviteResponse, error)
	RedeemGroupInvite(ctx context.Context, in *userv1.RedeemGroupInviteRequest) (*userv1.RedeemGroupInviteResponse, error)
}

type client struct {
	stub userv1.InviteServiceClient
	log  logger.Logger
	none utils.CallConfig
	read utils.CallConfig
}

// New wraps an existing InviteService client stub.
func New(stub userv1.InviteServiceClient, settings utils.Settings, log logger.Logger) Client {
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

func (c *client) GetGroupInviteLink(ctx context.Context, in *userv1.GetGroupInviteLinkRequest) (*userv1.GetGroupInviteLinkResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "invite.InviteService/GetGroupInviteLink", func(callCtx context.Context) (*userv1.GetGroupInviteLinkResponse, error) {
		return c.stub.GetGroupInviteLink(callCtx, in)
	})
}

func (c *client) SendGroupInvites(ctx context.Context, in *userv1.SendGroupInvitesRequest) (*userv1.SendGroupInvitesResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "invite.InviteService/SendGroupInvites", func(callCtx context.Context) (*userv1.SendGroupInvitesResponse, error) {
		return c.stub.SendGroupInvites(callCtx, in)
	})
}

func (c *client) PreviewGroupInvite(ctx context.Context, in *userv1.PreviewGroupInviteRequest) (*userv1.PreviewGroupInviteResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "invite.InviteService/PreviewGroupInvite", func(callCtx context.Context) (*userv1.PreviewGroupInviteResponse, error) {
		return c.stub.PreviewGroupInvite(callCtx, in)
	})
}

func (c *client) RedeemGroupInvite(ctx context.Context, in *userv1.RedeemGroupInviteRequest) (*userv1.RedeemGroupInviteResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "invite.InviteService/RedeemGroupInvite", func(callCtx context.Context) (*userv1.RedeemGroupInviteResponse, error) {
		return c.stub.RedeemGroupInvite(callCtx, in)
	})
}
