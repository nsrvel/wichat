package service

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	authrepo "github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	groupactivitysvc "github.com/wichat/wichat/backend/ms-user/internal/group-activity-log/service"
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
	ListGroupRolesWithSettings(ctx context.Context, accessToken, groupPublicID string) ([]models.GroupRole, bool, error)
	CreateGroupRole(ctx context.Context, accessToken, groupPublicID, name string) (*models.GroupRole, error)
	UpdateGroupRole(ctx context.Context, accessToken, groupPublicID, rolePublicID string, patch models.GroupRolePatch) (*models.GroupRole, error)
	DuplicateGroupRole(ctx context.Context, accessToken, groupPublicID, rolePublicID string) (*models.GroupRole, error)
	DeleteGroupRole(ctx context.Context, accessToken, groupPublicID, rolePublicID string) error
	ListGroupRoleMembers(ctx context.Context, accessToken, groupPublicID, rolePublicID string) ([]models.GroupMemberProfile, error)
	AddGroupRoleMembers(ctx context.Context, accessToken, groupPublicID, rolePublicID string, userIDs []string) (*models.GroupRole, error)
	RemoveGroupRoleMember(ctx context.Context, accessToken, groupPublicID, rolePublicID, targetUserID string) (*models.GroupRole, error)
	UpdateGroupHideRoles(ctx context.Context, accessToken, groupPublicID string, hideRoles bool) (bool, error)
	GetEffectivePermissions(ctx context.Context, groupID, userID string) ([]string, error)
	WithGroupActivityLog(r groupactivitysvc.Recorder) Service
}

type service struct {
	roles         grouprolerepo.PgRepository
	groups        grouprepo.PgRepository
	members       groupmemberrepo.PgRepository
	sessions      authrepo.SessionRepository
	groupActivity groupactivitysvc.Recorder
	secret        []byte
	log           logger.Logger
	now           func() time.Time
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

func (s service) requireGroupMember(ctx context.Context, accessToken, groupPublicID string) (string, models.Group, models.GroupMember, error) {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return "", models.Group{}, models.GroupMember{}, err
	}
	publicID, err := validator.GroupPublicID(groupPublicID)
	if err != nil {
		return "", models.Group{}, models.GroupMember{}, err
	}
	group, err := s.groups.FindGroupByPublicID(ctx, publicID)
	if err != nil {
		return "", models.Group{}, models.GroupMember{}, err
	}
	member, err := s.members.FindGroupMember(ctx, group.ID, userID)
	if err != nil {
		return "", models.Group{}, models.GroupMember{}, exception.PermissionDenied(exception.MsgForbidden, nil)
	}
	return userID, group, member, nil
}

func (s service) requireRoleManager(ctx context.Context, accessToken, groupPublicID string) (string, models.Group, models.GroupMember, []string, error) {
	userID, group, member, err := s.requireGroupMember(ctx, accessToken, groupPublicID)
	if err != nil {
		return "", models.Group{}, models.GroupMember{}, nil, err
	}
	perms, err := s.roles.GetEffectivePermissions(ctx, group.ID, userID)
	if err != nil {
		return "", models.Group{}, models.GroupMember{}, nil, err
	}
	if !access.HasPermission(member.Role == constants.GroupRoleOwner, perms, constants.PermManageRoles) {
		return "", models.Group{}, models.GroupMember{}, nil, exception.PermissionDenied(exception.MsgForbidden, nil)
	}
	return userID, group, member, perms, nil
}

func (s service) ListGroupRoles(ctx context.Context, accessToken, groupPublicID string) ([]models.GroupRole, error) {
	_, group, _, err := s.requireGroupMember(ctx, accessToken, groupPublicID)
	if err != nil {
		return nil, err
	}
	return s.roles.ListRolesForGroup(ctx, group.ID)
}

func (s service) ListGroupRolesWithSettings(ctx context.Context, accessToken, groupPublicID string) ([]models.GroupRole, bool, error) {
	_, group, _, err := s.requireGroupMember(ctx, accessToken, groupPublicID)
	if err != nil {
		return nil, false, err
	}
	roles, err := s.roles.ListRolesForGroup(ctx, group.ID)
	if err != nil {
		return nil, false, err
	}
	hideRoles, err := s.roles.GetHideRoles(ctx, group.ID)
	if err != nil {
		hideRoles = group.HideRoles
	}
	return roles, hideRoles, nil
}

func (s service) CreateGroupRole(ctx context.Context, accessToken, groupPublicID, rawName string) (*models.GroupRole, error) {
	actorID, group, _, _, err := s.requireRoleManager(ctx, accessToken, groupPublicID)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(rawName)
	if name == "" {
		name = "new role"
	}
	validName, err := validator.RoleName(name)
	if err != nil {
		return nil, err
	}
	role, err := s.roles.CreateRole(ctx, group.ID, validName, "#99aab5", false, false, []string{}, s.now())
	if err != nil {
		return nil, err
	}
	s.recordGroupActivity(ctx, models.GroupActivityLogEntry{
		GroupID:    group.ID,
		ActorID:    actorID,
		Category:   constants.GroupAuditCategoryRole,
		Action:     constants.GroupAuditRoleCreated,
		TargetType: constants.GroupAuditTargetRole,
		TargetID:   role.ID,
		TargetName: role.Name,
		Summary:    "Created role",
		Metadata:   map[string]any{"role_name": role.Name},
	})
	return &role, nil
}

func (s service) UpdateGroupRole(ctx context.Context, accessToken, groupPublicID, rolePublicID string, patch models.GroupRolePatch) (*models.GroupRole, error) {
	actorID, group, member, callerPerms, err := s.requireRoleManager(ctx, accessToken, groupPublicID)
	if err != nil {
		return nil, err
	}

	existing, err := s.roles.FindRoleByID(ctx, group.ID, rolePublicID)
	if err != nil {
		return nil, mapRoleErr(err)
	}

	if patch.Name != nil {
		name, err := validator.RoleName(*patch.Name)
		if err != nil {
			return nil, err
		}
		patch.Name = &name
	}
	if patch.Color != nil {
		color, err := validator.RoleColor(*patch.Color)
		if err != nil {
			return nil, err
		}
		patch.Color = &color
	}
	if patch.Permissions != nil {
		validPerms, err := validator.RolePermissions(patch.Permissions)
		if err != nil {
			return nil, err
		}
		callerIsAdmin := member.Role == constants.GroupRoleOwner || slices.Contains(callerPerms, constants.PermAdministrator)
		if !callerIsAdmin && (slices.Contains(validPerms, constants.PermAdministrator) != slices.Contains(existing.Permissions, constants.PermAdministrator)) {
			return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
		}
		patch.Permissions = validPerms
	}

	role, err := s.roles.UpdateRole(ctx, group.ID, rolePublicID, patch, s.now())
	if err != nil {
		return nil, mapRoleErr(err)
	}

	changes := map[string]models.ActivityFieldChange{}
	if existing.Name != role.Name {
		changes["name"] = models.ActivityFieldChange{Old: existing.Name, New: role.Name}
	}
	if existing.Color != role.Color {
		changes["color"] = models.ActivityFieldChange{Old: existing.Color, New: role.Color}
	}
	if existing.Hoist != role.Hoist {
		changes["hoist"] = models.ActivityFieldChange{Old: existing.Hoist, New: role.Hoist}
	}
	if existing.Mentionable != role.Mentionable {
		changes["mentionable"] = models.ActivityFieldChange{Old: existing.Mentionable, New: role.Mentionable}
	}
	if patch.Permissions != nil && !slices.Equal(existing.Permissions, role.Permissions) {
		changes["permissions"] = models.ActivityFieldChange{Old: len(existing.Permissions), New: len(role.Permissions)}
	}

	s.recordGroupActivity(ctx, models.GroupActivityLogEntry{
		GroupID:    group.ID,
		ActorID:    actorID,
		Category:   constants.GroupAuditCategoryRole,
		Action:     constants.GroupAuditRoleUpdated,
		TargetType: constants.GroupAuditTargetRole,
		TargetID:   role.ID,
		TargetName: role.Name,
		Summary:    "Updated role",
		Changes:    changes,
		Metadata:   map[string]any{"role_name": role.Name},
	})
	return &role, nil
}

func (s service) DuplicateGroupRole(ctx context.Context, accessToken, groupPublicID, rolePublicID string) (*models.GroupRole, error) {
	actorID, group, member, callerPerms, err := s.requireRoleManager(ctx, accessToken, groupPublicID)
	if err != nil {
		return nil, err
	}
	source, err := s.roles.FindRoleByID(ctx, group.ID, rolePublicID)
	if err != nil {
		return nil, mapRoleErr(err)
	}

	callerIsAdmin := member.Role == constants.GroupRoleOwner || slices.Contains(callerPerms, constants.PermAdministrator)
	perms := make([]string, 0, len(source.Permissions))
	for _, p := range source.Permissions {
		if p == constants.PermAdministrator && !callerIsAdmin {
			continue
		}
		perms = append(perms, p)
	}

	copyName := buildDuplicateRoleName(source.Name)
	role, err := s.roles.CreateRole(ctx, group.ID, copyName, source.Color, source.Hoist, source.Mentionable, perms, s.now())
	if err != nil {
		return nil, err
	}
	s.recordGroupActivity(ctx, models.GroupActivityLogEntry{
		GroupID:    group.ID,
		ActorID:    actorID,
		Category:   constants.GroupAuditCategoryRole,
		Action:     constants.GroupAuditRoleDuplicated,
		TargetType: constants.GroupAuditTargetRole,
		TargetID:   role.ID,
		TargetName: role.Name,
		Summary:    "Duplicated role",
		Metadata:   map[string]any{"source_role_name": source.Name, "role_name": role.Name},
	})
	return &role, nil
}

func (s service) DeleteGroupRole(ctx context.Context, accessToken, groupPublicID, rolePublicID string) error {
	actorID, group, _, _, err := s.requireRoleManager(ctx, accessToken, groupPublicID)
	if err != nil {
		return err
	}
	existing, err := s.roles.FindRoleByID(ctx, group.ID, rolePublicID)
	if err != nil {
		return mapRoleErr(err)
	}
	if err := s.roles.DeleteRole(ctx, group.ID, rolePublicID); err != nil {
		if errors.Is(err, constants.ErrConflict) {
			return exception.InvalidInput("cannot delete system or default role", nil)
		}
		return mapRoleErr(err)
	}
	s.recordGroupActivity(ctx, models.GroupActivityLogEntry{
		GroupID:    group.ID,
		ActorID:    actorID,
		Category:   constants.GroupAuditCategoryRole,
		Action:     constants.GroupAuditRoleDeleted,
		TargetType: constants.GroupAuditTargetRole,
		TargetID:   existing.ID,
		TargetName: existing.Name,
		Summary:    "Deleted role",
		Metadata:   map[string]any{"role_name": existing.Name},
	})
	return nil
}

func (s service) ListGroupRoleMembers(ctx context.Context, accessToken, groupPublicID, rolePublicID string) ([]models.GroupMemberProfile, error) {
	_, group, _, err := s.requireGroupMember(ctx, accessToken, groupPublicID)
	if err != nil {
		return nil, err
	}
	members, err := s.roles.ListRoleMembers(ctx, group.ID, rolePublicID)
	if err != nil {
		return nil, mapRoleErr(err)
	}
	return members, nil
}

func (s service) AddGroupRoleMembers(ctx context.Context, accessToken, groupPublicID, rolePublicID string, userIDs []string) (*models.GroupRole, error) {
	actorID, group, _, _, err := s.requireRoleManager(ctx, accessToken, groupPublicID)
	if err != nil {
		return nil, err
	}
	if err := s.roles.AddMemberRoles(ctx, group.ID, rolePublicID, userIDs); err != nil {
		return nil, mapRoleErr(err)
	}
	role, err := s.roles.FindRoleByID(ctx, group.ID, rolePublicID)
	if err != nil {
		return nil, mapRoleErr(err)
	}
	for _, uid := range userIDs {
		s.recordGroupActivity(ctx, models.GroupActivityLogEntry{
			GroupID:    group.ID,
			ActorID:    actorID,
			Category:   constants.GroupAuditCategoryRole,
			Action:     constants.GroupAuditMemberRoleAdded,
			TargetType: constants.GroupAuditTargetMember,
			TargetID:   uid,
			TargetName: role.Name,
			Summary:    "Added role to member",
			Metadata:   map[string]any{"role_id": role.ID, "role_name": role.Name},
		})
	}
	return &role, nil
}

func (s service) RemoveGroupRoleMember(ctx context.Context, accessToken, groupPublicID, rolePublicID, targetUserID string) (*models.GroupRole, error) {
	actorID, group, _, _, err := s.requireRoleManager(ctx, accessToken, groupPublicID)
	if err != nil {
		return nil, err
	}
	if err := s.roles.RemoveMemberRole(ctx, group.ID, rolePublicID, targetUserID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, exception.NotFound("member role not found", nil)
		}
		return nil, mapRoleErr(err)
	}
	role, err := s.roles.FindRoleByID(ctx, group.ID, rolePublicID)
	if err != nil {
		return nil, mapRoleErr(err)
	}
	s.recordGroupActivity(ctx, models.GroupActivityLogEntry{
		GroupID:    group.ID,
		ActorID:    actorID,
		Category:   constants.GroupAuditCategoryRole,
		Action:     constants.GroupAuditMemberRoleRemoved,
		TargetType: constants.GroupAuditTargetMember,
		TargetID:   targetUserID,
		TargetName: role.Name,
		Summary:    "Removed role from member",
		Metadata:   map[string]any{"role_id": role.ID, "role_name": role.Name},
	})
	return &role, nil
}

func (s service) UpdateGroupHideRoles(ctx context.Context, accessToken, groupPublicID string, hideRoles bool) (bool, error) {
	actorID, group, _, _, err := s.requireRoleManager(ctx, accessToken, groupPublicID)
	if err != nil {
		return false, err
	}
	updated, err := s.roles.SetHideRoles(ctx, group.ID, hideRoles, s.now())
	if err != nil {
		return false, mapRoleErr(err)
	}
	s.recordGroupActivity(ctx, models.GroupActivityLogEntry{
		GroupID:    group.ID,
		ActorID:    actorID,
		Category:   constants.GroupAuditCategoryProfile,
		Action:     constants.GroupAuditGroupHideRolesUpdated,
		TargetType: constants.GroupAuditTargetGroup,
		TargetID:   group.PublicID,
		TargetName: group.Name,
		Summary:    "Updated hide roles setting",
		Changes:    map[string]models.ActivityFieldChange{"hide_roles": {Old: group.HideRoles, New: updated}},
	})
	return updated, nil
}

func (s service) GetEffectivePermissions(ctx context.Context, groupID, userID string) ([]string, error) {
	return s.roles.GetEffectivePermissions(ctx, groupID, userID)
}

func mapRoleErr(err error) error {
	if errors.Is(err, constants.ErrNotFound) {
		return exception.NotFound("role not found", nil)
	}
	return err
}

func buildDuplicateRoleName(orig string) string {
	const suffix = " Copy"
	maxBase := constants.MaxRoleNameRunes - utf8.RuneCountInString(suffix)
	runes := []rune(strings.TrimSpace(orig))
	if len(runes) > maxBase {
		runes = runes[:maxBase]
	}
	base := strings.TrimSpace(string(runes))
	if base == "" {
		base = "role"
	}
	return base + suffix
}
