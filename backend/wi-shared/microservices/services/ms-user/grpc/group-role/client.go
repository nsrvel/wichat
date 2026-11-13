// Package grouprole is the gRPC client for ms-user group-role domain RPCs on GroupRoleService.
package grouprole

import (
	"context"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

// Client calls group-role domain GroupRoleService RPCs with shared timeouts, retries, and circuit breaking.
type Client interface {
	ListGroupRoles(ctx context.Context, in *userv1.ListGroupRolesRequest) (*userv1.ListGroupRolesResponse, error)
	UpdateGroupRole(ctx context.Context, in *userv1.UpdateGroupRoleRequest) (*userv1.UpdateGroupRoleResponse, error)
	AddGroupRoleMembers(ctx context.Context, in *userv1.AddGroupRoleMembersRequest) (*userv1.AddGroupRoleMembersResponse, error)
	RemoveGroupRoleMember(ctx context.Context, in *userv1.RemoveGroupRoleMemberRequest) (*userv1.RemoveGroupRoleMemberResponse, error)
}

type client struct {
	stub userv1.GroupRoleServiceClient
	log  logger.Logger
	none utils.CallConfig
	read utils.CallConfig
}

// New wraps an existing GroupRoleService client stub.
func New(stub userv1.GroupRoleServiceClient, settings utils.Settings, log logger.Logger) Client {
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

func (c *client) ListGroupRoles(ctx context.Context, in *userv1.ListGroupRolesRequest) (*userv1.ListGroupRolesResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "grouprole.GroupRoleService/ListGroupRoles", func(callCtx context.Context) (*userv1.ListGroupRolesResponse, error) {
		return c.stub.ListGroupRoles(callCtx, in)
	})
}

func (c *client) UpdateGroupRole(ctx context.Context, in *userv1.UpdateGroupRoleRequest) (*userv1.UpdateGroupRoleResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "grouprole.GroupRoleService/UpdateGroupRole", func(callCtx context.Context) (*userv1.UpdateGroupRoleResponse, error) {
		return c.stub.UpdateGroupRole(callCtx, in)
	})
}

func (c *client) AddGroupRoleMembers(ctx context.Context, in *userv1.AddGroupRoleMembersRequest) (*userv1.AddGroupRoleMembersResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "grouprole.GroupRoleService/AddGroupRoleMembers", func(callCtx context.Context) (*userv1.AddGroupRoleMembersResponse, error) {
		return c.stub.AddGroupRoleMembers(callCtx, in)
	})
}

func (c *client) RemoveGroupRoleMember(ctx context.Context, in *userv1.RemoveGroupRoleMemberRequest) (*userv1.RemoveGroupRoleMemberResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "grouprole.GroupRoleService/RemoveGroupRoleMember", func(callCtx context.Context) (*userv1.RemoveGroupRoleMemberResponse, error) {
		return c.stub.RemoveGroupRoleMember(callCtx, in)
	})
}
