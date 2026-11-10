// Package group is the gRPC client for ms-user group-domain RPCs on AuthService.
package group

import (
	"context"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

// Client calls group-domain AuthService RPCs with shared timeouts, retries, and circuit breaking.
type Client interface {
	ListMyGroups(ctx context.Context, in *userv1.ListMyGroupsRequest) (*userv1.ListMyGroupsResponse, error)
	ListJoinableGroups(ctx context.Context, in *userv1.ListJoinableGroupsRequest) (*userv1.ListJoinableGroupsResponse, error)
	CreateGroup(ctx context.Context, in *userv1.CreateGroupRequest) (*userv1.CreateGroupResponse, error)
	JoinGroup(ctx context.Context, in *userv1.JoinGroupRequest) (*userv1.JoinGroupResponse, error)
	LeaveGroup(ctx context.Context, in *userv1.LeaveGroupRequest) (*userv1.LeaveGroupResponse, error)
	DeleteGroup(ctx context.Context, in *userv1.DeleteGroupRequest) (*userv1.DeleteGroupResponse, error)
	UploadGroupIcon(ctx context.Context, in *userv1.UploadGroupIconRequest) (*userv1.UploadGroupIconResponse, error)
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

func (c *client) ListMyGroups(ctx context.Context, in *userv1.ListMyGroupsRequest) (*userv1.ListMyGroupsResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "auth.AuthService/ListMyGroups", func(callCtx context.Context) (*userv1.ListMyGroupsResponse, error) {
		return c.stub.ListMyGroups(callCtx, in)
	})
}

func (c *client) ListJoinableGroups(ctx context.Context, in *userv1.ListJoinableGroupsRequest) (*userv1.ListJoinableGroupsResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "auth.AuthService/ListJoinableGroups", func(callCtx context.Context) (*userv1.ListJoinableGroupsResponse, error) {
		return c.stub.ListJoinableGroups(callCtx, in)
	})
}

func (c *client) CreateGroup(ctx context.Context, in *userv1.CreateGroupRequest) (*userv1.CreateGroupResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/CreateGroup", func(callCtx context.Context) (*userv1.CreateGroupResponse, error) {
		return c.stub.CreateGroup(callCtx, in)
	})
}

func (c *client) JoinGroup(ctx context.Context, in *userv1.JoinGroupRequest) (*userv1.JoinGroupResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/JoinGroup", func(callCtx context.Context) (*userv1.JoinGroupResponse, error) {
		return c.stub.JoinGroup(callCtx, in)
	})
}

func (c *client) LeaveGroup(ctx context.Context, in *userv1.LeaveGroupRequest) (*userv1.LeaveGroupResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/LeaveGroup", func(callCtx context.Context) (*userv1.LeaveGroupResponse, error) {
		return c.stub.LeaveGroup(callCtx, in)
	})
}

func (c *client) DeleteGroup(ctx context.Context, in *userv1.DeleteGroupRequest) (*userv1.DeleteGroupResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/DeleteGroup", func(callCtx context.Context) (*userv1.DeleteGroupResponse, error) {
		return c.stub.DeleteGroup(callCtx, in)
	})
}

func (c *client) UploadGroupIcon(ctx context.Context, in *userv1.UploadGroupIconRequest) (*userv1.UploadGroupIconResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "auth.AuthService/UploadGroupIcon", func(callCtx context.Context) (*userv1.UploadGroupIconResponse, error) {
		return c.stub.UploadGroupIcon(callCtx, in)
	})
}
