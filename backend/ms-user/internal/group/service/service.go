package service

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	authrepo "github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	channelservice "github.com/wichat/wichat/backend/ms-user/internal/channel/service"
	groupactivitysvc "github.com/wichat/wichat/backend/ms-user/internal/group-activity-log/service"
	groupmemberrepo "github.com/wichat/wichat/backend/ms-user/internal/group-member/repository"
	grouprolerepo "github.com/wichat/wichat/backend/ms-user/internal/group-role/repository"
	grouprepo "github.com/wichat/wichat/backend/ms-user/internal/group/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	activitysvc "github.com/wichat/wichat/backend/ms-user/internal/user-activity-log/service"
	userrepo "github.com/wichat/wichat/backend/ms-user/internal/user/repository"
	"github.com/wichat/wichat/backend/ms-user/pkg/access"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/ms-user/pkg/token"
	"github.com/wichat/wichat/backend/ms-user/pkg/validator"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/infra/minio"
)

// Service is the group lifecycle and profile use-case API.
type Service interface {
	ListMyGroups(ctx context.Context, accessToken string) ([]models.PublicGroup, error)
	CreateGroup(ctx context.Context, accessToken, name string, template ...string) (*models.PublicGroup, error)
	ReorderMyGroups(ctx context.Context, accessToken string, groupPublicIDs []string) error
	DeleteGroup(ctx context.Context, accessToken, groupID string) error
	UploadGroupIcon(ctx context.Context, accessToken, groupPublicID, contentType string, body io.Reader, size int64) (*models.PublicGroup, error)
	UpdateGroupProfile(
		ctx context.Context,
		accessToken,
		groupPublicID,
		name,
		description string,
		isPrivate bool,
		bannerColor string,
		clearLogo,
		clearBannerImage bool,
	) (*models.PublicGroup, error)
	UploadGroupBanner(ctx context.Context, accessToken, groupPublicID, contentType string, body io.Reader, size int64) (*models.PublicGroup, error)
	WithActivityLog(r activitysvc.Recorder) Service
	WithGroupActivityLog(r groupactivitysvc.Recorder) Service
}

type service struct {
	groups        grouprepo.PgRepository
	members       groupmemberrepo.PgRepository
	roles         grouprolerepo.PgRepository
	channels      channelservice.Service
	users         userrepo.PgRepository
	objects       grouprepo.MinioRepository
	sessions      authrepo.SessionRepository
	activity      activitysvc.Recorder
	groupActivity groupactivitysvc.Recorder
	secret        []byte
	log           logger.Logger
	now           func() time.Time
}

// NewService builds group use cases.
func NewService(
	groups grouprepo.PgRepository,
	members groupmemberrepo.PgRepository,
	roles grouprolerepo.PgRepository,
	channels channelservice.Service,
	users userrepo.PgRepository,
	objects grouprepo.MinioRepository,
	sessions authrepo.SessionRepository,
	secret []byte,
	log logger.Logger,
) Service {
	copied := make([]byte, len(secret))
	copy(copied, secret)

	return service{
		groups:   groups,
		members:  members,
		roles:    roles,
		channels: channels,
		users:    users,
		objects:  objects,
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

func (s service) recordActivity(ctx context.Context, entry models.UserActivityLogEntry) {
	if s.activity == nil {
		return
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = s.now()
	}
	s.activity.Record(ctx, entry)
}

func (s service) ListMyGroups(ctx context.Context, accessToken string) ([]models.PublicGroup, error) {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return nil, err
	}

	groups, err := s.groups.ListGroupsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	out := make([]models.PublicGroup, len(groups))
	for i, g := range groups {
		if s.roles != nil {
			perms, err := s.roles.GetEffectivePermissions(ctx, g.ID, userID)
			if err == nil {
				g.EffectivePermissions = perms
			}
		}
		out[i] = g.Public()
	}

	return out, nil
}

func normalizeGroupTemplate(raw string) (string, error) {
	t := strings.TrimSpace(strings.ToLower(raw))
	switch t {
	case "", "custom", "office", "project", "study", "gaming":
		return t, nil
	default:
		return "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "template"})
	}
}

func (s service) CreateGroup(ctx context.Context, accessToken, name string, template ...string) (*models.PublicGroup, error) {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return nil, err
	}

	trimmed, err := validator.GroupName(name)
	if err != nil {
		return nil, err
	}

	rawTemplate := ""
	if len(template) > 0 {
		rawTemplate = template[0]
	}
	templateKey, err := normalizeGroupTemplate(rawTemplate)
	if err != nil {
		return nil, err
	}

	groupID, err := token.NewID()
	if err != nil {
		return nil, exception.Internal(err)
	}

	now := s.now()
	var group models.Group
	var member models.GroupMember
	for range 5 {
		publicID, err := token.NewPublicID()
		if err != nil {
			return nil, exception.Internal(err)
		}
		group = models.Group{
			ID:        groupID,
			PublicID:  publicID,
			Name:      trimmed,
			LogoURL:   "",
			CreatedAt: now,
			UpdatedAt: now,
		}
		member = models.GroupMember{
			GroupID:    groupID,
			UserID:     userID,
			Role:       constants.GroupRoleOwner,
			JoinMethod: constants.JoinMethodOwner,
			CreatedAt:  now,
		}
		err = s.groups.CreateGroupWithOwner(ctx, group, member)
		if err == nil {
			break
		}
		if errors.Is(err, constants.ErrConflict) {
			continue
		}

		return nil, err
	}
	if group.PublicID == "" {
		return nil, exception.Internal(errors.New("group public id"))
	}

	if err := s.channels.SeedDefaultChannelsWithTemplate(ctx, groupID, userID, templateKey, now); err != nil {
		if delErr := s.groups.DeleteGroup(ctx, groupID); delErr != nil {
			s.log.Error("create group rollback", "group_id", groupID, "seed_err", err, "delete_err", delErr)
		}

		return nil, err
	}

	saved, err := s.groups.FindGroupByID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	pub := saved.Public()
	pub.Role = constants.GroupRoleOwner

	s.recordActivity(ctx, models.UserActivityLogEntry{
		UserID:     userID,
		ActorID:    userID,
		Category:   constants.ActivityCategoryGroup,
		Action:     constants.ActivityGroupCreated,
		TargetType: constants.ActivityTargetGroup,
		TargetID:   saved.PublicID,
		Summary:    "Created group",
		Metadata:   map[string]any{"group_id": saved.ID, "group_name": saved.Name},
		CreatedAt:  now,
	})
	s.recordGroupActivity(ctx, models.GroupActivityLogEntry{
		GroupID:    saved.ID,
		ActorID:    userID,
		Category:   constants.GroupAuditCategoryProfile,
		Action:     constants.GroupAuditGroupCreated,
		TargetType: constants.GroupAuditTargetGroup,
		TargetID:   saved.PublicID,
		TargetName: saved.Name,
		Summary:    "Created group",
		Metadata:   map[string]any{"group_name": saved.Name},
		CreatedAt:  now,
	})
	return &pub, nil
}

func (s service) ReorderMyGroups(ctx context.Context, accessToken string, groupPublicIDs []string) error {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return err
	}

	mine, err := s.groups.ListGroupsForUser(ctx, userID)
	if err != nil {
		return err
	}
	if len(groupPublicIDs) != len(mine) {
		return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_ids"})
	}

	byPublic := make(map[string]string, len(mine))
	for _, g := range mine {
		byPublic[g.PublicID] = g.ID
	}

	seen := make(map[string]struct{}, len(groupPublicIDs))
	orderedUUIDs := make([]string, 0, len(groupPublicIDs))
	for _, raw := range groupPublicIDs {
		publicID, err := validator.GroupPublicID(raw)
		if err != nil {
			return err
		}
		if _, dup := seen[publicID]; dup {
			return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_ids"})
		}
		seen[publicID] = struct{}{}

		uuid, ok := byPublic[publicID]
		if !ok {
			return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_ids"})
		}
		orderedUUIDs = append(orderedUUIDs, uuid)
	}
	if len(seen) != len(mine) {
		return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_ids"})
	}

	if err := s.groups.SetRailPositions(ctx, userID, orderedUUIDs); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_ids"})
		}
		return err
	}

	return nil
}

func (s service) DeleteGroup(ctx context.Context, accessToken, groupID string) error {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return err
	}

	publicID, err := validator.GroupPublicID(groupID)
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
	if member.Role != constants.GroupRoleOwner {
		return exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	if err := s.groups.DeleteGroup(ctx, group.ID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_id"})
		}

		return err
	}

	return nil
}

func (s service) UploadGroupIcon(
	ctx context.Context,
	accessToken,
	groupPublicID,
	contentType string,
	body io.Reader,
	size int64,
) (*models.PublicGroup, error) {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return nil, err
	}

	if err := validator.GroupIconUpload(contentType, size); err != nil {
		return nil, err
	}

	publicID, err := validator.GroupPublicID(groupPublicID)
	if err != nil {
		return nil, err
	}

	group, err := s.groups.FindGroupByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_id"})
		}

		return nil, err
	}

	member, err := s.members.FindGroupMember(ctx, group.ID, userID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
		}

		return nil, err
	}
	perms, err := s.roles.GetEffectivePermissions(ctx, group.ID, userID)
	if err != nil {
		return nil, err
	}
	if !access.HasPermission(member.Role == constants.GroupRoleOwner, perms, constants.PermManageGroup) {
		return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	logoURL, err := s.objects.PutGroupIcon(ctx, group.PublicID, contentType, body, size)
	if err != nil {
		if errors.Is(err, minio.ErrDisabled) {
			return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "file"})
		}

		return nil, err
	}

	now := s.now()
	if err := s.groups.UpdateGroupLogoURL(ctx, group.ID, logoURL, now); err != nil {
		return nil, err
	}

	saved, err := s.groups.FindGroupByID(ctx, group.ID)
	if err != nil {
		return nil, err
	}

	pub := saved.Public()
	pub.Role = member.Role

	s.recordGroupActivity(ctx, models.GroupActivityLogEntry{
		GroupID:    group.ID,
		ActorID:    userID,
		Category:   constants.GroupAuditCategoryProfile,
		Action:     constants.GroupAuditGroupIconUpdated,
		TargetType: constants.GroupAuditTargetGroup,
		TargetID:   saved.PublicID,
		TargetName: saved.Name,
		Summary:    "Updated group icon",
		CreatedAt:  now,
	})

	return &pub, nil
}

func (s service) UpdateGroupProfile(
	ctx context.Context,
	accessToken,
	groupPublicID,
	name,
	description string,
	isPrivate bool,
	bannerColor string,
	clearLogo,
	clearBannerImage bool,
) (*models.PublicGroup, error) {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return nil, err
	}

	trimmedName, err := validator.GroupName(name)
	if err != nil {
		return nil, err
	}

	trimmedDescription, err := validator.GroupDescription(description)
	if err != nil {
		return nil, err
	}

	normalizedBannerColor, err := validator.GroupBannerColor(bannerColor)
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
			return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_id"})
		}

		return nil, err
	}

	member, err := s.members.FindGroupMember(ctx, group.ID, userID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
		}

		return nil, err
	}
	perms, err := s.roles.GetEffectivePermissions(ctx, group.ID, userID)
	if err != nil {
		return nil, err
	}
	if !access.HasPermission(member.Role == constants.GroupRoleOwner, perms, constants.PermManageGroup) {
		return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	logoURL := group.LogoURL
	if clearLogo {
		if err := s.objects.RemoveGroupIcon(ctx, group.PublicID); err != nil && !errors.Is(err, minio.ErrDisabled) {
			return nil, err
		}
		logoURL = ""
	}

	bannerURL := group.BannerURL
	removeBannerObject := clearBannerImage || normalizedBannerColor != ""
	if removeBannerObject && bannerURL != "" {
		if err := s.objects.RemoveGroupBanner(ctx, group.PublicID); err != nil && !errors.Is(err, minio.ErrDisabled) {
			return nil, err
		}
		bannerURL = ""
	}
	if normalizedBannerColor != "" {
		bannerURL = ""
	}

	now := s.now()
	patch := models.GroupProfilePatch{
		Name:             trimmedName,
		Description:      trimmedDescription,
		IsPrivate:        isPrivate,
		BannerColor:      normalizedBannerColor,
		BannerURL:        bannerURL,
		LogoURL:          logoURL,
		ClearBannerAsset: removeBannerObject,
	}
	if err := s.groups.UpdateGroupProfile(ctx, group.ID, patch, now); err != nil {
		return nil, err
	}

	saved, err := s.groups.FindGroupByID(ctx, group.ID)
	if err != nil {
		return nil, err
	}

	pub := saved.Public()
	pub.Role = member.Role

	changes := map[string]models.ActivityFieldChange{}
	if group.Name != saved.Name {
		changes["name"] = models.ActivityFieldChange{Old: group.Name, New: saved.Name}
	}
	if group.Description != saved.Description {
		changes["description"] = models.ActivityFieldChange{Old: group.Description, New: saved.Description}
	}
	if group.IsPrivate != saved.IsPrivate {
		changes["is_private"] = models.ActivityFieldChange{Old: group.IsPrivate, New: saved.IsPrivate}
	}
	if group.BannerColor != saved.BannerColor {
		changes["banner_color"] = models.ActivityFieldChange{Old: group.BannerColor, New: saved.BannerColor}
	}
	if clearLogo && group.LogoURL != "" {
		changes["logo"] = models.ActivityFieldChange{Old: "set", New: "cleared"}
	}
	if removeBannerObject && group.BannerURL != "" {
		changes["banner_image"] = models.ActivityFieldChange{Old: "set", New: "cleared"}
	}

	s.recordGroupActivity(ctx, models.GroupActivityLogEntry{
		GroupID:    group.ID,
		ActorID:    userID,
		Category:   constants.GroupAuditCategoryProfile,
		Action:     constants.GroupAuditGroupProfileUpdated,
		TargetType: constants.GroupAuditTargetGroup,
		TargetID:   saved.PublicID,
		TargetName: saved.Name,
		Summary:    "Updated group profile",
		Changes:    changes,
		CreatedAt:  now,
	})

	return &pub, nil
}

func (s service) UploadGroupBanner(
	ctx context.Context,
	accessToken,
	groupPublicID,
	contentType string,
	body io.Reader,
	size int64,
) (*models.PublicGroup, error) {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return nil, err
	}

	if err := validator.GroupBannerUpload(contentType, size); err != nil {
		return nil, err
	}

	publicID, err := validator.GroupPublicID(groupPublicID)
	if err != nil {
		return nil, err
	}

	group, err := s.groups.FindGroupByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_id"})
		}

		return nil, err
	}

	member, err := s.members.FindGroupMember(ctx, group.ID, userID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
		}

		return nil, err
	}
	perms, err := s.roles.GetEffectivePermissions(ctx, group.ID, userID)
	if err != nil {
		return nil, err
	}
	if !access.HasPermission(member.Role == constants.GroupRoleOwner, perms, constants.PermManageGroup) {
		return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	if group.BannerURL != "" {
		_ = s.objects.RemoveGroupBanner(ctx, group.PublicID)
	}

	bannerURL, err := s.objects.PutGroupBanner(ctx, group.PublicID, contentType, body, size)
	if err != nil {
		if errors.Is(err, minio.ErrDisabled) {
			return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "file"})
		}

		return nil, err
	}

	now := s.now()
	patch := models.GroupProfilePatch{
		Name:        group.Name,
		Description: group.Description,
		IsPrivate:   group.IsPrivate,
		BannerColor: "",
		BannerURL:   bannerURL,
		LogoURL:     group.LogoURL,
	}
	if err := s.groups.UpdateGroupProfile(ctx, group.ID, patch, now); err != nil {
		return nil, err
	}

	saved, err := s.groups.FindGroupByID(ctx, group.ID)
	if err != nil {
		return nil, err
	}

	pub := saved.Public()
	pub.Role = member.Role

	s.recordGroupActivity(ctx, models.GroupActivityLogEntry{
		GroupID:    group.ID,
		ActorID:    userID,
		Category:   constants.GroupAuditCategoryProfile,
		Action:     constants.GroupAuditGroupBannerUpdated,
		TargetType: constants.GroupAuditTargetGroup,
		TargetID:   saved.PublicID,
		TargetName: saved.Name,
		Summary:    "Updated group banner",
		CreatedAt:  now,
	})

	return &pub, nil
}
