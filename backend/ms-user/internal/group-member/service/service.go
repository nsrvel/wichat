package service

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	authrepo "github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	groupmemberrepo "github.com/wichat/wichat/backend/ms-user/internal/group-member/repository"
	grouprolerepo "github.com/wichat/wichat/backend/ms-user/internal/group-role/repository"
	grouprepo "github.com/wichat/wichat/backend/ms-user/internal/group/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	presencesvc "github.com/wichat/wichat/backend/ms-user/internal/presence/service"
	activitysvc "github.com/wichat/wichat/backend/ms-user/internal/user-activity-log/service"
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
	WithActivityLog(r activitysvc.Recorder) Service
	WithPresence(p presencesvc.Reader) Service
}

type service struct {
	members  groupmemberrepo.PgRepository
	groups   grouprepo.PgRepository
	roles    grouprolerepo.PgRepository
	sessions authrepo.SessionRepository
	activity activitysvc.Recorder
	presence presencesvc.Reader
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

func (s service) WithActivityLog(r activitysvc.Recorder) Service {
	s.activity = r
	return s
}

func (s service) WithPresence(p presencesvc.Reader) Service {
	s.presence = p
	return s
}

func (s service) recordActivity(ctx context.Context, entry models.UserActivityLogEntry) {
	if s.activity == nil {
		return
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = s.now()
	}
	s.activity.Record(ctx, entry)
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
	if err := s.members.InsertGroupMember(ctx, models.GroupMember{
		GroupID:   group.ID,
		UserID:    userID,
		Role:      constants.GroupRoleMember,
		CreatedAt: now,
	}); err != nil {
		return err
	}
	s.recordActivity(ctx, models.UserActivityLogEntry{
		UserID:     userID,
		ActorID:    userID,
		Category:   constants.ActivityCategoryGroup,
		Action:     constants.ActivityGroupJoined,
		TargetType: constants.ActivityTargetGroup,
		TargetID:   group.PublicID,
		Summary:    "Joined group",
		Metadata:   map[string]any{"group_id": group.ID, "group_name": group.Name, "via": "direct"},
		CreatedAt:  now,
	})
	return nil
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

	s.recordActivity(ctx, models.UserActivityLogEntry{
		UserID:     userID,
		ActorID:    userID,
		Category:   constants.ActivityCategoryGroup,
		Action:     constants.ActivityGroupLeft,
		TargetType: constants.ActivityTargetGroup,
		TargetID:   group.PublicID,
		Summary:    "Left group",
		Metadata:   map[string]any{"group_id": group.ID, "group_name": group.Name},
	})
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

	if len(members) > 0 {
		userIDs := make([]string, len(members))
		for i, m := range members {
			userIDs[i] = m.UserID
		}
		if s.roles != nil {
			if rolesMap, err := s.roles.ListRolesForMembers(ctx, group.ID, userIDs); err == nil {
				for i := range members {
					members[i].Roles = rolesMap[members[i].UserID]
				}
			}
		}
		if s.presence != nil {
			if presenceMap, err := s.presence.BatchGetEffectivePresence(ctx, userIDs); err == nil {
				for i := range members {
					if snap, ok := presenceMap[members[i].UserID]; ok {
						members[i].IsOnline = snap.IsOnline
						members[i].PresenceStatus = snap.PresenceStatus
					}
				}
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
	if member.Role != constants.GroupRoleOwner {
		allowed, err := s.filterModeratableTargets(ctx, group.ID, userID, []string{targetUserID})
		if err != nil {
			return err
		}
		if len(allowed) == 0 {
			return exception.PermissionDenied(exception.MsgForbidden, nil)
		}
	}

	if err := s.members.DeleteGroupMember(ctx, group.ID, targetUserID); err != nil {
		return err
	}
	s.recordActivity(ctx, models.UserActivityLogEntry{
		UserID:     targetUserID,
		ActorID:    userID,
		Category:   constants.ActivityCategoryGroup,
		Action:     constants.ActivityGroupKicked,
		TargetType: constants.ActivityTargetGroup,
		TargetID:   group.PublicID,
		Summary:    "Removed from group",
		Metadata:   map[string]any{"group_id": group.ID, "group_name": group.Name},
	})
	return nil
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
	if member.Role != constants.GroupRoleOwner {
		filtered, err := s.filterModeratableTargets(ctx, group.ID, userID, validIDs)
		if err != nil {
			return 0, err
		}
		if len(filtered) == 0 {
			return 0, exception.PermissionDenied(exception.MsgForbidden, nil)
		}
		validIDs = filtered
	}

	count, err := s.members.BanMembers(ctx, group.ID, validIDs, userID, reason, deleteSeconds)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		trimmedReason := strings.TrimSpace(reason)
		for _, targetID := range validIDs {
			meta := map[string]any{"group_id": group.ID, "group_name": group.Name}
			if trimmedReason != "" {
				meta["reason"] = trimmedReason
			}
			s.recordActivity(ctx, models.UserActivityLogEntry{
				UserID:     targetID,
				ActorID:    userID,
				Category:   constants.ActivityCategoryGroup,
				Action:     constants.ActivityGroupBanned,
				TargetType: constants.ActivityTargetGroup,
				TargetID:   group.PublicID,
				Summary:    "Banned from group",
				Metadata:   meta,
			})
		}
	}
	return count, nil
}

func bestCustomRolePosition(roles []models.GroupRoleSummary) int32 {
	best := int32(math.MaxInt32)
	for _, r := range roles {
		if r.Slug == "everyone" {
			continue
		}
		if r.Position < best {
			best = r.Position
		}
	}
	return best
}

func (s service) filterModeratableTargets(ctx context.Context, groupID, actorID string, targetIDs []string) ([]string, error) {
	queryIDs := append([]string{actorID}, targetIDs...)
	roleMap, err := s.roles.ListRolesForMembers(ctx, groupID, queryIDs)
	if err != nil {
		return nil, err
	}
	actorBest := bestCustomRolePosition(roleMap[actorID])
	allowed := make([]string, 0, len(targetIDs))
	for _, tid := range targetIDs {
		targetBest := bestCustomRolePosition(roleMap[tid])
		if targetBest == math.MaxInt32 || actorBest < targetBest {
			allowed = append(allowed, tid)
		}
	}
	return allowed, nil
}
