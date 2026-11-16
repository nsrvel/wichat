package service

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	authrepo "github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	categoryrepo "github.com/wichat/wichat/backend/ms-user/internal/channel-category/repository"
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

// Service manages channel categories and category permission overrides.
type Service interface {
	CreateGroupCategory(ctx context.Context, accessToken, groupPublicID, name string, isPrivate bool) (models.PublicChannelCategory, error)
	UpdateGroupCategory(ctx context.Context, accessToken, groupPublicID, categoryPublicID string, name *string, isPrivate *bool) (models.PublicChannelCategory, error)
	DeleteGroupCategory(ctx context.Context, accessToken, groupPublicID, categoryPublicID string) error
	ListCategoryPermissions(ctx context.Context, accessToken, groupPublicID, categoryPublicID string) ([]models.PublicChannelCategoryPermissionOverride, error)
	UpsertCategoryPermission(ctx context.Context, accessToken, groupPublicID, categoryPublicID, targetType, targetPublicID string, allow, deny []string) (models.PublicChannelCategoryPermissionOverride, error)
	DeleteCategoryPermission(ctx context.Context, accessToken, groupPublicID, categoryPublicID, targetType, targetPublicID string) error
	ReorderChannelCategories(ctx context.Context, accessToken, groupPublicID string, categoryPublicIDs []string) error
	WithGroupActivityLog(r groupactivitysvc.Recorder) Service
}

type service struct {
	categories    categoryrepo.PgRepository
	groups        grouprepo.PgRepository
	members       groupmemberrepo.PgRepository
	roles         grouprolerepo.PgRepository
	sessions      authrepo.SessionRepository
	groupActivity groupactivitysvc.Recorder
	secret        []byte
	log           logger.Logger
	now           func() time.Time
}

// NewService builds channel category use cases.
func NewService(
	categories categoryrepo.PgRepository,
	groups grouprepo.PgRepository,
	members groupmemberrepo.PgRepository,
	roles grouprolerepo.PgRepository,
	sessions authrepo.SessionRepository,
	secret []byte,
	log logger.Logger,
) Service {
	copied := make([]byte, len(secret))
	copy(copied, secret)

	return service{
		categories: categories,
		groups:     groups,
		members:    members,
		roles:      roles,
		sessions:   sessions,
		secret:     copied,
		log:        log,
		now:        func() time.Time { return time.Now().UTC() },
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

func (s service) CreateGroupCategory(
	ctx context.Context,
	accessToken, groupPublicID, name string,
	isPrivate bool,
) (models.PublicChannelCategory, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" || utf8.RuneCountInString(trimmed) > 64 {
		return models.PublicChannelCategory{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "name"})
	}

	userID, group, err := s.resolveGroupWithManageChannels(ctx, accessToken, groupPublicID)
	if err != nil {
		return models.PublicChannelCategory{}, err
	}

	cat, err := s.categories.CreateCategory(ctx, group.ID, trimmed, isPrivate, userID, s.now())
	if err != nil {
		return models.PublicChannelCategory{}, err
	}

	s.recordGroupActivity(ctx, models.GroupActivityLogEntry{
		GroupID:    group.ID,
		ActorID:    userID,
		Category:   "channel",
		Action:     "category.create",
		TargetType: "category",
		TargetID:   cat.PublicID,
		Metadata:   map[string]any{"name": cat.Name, "is_private": cat.IsPrivate},
	})

	return cat.Public(), nil
}

func (s service) UpdateGroupCategory(
	ctx context.Context,
	accessToken, groupPublicID, categoryPublicID string,
	name *string,
	isPrivate *bool,
) (models.PublicChannelCategory, error) {
	catPubID := strings.TrimSpace(categoryPublicID)
	if catPubID == "" {
		return models.PublicChannelCategory{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "category_id"})
	}

	var trimmedPtr *string
	if name != nil {
		trimmed := strings.TrimSpace(*name)
		if trimmed == "" || utf8.RuneCountInString(trimmed) > 64 {
			return models.PublicChannelCategory{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "name"})
		}
		trimmedPtr = &trimmed
	}
	if trimmedPtr == nil && isPrivate == nil {
		return models.PublicChannelCategory{}, exception.InvalidInput(exception.MsgInvalidInput, nil)
	}

	userID, group, err := s.resolveGroupWithManageChannels(ctx, accessToken, groupPublicID)
	if err != nil {
		return models.PublicChannelCategory{}, err
	}

	cat, err := s.categories.UpdateCategory(ctx, group.ID, catPubID, trimmedPtr, isPrivate, s.now())
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return models.PublicChannelCategory{}, exception.NotFound(exception.MsgInvalidInput, map[string]any{"field": "category_id"})
		}

		return models.PublicChannelCategory{}, err
	}

	s.recordGroupActivity(ctx, models.GroupActivityLogEntry{
		GroupID:    group.ID,
		ActorID:    userID,
		Category:   "channel",
		Action:     "category.update",
		TargetType: "category",
		TargetID:   cat.PublicID,
		Metadata:   map[string]any{"name": cat.Name, "is_private": cat.IsPrivate},
	})

	return cat.Public(), nil
}

func (s service) DeleteGroupCategory(
	ctx context.Context,
	accessToken, groupPublicID, categoryPublicID string,
) error {
	catPubID := strings.TrimSpace(categoryPublicID)
	if catPubID == "" {
		return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "category_id"})
	}

	userID, group, err := s.resolveGroupWithManageChannels(ctx, accessToken, groupPublicID)
	if err != nil {
		return err
	}

	if err := s.categories.DeleteCategory(ctx, group.ID, catPubID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return exception.NotFound(exception.MsgInvalidInput, map[string]any{"field": "category_id"})
		}

		return err
	}

	s.recordGroupActivity(ctx, models.GroupActivityLogEntry{
		GroupID:    group.ID,
		ActorID:    userID,
		Category:   "channel",
		Action:     "category.delete",
		TargetType: "category",
		TargetID:   catPubID,
	})

	return nil
}

func (s service) ListCategoryPermissions(
	ctx context.Context,
	accessToken, groupPublicID, categoryPublicID string,
) ([]models.PublicChannelCategoryPermissionOverride, error) {
	catPubID := strings.TrimSpace(categoryPublicID)
	if catPubID == "" {
		return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "category_id"})
	}

	_, group, err := s.resolveGroupWithManageChannels(ctx, accessToken, groupPublicID)
	if err != nil {
		return nil, err
	}

	overrides, err := s.categories.ListCategoryPermissionOverrides(ctx, group.ID, catPubID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, exception.NotFound(exception.MsgInvalidInput, map[string]any{"field": "category_id"})
		}

		return nil, err
	}

	return overrides, nil
}

func sanitizeCategoryPermList(perms []string) ([]string, error) {
	out := make([]string, 0, len(perms))
	seen := make(map[string]struct{}, len(perms))
	for _, raw := range perms {
		p := strings.TrimSpace(raw)
		if p == "" {
			continue
		}
		if !slices.Contains(constants.AllKnownPermissions, p) {
			return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "permissions"})
		}
		if _, dup := seen[p]; dup {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}

	return out, nil
}

func (s service) UpsertCategoryPermission(
	ctx context.Context,
	accessToken, groupPublicID, categoryPublicID, targetType, targetPublicID string,
	allow, deny []string,
) (models.PublicChannelCategoryPermissionOverride, error) {
	catPubID := strings.TrimSpace(categoryPublicID)
	tType := strings.TrimSpace(targetType)
	tID := strings.TrimSpace(targetPublicID)
	if catPubID == "" || tID == "" || (tType != "role" && tType != "member") {
		return models.PublicChannelCategoryPermissionOverride{}, exception.InvalidInput(exception.MsgInvalidInput, nil)
	}

	cleanAllow, err := sanitizeCategoryPermList(allow)
	if err != nil {
		return models.PublicChannelCategoryPermissionOverride{}, err
	}
	cleanDeny, err := sanitizeCategoryPermList(deny)
	if err != nil {
		return models.PublicChannelCategoryPermissionOverride{}, err
	}
	for _, p := range cleanAllow {
		if slices.Contains(cleanDeny, p) {
			return models.PublicChannelCategoryPermissionOverride{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "permissions"})
		}
	}

	userID, group, err := s.resolveGroupWithManageChannels(ctx, accessToken, groupPublicID)
	if err != nil {
		return models.PublicChannelCategoryPermissionOverride{}, err
	}

	override, err := s.categories.UpsertCategoryPermissionOverride(ctx, group.ID, catPubID, tType, tID, cleanAllow, cleanDeny, s.now())
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return models.PublicChannelCategoryPermissionOverride{}, exception.NotFound(exception.MsgInvalidInput, nil)
		}

		return models.PublicChannelCategoryPermissionOverride{}, err
	}

	s.recordGroupActivity(ctx, models.GroupActivityLogEntry{
		GroupID:    group.ID,
		ActorID:    userID,
		Category:   "channel",
		Action:     "category.permissions.update",
		TargetType: "category",
		TargetID:   catPubID,
		Metadata:   map[string]any{"target_type": tType, "target_id": tID},
	})

	return override, nil
}

func (s service) DeleteCategoryPermission(
	ctx context.Context,
	accessToken, groupPublicID, categoryPublicID, targetType, targetPublicID string,
) error {
	catPubID := strings.TrimSpace(categoryPublicID)
	tType := strings.TrimSpace(targetType)
	tID := strings.TrimSpace(targetPublicID)
	if catPubID == "" || tID == "" || (tType != "role" && tType != "member") {
		return exception.InvalidInput(exception.MsgInvalidInput, nil)
	}

	userID, group, err := s.resolveGroupWithManageChannels(ctx, accessToken, groupPublicID)
	if err != nil {
		return err
	}

	if err := s.categories.DeleteCategoryPermissionOverride(ctx, group.ID, catPubID, tType, tID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return exception.NotFound(exception.MsgInvalidInput, nil)
		}

		return err
	}

	s.recordGroupActivity(ctx, models.GroupActivityLogEntry{
		GroupID:    group.ID,
		ActorID:    userID,
		Category:   "channel",
		Action:     "category.permissions.delete",
		TargetType: "category",
		TargetID:   catPubID,
		Metadata:   map[string]any{"target_type": tType, "target_id": tID},
	})

	return nil
}

func (s service) ReorderChannelCategories(
	ctx context.Context,
	accessToken, groupPublicID string,
	categoryPublicIDs []string,
) error {
	_, group, err := s.resolveGroupWithManageChannels(ctx, accessToken, groupPublicID)
	if err != nil {
		return err
	}

	categories, err := s.categories.ListCategoriesByGroupID(ctx, group.ID)
	if err != nil {
		return err
	}
	if len(categoryPublicIDs) != len(categories) {
		return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "category_ids"})
	}

	byPublic := make(map[string]string, len(categories))
	for _, cat := range categories {
		byPublic[cat.PublicID] = cat.ID
	}
	orderedUUIDs, err := mapOrderedCategoryPublicIDs(categoryPublicIDs, byPublic, "category_ids")
	if err != nil {
		return err
	}

	return s.categories.SetCategoryPositions(ctx, group.ID, orderedUUIDs)
}

func (s service) resolveGroupWithManageChannels(
	ctx context.Context,
	accessToken, groupPublicID string,
) (string, models.Group, error) {
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
		if errors.Is(err, constants.ErrNotFound) {
			return "", models.Group{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_id"})
		}

		return "", models.Group{}, err
	}

	member, err := s.members.FindGroupMember(ctx, group.ID, userID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return "", models.Group{}, exception.PermissionDenied(exception.MsgForbidden, nil)
		}

		return "", models.Group{}, err
	}

	if member.Role != constants.GroupRoleOwner {
		if s.roles == nil {
			return "", models.Group{}, exception.PermissionDenied(exception.MsgForbidden, nil)
		}
		perms, err := s.roles.GetEffectivePermissions(ctx, group.ID, userID)
		if err != nil {
			return "", models.Group{}, err
		}
		if !slices.Contains(perms, constants.PermManageChannels) {
			return "", models.Group{}, exception.PermissionDenied(exception.MsgForbidden, nil)
		}
	}

	return userID, group, nil
}

func mapOrderedCategoryPublicIDs(rawIDs []string, byPublic map[string]string, field string) ([]string, error) {
	seen := make(map[string]struct{}, len(rawIDs))
	orderedUUIDs := make([]string, 0, len(rawIDs))
	for _, raw := range rawIDs {
		publicID, err := validator.ChannelPublicID(raw)
		if err != nil {
			return nil, err
		}
		if _, dup := seen[publicID]; dup {
			return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": field})
		}
		uuid, ok := byPublic[publicID]
		if !ok {
			return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": field})
		}
		seen[publicID] = struct{}{}
		orderedUUIDs = append(orderedUUIDs, uuid)
	}

	return orderedUUIDs, nil
}
