package service

import (
	"context"
	"errors"
	"time"

	authrepo "github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	groupmemberrepo "github.com/wichat/wichat/backend/ms-user/internal/group-member/repository"
	grouprolerepo "github.com/wichat/wichat/backend/ms-user/internal/group-role/repository"
	grouprepo "github.com/wichat/wichat/backend/ms-user/internal/group/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/pkg/access"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/ms-user/pkg/validator"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// Service is the group role and permission use-case API.
type Service interface {
	ListGroupRoles(ctx context.Context, accessToken, groupPublicID string) ([]models.GroupRole, error)
	UpdateGroupRole(ctx context.Context, accessToken, groupPublicID, rolePublicID string, patch models.GroupRolePatch) (*models.GroupRole, error)
	AddGroupRoleMembers(ctx context.Context, accessToken, groupPublicID, rolePublicID string, userIDs []string) (*models.GroupRole, error)
	RemoveGroupRoleMember(ctx context.Context, accessToken, groupPublicID, rolePublicID, targetUserID string) (*models.GroupRole, error)
	GetEffectivePermissions(ctx context.Context, groupID, userID string) ([]string, error)
}

type service struct {
	roles    grouprolerepo.PgRepository
	groups   grouprepo.PgRepository
	members  groupmemberrepo.PgRepository
	sessions authrepo.SessionRepository
	secret   []byte
	log      logger.Logger
	now      func() time.Time
}

// NewService builds group role use cases.
func NewService(
	roles grouprolerepo.PgRepository,
	groups grouprepo.PgRepository,
	members groupmemberrepo.PgRepository,
	sessions authrepo.SessionRepository,
	secret []byte,
	log logger.Logger,
) Service {
	copied := make([]byte, len(secret))
	copy(copied, secret)

	return service{
		roles:    roles,
		groups:   groups,
		members:  members,
		sessions: sessions,
		secret:   copied,
		log:      log,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

func (s service) ListGroupRoles(ctx context.Context, accessToken, groupPublicID string) ([]models.GroupRole, error) {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return nil, err
	}
	publicID, err := validator.GroupPublicID(groupPublicID)
	if err != nil {
		return nil, err
	}
	group, err := s.groups.FindGroupByPublicID(ctx, publicID)
	if err != nil {
		return nil, err
	}
	_, err = s.members.FindGroupMember(ctx, group.ID, userID)
	if err != nil {
		return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	return s.roles.ListRolesForGroup(ctx, group.ID)
}

func (s service) UpdateGroupRole(ctx context.Context, accessToken, groupPublicID, rolePublicID string, patch models.GroupRolePatch) (*models.GroupRole, error) {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return nil, err
	}
	publicID, err := validator.GroupPublicID(groupPublicID)
	if err != nil {
		return nil, err
	}
	group, err := s.groups.FindGroupByPublicID(ctx, publicID)
	if err != nil {
		return nil, err
	}
	member, err := s.members.FindGroupMember(ctx, group.ID, userID)
	if err != nil {
		return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	perms, err := s.roles.GetEffectivePermissions(ctx, group.ID, userID)
	if err != nil {
		return nil, err
	}
	if !access.HasPermission(member.Role == constants.GroupRoleOwner, perms, constants.PermManageRoles) {
		return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	role, err := s.roles.UpdateRole(ctx, group.ID, rolePublicID, patch, s.now())
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (s service) AddGroupRoleMembers(ctx context.Context, accessToken, groupPublicID, rolePublicID string, userIDs []string) (*models.GroupRole, error) {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return nil, err
	}
	publicID, err := validator.GroupPublicID(groupPublicID)
	if err != nil {
		return nil, err
	}
	group, err := s.groups.FindGroupByPublicID(ctx, publicID)
	if err != nil {
		return nil, err
	}
	member, err := s.members.FindGroupMember(ctx, group.ID, userID)
	if err != nil {
		return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	perms, err := s.roles.GetEffectivePermissions(ctx, group.ID, userID)
	if err != nil {
		return nil, err
	}
	if !access.HasPermission(member.Role == constants.GroupRoleOwner, perms, constants.PermManageRoles) {
		return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	if err := s.roles.AddMemberRoles(ctx, group.ID, rolePublicID, userIDs); err != nil {
		return nil, err
	}
	role, err := s.roles.FindRoleByID(ctx, group.ID, rolePublicID)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (s service) RemoveGroupRoleMember(ctx context.Context, accessToken, groupPublicID, rolePublicID, targetUserID string) (*models.GroupRole, error) {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return nil, err
	}
	publicID, err := validator.GroupPublicID(groupPublicID)
	if err != nil {
		return nil, err
	}
	group, err := s.groups.FindGroupByPublicID(ctx, publicID)
	if err != nil {
		return nil, err
	}
	member, err := s.members.FindGroupMember(ctx, group.ID, userID)
	if err != nil {
		return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	perms, err := s.roles.GetEffectivePermissions(ctx, group.ID, userID)
	if err != nil {
		return nil, err
	}
	if !access.HasPermission(member.Role == constants.GroupRoleOwner, perms, constants.PermManageRoles) {
		return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	if err := s.roles.RemoveMemberRole(ctx, group.ID, rolePublicID, targetUserID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, exception.NotFound("member role not found", nil)
		}
		return nil, err
	}
	role, err := s.roles.FindRoleByID(ctx, group.ID, rolePublicID)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (s service) GetEffectivePermissions(ctx context.Context, groupID, userID string) ([]string, error) {
	return s.roles.GetEffectivePermissions(ctx, groupID, userID)
}
