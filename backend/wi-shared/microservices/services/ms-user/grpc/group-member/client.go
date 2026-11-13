// Package groupmember is the gRPC client for ms-user group-member domain RPCs on GroupMemberService.
package groupmember

import (
	"context"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

// Client calls group-member domain GroupMemberService RPCs with shared timeouts, retries, and circuit breaking.
type Client interface {
	JoinGroup(ctx context.Context, in *userv1.JoinGroupRequest) (*userv1.JoinGroupResponse, error)
	LeaveGroup(ctx context.Context, in *userv1.LeaveGroupRequest) (*userv1.LeaveGroupResponse, error)
	ListGroupMembers(ctx context.Context, in *userv1.ListGroupMembersRequest) (*userv1.ListGroupMembersResponse, error)
	KickGroupMember(ctx context.Context, in *userv1.KickGroupMemberRequest) (*userv1.KickGroupMemberResponse, error)
	BanGroupMembers(ctx context.Context, in *userv1.BanGroupMembersRequest) (*userv1.BanGroupMembersResponse, error)
}

type client struct {
	stub userv1.GroupMemberServiceClient
	log  logger.Logger
	none utils.CallConfig
	read utils.CallConfig
}

// New wraps an existing GroupMemberService client stub.
func New(stub userv1.GroupMemberServiceClient, settings utils.Settings, log logger.Logger) Client {
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

func (c *client) JoinGroup(ctx context.Context, in *userv1.JoinGroupRequest) (*userv1.JoinGroupResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "groupmember.GroupMemberService/JoinGroup", func(callCtx context.Context) (*userv1.JoinGroupResponse, error) {
		return c.stub.JoinGroup(callCtx, in)
	})
}

func (c *client) LeaveGroup(ctx context.Context, in *userv1.LeaveGroupRequest) (*userv1.LeaveGroupResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "groupmember.GroupMemberService/LeaveGroup", func(callCtx context.Context) (*userv1.LeaveGroupResponse, error) {
		return c.stub.LeaveGroup(callCtx, in)
	})
}

func (c *client) ListGroupMembers(ctx context.Context, in *userv1.ListGroupMembersRequest) (*userv1.ListGroupMembersResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "groupmember.GroupMemberService/ListGroupMembers", func(callCtx context.Context) (*userv1.ListGroupMembersResponse, error) {
		return c.stub.ListGroupMembers(callCtx, in)
	})
}

func (c *client) KickGroupMember(ctx context.Context, in *userv1.KickGroupMemberRequest) (*userv1.KickGroupMemberResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "groupmember.GroupMemberService/KickGroupMember", func(callCtx context.Context) (*userv1.KickGroupMemberResponse, error) {
		return c.stub.KickGroupMember(callCtx, in)
	})
}

func (c *client) BanGroupMembers(ctx context.Context, in *userv1.BanGroupMembersRequest) (*userv1.BanGroupMembersResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "groupmember.GroupMemberService/BanGroupMembers", func(callCtx context.Context) (*userv1.BanGroupMembersResponse, error) {
		return c.stub.BanGroupMembers(callCtx, in)
	})
}
