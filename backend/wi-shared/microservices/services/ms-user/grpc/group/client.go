// Package group is the gRPC client for ms-user group-domain RPCs on GroupService.
package group

import (
	"context"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

// Client calls group-domain GroupService RPCs with shared timeouts, retries, and circuit breaking.
type Client interface {
	ListMyGroups(ctx context.Context, in *userv1.ListMyGroupsRequest) (*userv1.ListMyGroupsResponse, error)
	ListJoinableGroups(ctx context.Context, in *userv1.ListJoinableGroupsRequest) (*userv1.ListJoinableGroupsResponse, error)
	CreateGroup(ctx context.Context, in *userv1.CreateGroupRequest) (*userv1.CreateGroupResponse, error)
	ReorderMyGroups(ctx context.Context, in *userv1.ReorderMyGroupsRequest) (*userv1.ReorderMyGroupsResponse, error)
	DeleteGroup(ctx context.Context, in *userv1.DeleteGroupRequest) (*userv1.DeleteGroupResponse, error)
	UploadGroupIcon(ctx context.Context, in *userv1.UploadGroupIconRequest) (*userv1.UploadGroupIconResponse, error)
	UpdateGroupProfile(ctx context.Context, in *userv1.UpdateGroupProfileRequest) (*userv1.UpdateGroupProfileResponse, error)
	UploadGroupBanner(ctx context.Context, in *userv1.UploadGroupBannerRequest) (*userv1.UploadGroupBannerResponse, error)
}

type client struct {
	stub userv1.GroupServiceClient
	log  logger.Logger
	none utils.CallConfig
	read utils.CallConfig
}

// New wraps an existing GroupService client stub.
func New(stub userv1.GroupServiceClient, settings utils.Settings, log logger.Logger) Client {
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

func (c *client) ListMyGroups(ctx context.Context, in *userv1.ListMyGroupsRequest) (*userv1.ListMyGroupsResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "group.GroupService/ListMyGroups", func(callCtx context.Context) (*userv1.ListMyGroupsResponse, error) {
		return c.stub.ListMyGroups(callCtx, in)
	})
}

func (c *client) ListJoinableGroups(ctx context.Context, in *userv1.ListJoinableGroupsRequest) (*userv1.ListJoinableGroupsResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "group.GroupService/ListJoinableGroups", func(callCtx context.Context) (*userv1.ListJoinableGroupsResponse, error) {
		return c.stub.ListJoinableGroups(callCtx, in)
	})
}

func (c *client) CreateGroup(ctx context.Context, in *userv1.CreateGroupRequest) (*userv1.CreateGroupResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "group.GroupService/CreateGroup", func(callCtx context.Context) (*userv1.CreateGroupResponse, error) {
		return c.stub.CreateGroup(callCtx, in)
	})
}

func (c *client) ReorderMyGroups(ctx context.Context, in *userv1.ReorderMyGroupsRequest) (*userv1.ReorderMyGroupsResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "group.GroupService/ReorderMyGroups", func(callCtx context.Context) (*userv1.ReorderMyGroupsResponse, error) {
		return c.stub.ReorderMyGroups(callCtx, in)
	})
}

func (c *client) DeleteGroup(ctx context.Context, in *userv1.DeleteGroupRequest) (*userv1.DeleteGroupResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "group.GroupService/DeleteGroup", func(callCtx context.Context) (*userv1.DeleteGroupResponse, error) {
		return c.stub.DeleteGroup(callCtx, in)
	})
}

func (c *client) UploadGroupIcon(ctx context.Context, in *userv1.UploadGroupIconRequest) (*userv1.UploadGroupIconResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "group.GroupService/UploadGroupIcon", func(callCtx context.Context) (*userv1.UploadGroupIconResponse, error) {
		return c.stub.UploadGroupIcon(callCtx, in)
	})
}

func (c *client) UpdateGroupProfile(ctx context.Context, in *userv1.UpdateGroupProfileRequest) (*userv1.UpdateGroupProfileResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "group.GroupService/UpdateGroupProfile", func(callCtx context.Context) (*userv1.UpdateGroupProfileResponse, error) {
		return c.stub.UpdateGroupProfile(callCtx, in)
	})
}

func (c *client) UploadGroupBanner(ctx context.Context, in *userv1.UploadGroupBannerRequest) (*userv1.UploadGroupBannerResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "group.GroupService/UploadGroupBanner", func(callCtx context.Context) (*userv1.UploadGroupBannerResponse, error) {
		return c.stub.UploadGroupBanner(callCtx, in)
	})
}
