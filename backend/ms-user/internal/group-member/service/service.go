package service

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	authrepo "github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	groupactivitysvc "github.com/wichat/wichat/backend/ms-user/internal/group-activity-log/service"
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
	KickGroupMember(ctx context.Context, accessToken, groupPublicID, targetUserID, reason string) error
	BanGroupMembers(ctx context.Context, accessToken, groupPublicID string, targetUserIDs []string, reason string, deleteSeconds int, durationSeconds int64) (int, error)
	TimeoutGroupMember(ctx context.Context, accessToken, groupPublicID, targetUserID string, durationSeconds int, reason string) (*time.Time, error)
	ListGroupBans(ctx context.Context, accessToken, groupPublicID, search string, limit int, cursor *models.GroupMemberCursor) (*models.GroupBanPage, error)
	UnbanGroupMember(ctx context.Context, accessToken, groupPublicID, targetUserID string) error
	WithActivityLog(r activitysvc.Recorder) Service
	WithGroupActivityLog(r groupactivitysvc.Recorder) Service
	WithPresence(p presencesvc.Reader) Service
}

type service struct {
	members       groupmemberrepo.PgRepository
	groups        grouprepo.PgRepository
	roles         grouprolerepo.PgRepository
	sessions      authrepo.SessionRepository
	activity      activitysvc.Recorder
	groupActivity groupactivitysvc.Recorder
	presence      presencesvc.Reader
	secret        []byte
	log           logger.Logger
	now           func() time.Time
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

func (s service) WithGroupActivityLog(r groupactivitysvc.Recorder) Service {
	s.groupActivity = r
	return s
}

func (s service) recordGroupActivity(ctx context.Context, entry models.GroupActivityLogEntry) {
	if s.groupActivity == nil {
		return
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = s.now()
	}
	s.groupActivity.Record(ctx, entry)
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
		GroupID:    group.ID,
		UserID:     userID,
		Role:       constants.GroupRoleMember,
		JoinMethod: constants.JoinMethodOpen,
		CreatedAt:  now,
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
	s.recordGroupActivity(ctx, models.GroupActivityLogEntry{
		GroupID:    group.ID,
		ActorID:    userID,
		Category:   constants.GroupAuditCategoryMember,
		Action:     constants.GroupAuditMemberJoined,
		TargetType: constants.GroupAuditTargetMember,
		TargetID:   userID,
		Summary:    "Joined group",
		Metadata:   map[string]any{"via": "direct"},
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
	s.recordGroupActivity(ctx, models.GroupActivityLogEntry{
		GroupID:    group.ID,
		ActorID:    userID,
		Category:   constants.GroupAuditCategoryMember,
		Action:     constants.GroupAuditMemberLeft,
		TargetType: constants.GroupAuditTargetMember,
		TargetID:   userID,
		Summary:    "Left group",
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

func (s service) KickGroupMember(ctx context.Context, accessToken, groupPublicID, targetUserID, reason string) error {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return err
	}
	if userID == targetUserID {
		return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "user_id"})
	}
	_ = strings.TrimSpace(reason)

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
	kickMeta := map[string]any{}
	if trimmedReason := strings.TrimSpace(reason); trimmedReason != "" {
		kickMeta["reason"] = trimmedReason
	}
	s.recordGroupActivity(ctx, models.GroupActivityLogEntry{
		GroupID:    group.ID,
		ActorID:    userID,
		Category:   constants.GroupAuditCategoryModeration,
		Action:     constants.GroupAuditMemberKicked,
		TargetType: constants.GroupAuditTargetMember,
		TargetID:   targetUserID,
		Summary:    "Kicked member",
		Metadata:   kickMeta,
	})
	return nil
}

const maxTimeoutDurationSeconds = 604800 // 7 days

func (s service) TimeoutGroupMember(
	ctx context.Context,
	accessToken, groupPublicID, targetUserID string,
	durationSeconds int,
	reason string,
) (*time.Time, error) {
	if durationSeconds < 0 || durationSeconds > maxTimeoutDurationSeconds {
		return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "duration_seconds"})
	}
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return nil, err
	}
	if userID == targetUserID {
		return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "user_id"})
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
	targetMember, err := s.members.FindGroupMember(ctx, group.ID, targetUserID)
	if err != nil {
		return nil, err
	}
	if targetMember.Role == constants.GroupRoleOwner {
		return nil, exception.PermissionDenied("cannot timeout group owner", nil)
	}

	perms, err := s.roles.GetEffectivePermissions(ctx, group.ID, userID)
	if err != nil {
		return nil, err
	}
	if !access.HasPermission(member.Role == constants.GroupRoleOwner, perms, constants.PermTimeoutMembers) {
		return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
	}
	if member.Role != constants.GroupRoleOwner {
		allowed, err := s.filterModeratableTargets(ctx, group.ID, userID, []string{targetUserID})
		if err != nil {
			return nil, err
		}
		if len(allowed) == 0 {
			return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
		}
	}

	var timeoutUntil *time.Time
	now := s.now().UTC()
	if durationSeconds > 0 {
		until := now.Add(time.Duration(durationSeconds) * time.Second)
		timeoutUntil = &until
	}

	if err := s.members.UpdateGroupMemberTimeout(ctx, group.ID, targetUserID, timeoutUntil); err != nil {
		return nil, err
	}

	summary := "Timed out member"
	groupAction := constants.GroupAuditMemberTimedOut
	if timeoutUntil == nil {
		summary = "Cleared member timeout"
		groupAction = constants.GroupAuditMemberTimeoutRemoved
	}
	s.recordActivity(ctx, models.UserActivityLogEntry{
		UserID:     targetUserID,
		ActorID:    userID,
		Category:   constants.ActivityCategoryGroup,
		Action:     "group.member_timed_out",
		TargetType: constants.ActivityTargetGroup,
		TargetID:   group.PublicID,
		Summary:    summary,
		Metadata: map[string]any{
			"group_id":         group.ID,
			"group_name":       group.Name,
			"duration_seconds": durationSeconds,
			"reason":           strings.TrimSpace(reason),
		},
		CreatedAt: now,
	})
	s.recordGroupActivity(ctx, models.GroupActivityLogEntry{
		GroupID:    group.ID,
		ActorID:    userID,
		Category:   constants.GroupAuditCategoryModeration,
		Action:     groupAction,
		TargetType: constants.GroupAuditTargetMember,
		TargetID:   targetUserID,
		Summary:    summary,
		Metadata: map[string]any{
			"duration_seconds": durationSeconds,
			"reason":           strings.TrimSpace(reason),
		},
		CreatedAt: now,
	})
	return timeoutUntil, nil
}

func (s service) BanGroupMembers(ctx context.Context, accessToken, groupPublicID string, targetUserIDs []string, reason string, deleteSeconds int, durationSeconds int64) (int, error) {
	if durationSeconds < 0 {
		return 0, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "duration_seconds"})
	}
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

	var expiresAt *time.Time
	if durationSeconds > 0 {
		exp := s.now().Add(time.Duration(durationSeconds) * time.Second)
		expiresAt = &exp
	}

	count, err := s.members.BanMembers(ctx, group.ID, validIDs, userID, strings.TrimSpace(reason), deleteSeconds, expiresAt)
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
			s.recordGroupActivity(ctx, models.GroupActivityLogEntry{
				GroupID:    group.ID,
				ActorID:    userID,
				Category:   constants.GroupAuditCategoryModeration,
				Action:     constants.GroupAuditMemberBanned,
				TargetType: constants.GroupAuditTargetMember,
				TargetID:   targetID,
				Summary:    "Banned member",
				Metadata:   meta,
			})
		}
	}
	return count, nil
}

func (s service) requireBanManager(ctx context.Context, accessToken, groupPublicID string) (string, models.Group, error) {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return "", models.Group{}, err
	}
	publicID, err := validator.GroupPublicID(groupPublicID)
	if err != nil {
		return "", models.Group{}, err
	}
	group, err := s.groups.FindGroupByPublicID(ctx, publicID)
	if err != nil {
		return "", models.Group{}, err
	}
	member, err := s.members.FindGroupMember(ctx, group.ID, userID)
	if err != nil {
		return "", models.Group{}, exception.PermissionDenied(exception.MsgForbidden, nil)
	}
	perms, err := s.roles.GetEffectivePermissions(ctx, group.ID, userID)
	if err != nil {
		return "", models.Group{}, err
	}
	if !access.HasPermission(member.Role == constants.GroupRoleOwner, perms, constants.PermBanMembers) {
		return "", models.Group{}, exception.PermissionDenied(exception.MsgForbidden, nil)
	}
	return userID, group, nil
}

func (s service) ListGroupBans(
	ctx context.Context,
	accessToken, groupPublicID, search string,
	limit int,
	cursor *models.GroupMemberCursor,
) (*models.GroupBanPage, error) {
	_, group, err := s.requireBanManager(ctx, accessToken, groupPublicID)
	if err != nil {
		return nil, err
	}
	bans, next, total, err := s.members.ListBansPage(ctx, group.ID, search, limit, cursor)
	if err != nil {
		return nil, err
	}
	return &models.GroupBanPage{
		Bans:       bans,
		NextCursor: next,
		Total:      total,
	}, nil
}

func (s service) UnbanGroupMember(ctx context.Context, accessToken, groupPublicID, targetUserID string) error {
	actorID, group, err := s.requireBanManager(ctx, accessToken, groupPublicID)
	if err != nil {
		return err
	}
	targetID := strings.TrimSpace(targetUserID)
	if targetID == "" {
		return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "user_id"})
	}
	if err := s.members.UnbanMember(ctx, group.ID, targetID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return exception.NotFound(exception.MsgInvalidInput, map[string]any{"field": "user_id"})
		}
		return err
	}
	s.recordActivity(ctx, models.UserActivityLogEntry{
		UserID:     actorID,
		ActorID:    actorID,
		Category:   constants.ActivityCategoryGroup,
		Action:     "group.member_unbanned",
		TargetType: constants.ActivityTargetGroup,
		TargetID:   group.PublicID,
		Summary:    "Unbanned member",
		Metadata:   map[string]any{"group_id": group.ID, "target_user_id": targetID},
		CreatedAt:  s.now(),
	})
	s.recordGroupActivity(ctx, models.GroupActivityLogEntry{
		GroupID:    group.ID,
		ActorID:    actorID,
		Category:   constants.GroupAuditCategoryModeration,
		Action:     constants.GroupAuditMemberUnbanned,
		TargetType: constants.GroupAuditTargetMember,
		TargetID:   targetID,
		Summary:    "Unbanned member",
		CreatedAt:  s.now(),
	})
	return nil
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
