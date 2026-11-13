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

// Service is the group membership, list, kick, and ban use-case API.
type Service interface {
	JoinGroup(ctx context.Context, accessToken, groupPublicID string) error
	LeaveGroup(ctx context.Context, accessToken, groupPublicID string) error
	ListGroupMembers(ctx context.Context, accessToken, groupPublicID string, limit int, cursor *models.GroupMemberCursor) (*models.GroupMemberPage, error)
	KickGroupMember(ctx context.Context, accessToken, groupPublicID, targetUserID string) error
	BanGroupMembers(ctx context.Context, accessToken, groupPublicID string, targetUserIDs []string, reason string, deleteSeconds int) (int, error)
}

type service struct {
	members  groupmemberrepo.PgRepository
	groups   grouprepo.PgRepository
	roles    grouprolerepo.PgRepository
	sessions authrepo.SessionRepository
	secret   []byte
	log      logger.Logger
	now      func() time.Time
}

// NewService builds group member use cases.
func NewService(
	members groupmemberrepo.PgRepository,
	groups grouprepo.PgRepository,
	roles grouprolerepo.PgRepository,
	sessions authrepo.SessionRepository,
	secret []byte,
	log logger.Logger,
) Service {
	copied := make([]byte, len(secret))
	copy(copied, secret)

	return service{
		members:  members,
		groups:   groups,
		roles:    roles,
		sessions: sessions,
		secret:   copied,
		log:      log,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

func (s service) JoinGroup(ctx context.Context, accessToken, groupPublicID string) error {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return err
	}

	publicID, err := validator.GroupPublicID(groupPublicID)
	if err != nil {
		return err
	}

	group, err := s.groups.FindGroupByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_id"})
		}
		return err
	}

	if banned, err := s.members.IsMemberBanned(ctx, group.ID, userID); err != nil {
		return err
	} else if banned {
		return exception.PermissionDenied("user is banned from this group", nil)
	}

	if _, err := s.members.FindGroupMember(ctx, group.ID, userID); err != nil {
		if !errors.Is(err, constants.ErrNotFound) {
			return err
		}
	} else {
		return nil
	}

	now := s.now()
	return s.members.InsertGroupMember(ctx, models.GroupMember{
		GroupID:   group.ID,
		UserID:    userID,
		Role:      constants.GroupRoleMember,
		CreatedAt: now,
	})
}

func (s service) LeaveGroup(ctx context.Context, accessToken, groupPublicID string) error {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return err
	}

	publicID, err := validator.GroupPublicID(groupPublicID)
	if err != nil {
		return err
	}

	group, err := s.groups.FindGroupByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_id"})
		}
		return err
	}

	member, err := s.members.FindGroupMember(ctx, group.ID, userID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_id"})
		}
		return err
	}
	if member.Role == constants.GroupRoleOwner {
		return exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	if err := s.members.DeleteGroupMember(ctx, group.ID, userID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_id"})
		}
		return err
	}

	return nil
}

func (s service) ListGroupMembers(
	ctx context.Context,
	accessToken, groupPublicID string,
	limit int,
	cursor *models.GroupMemberCursor,
) (*models.GroupMemberPage, error) {
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
		if errors.Is(err, constants.ErrNotFound) {
			return nil, exception.NotFound(exception.MsgInvalidInput, map[string]any{"field": "group_id"})
		}
		return nil, err
	}

	if _, err := s.members.FindGroupMember(ctx, group.ID, userID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
		}
		return nil, err
	}

	total, err := s.members.CountGroupMembers(ctx, group.ID)
	if err != nil {
		return nil, err
	}

	members, next, err := s.members.ListGroupMembersPage(ctx, group.ID, limit, cursor)
	if err != nil {
		return nil, err
	}

	if len(members) > 0 && s.roles != nil {
		userIDs := make([]string, len(members))
		for i, m := range members {
			userIDs[i] = m.UserID
		}
		if rolesMap, err := s.roles.ListRolesForMembers(ctx, group.ID, userIDs); err == nil {
			for i := range members {
				members[i].Roles = rolesMap[members[i].UserID]
			}
		}
	}

	return &models.GroupMemberPage{Members: members, NextCursor: next, Total: total}, nil
}

func (s service) KickGroupMember(ctx context.Context, accessToken, groupPublicID, targetUserID string) error {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return err
	}
	if userID == targetUserID {
		return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "user_id"})
	}

	publicID, err := validator.GroupPublicID(groupPublicID)
	if err != nil {
		return err
	}
	group, err := s.groups.FindGroupByPublicID(ctx, publicID)
	if err != nil {
		return err
	}
	member, err := s.members.FindGroupMember(ctx, group.ID, userID)
	if err != nil {
		return exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	targetMember, err := s.members.FindGroupMember(ctx, group.ID, targetUserID)
	if err != nil {
		return err
	}
	if targetMember.Role == constants.GroupRoleOwner {
		return exception.PermissionDenied("cannot kick group owner", nil)
	}

	perms, err := s.roles.GetEffectivePermissions(ctx, group.ID, userID)
	if err != nil {
		return err
	}
	if !access.HasPermission(member.Role == constants.GroupRoleOwner, perms, constants.PermKickMembers) {
		return exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	return s.members.DeleteGroupMember(ctx, group.ID, targetUserID)
}

func (s service) BanGroupMembers(ctx context.Context, accessToken, groupPublicID string, targetUserIDs []string, reason string, deleteSeconds int) (int, error) {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return 0, err
	}
	publicID, err := validator.GroupPublicID(groupPublicID)
	if err != nil {
		return 0, err
	}
	group, err := s.groups.FindGroupByPublicID(ctx, publicID)
	if err != nil {
		return 0, err
	}
	member, err := s.members.FindGroupMember(ctx, group.ID, userID)
	if err != nil {
		return 0, exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	perms, err := s.roles.GetEffectivePermissions(ctx, group.ID, userID)
	if err != nil {
		return 0, err
	}
	if !access.HasPermission(member.Role == constants.GroupRoleOwner, perms, constants.PermBanMembers) {
		return 0, exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	validIDs := make([]string, 0, len(targetUserIDs))
	for _, id := range targetUserIDs {
		if id != userID && id != "" {
			validIDs = append(validIDs, id)
		}
	}
	if len(validIDs) == 0 {
		return 0, nil
	}

	return s.members.BanMembers(ctx, group.ID, validIDs, userID, reason, deleteSeconds)
}
