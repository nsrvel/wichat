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
	CreateGroupRole(ctx context.Context, in *userv1.CreateGroupRoleRequest) (*userv1.CreateGroupRoleResponse, error)
	UpdateGroupRole(ctx context.Context, in *userv1.UpdateGroupRoleRequest) (*userv1.UpdateGroupRoleResponse, error)
	DuplicateGroupRole(ctx context.Context, in *userv1.DuplicateGroupRoleRequest) (*userv1.DuplicateGroupRoleResponse, error)
	DeleteGroupRole(ctx context.Context, in *userv1.DeleteGroupRoleRequest) (*userv1.DeleteGroupRoleResponse, error)
	ListGroupRoleMembers(ctx context.Context, in *userv1.ListGroupRoleMembersRequest) (*userv1.ListGroupRoleMembersResponse, error)
	AddGroupRoleMembers(ctx context.Context, in *userv1.AddGroupRoleMembersRequest) (*userv1.AddGroupRoleMembersResponse, error)
	RemoveGroupRoleMember(ctx context.Context, in *userv1.RemoveGroupRoleMemberRequest) (*userv1.RemoveGroupRoleMemberResponse, error)
	UpdateGroupHideRoles(ctx context.Context, in *userv1.UpdateGroupHideRolesRequest) (*userv1.UpdateGroupHideRolesResponse, error)
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

func (c *client) CreateGroupRole(ctx context.Context, in *userv1.CreateGroupRoleRequest) (*userv1.CreateGroupRoleResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "grouprole.GroupRoleService/CreateGroupRole", func(callCtx context.Context) (*userv1.CreateGroupRoleResponse, error) {
		return c.stub.CreateGroupRole(callCtx, in)
	})
}

func (c *client) UpdateGroupRole(ctx context.Context, in *userv1.UpdateGroupRoleRequest) (*userv1.UpdateGroupRoleResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "grouprole.GroupRoleService/UpdateGroupRole", func(callCtx context.Context) (*userv1.UpdateGroupRoleResponse, error) {
		return c.stub.UpdateGroupRole(callCtx, in)
	})
}

func (c *client) DuplicateGroupRole(ctx context.Context, in *userv1.DuplicateGroupRoleRequest) (*userv1.DuplicateGroupRoleResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "grouprole.GroupRoleService/DuplicateGroupRole", func(callCtx context.Context) (*userv1.DuplicateGroupRoleResponse, error) {
		return c.stub.DuplicateGroupRole(callCtx, in)
	})
}

func (c *client) DeleteGroupRole(ctx context.Context, in *userv1.DeleteGroupRoleRequest) (*userv1.DeleteGroupRoleResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "grouprole.GroupRoleService/DeleteGroupRole", func(callCtx context.Context) (*userv1.DeleteGroupRoleResponse, error) {
		return c.stub.DeleteGroupRole(callCtx, in)
	})
}

func (c *client) ListGroupRoleMembers(ctx context.Context, in *userv1.ListGroupRoleMembersRequest) (*userv1.ListGroupRoleMembersResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "grouprole.GroupRoleService/ListGroupRoleMembers", func(callCtx context.Context) (*userv1.ListGroupRoleMembersResponse, error) {
		return c.stub.ListGroupRoleMembers(callCtx, in)
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

func (c *client) UpdateGroupHideRoles(ctx context.Context, in *userv1.UpdateGroupHideRolesRequest) (*userv1.UpdateGroupHideRolesResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "grouprole.GroupRoleService/UpdateGroupHideRoles", func(callCtx context.Context) (*userv1.UpdateGroupHideRolesResponse, error) {
		return c.stub.UpdateGroupHideRoles(callCtx, in)
	})
}
