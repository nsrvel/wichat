package service

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	authrepo "github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	channelrepo "github.com/wichat/wichat/backend/ms-user/internal/channel/repository"
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

// Service is channel metadata for a group workspace.
type Service interface {
	ListGroupChannels(ctx context.Context, accessToken, groupPublicID string) ([]models.PublicChannel, []models.PublicChannelCategory, error)
	CreateGroupCategory(ctx context.Context, accessToken, groupPublicID, name string, isPrivate bool) (models.PublicChannelCategory, error)
	CreateGroupChannel(ctx context.Context, accessToken, groupPublicID, name, channelType, categoryPublicID string, isPrivate bool) (models.PublicChannel, error)
	ReorderChannelCategories(ctx context.Context, accessToken, groupPublicID string, categoryPublicIDs []string) error
	ReorderGroupChannels(ctx context.Context, accessToken, groupPublicID, sidebarSection, categoryPublicID string, channelPublicIDs []string) error
	AssertGroupChannelAccess(ctx context.Context, accessToken, groupPublicID, channelPublicID string) (models.ChannelAccess, error)
	SeedDefaultChannels(ctx context.Context, groupID, ownerUserID string, createdAt time.Time) error
}

type service struct {
	channels channelrepo.PgRepository
	groups   grouprepo.PgRepository
	members  groupmemberrepo.PgRepository
	roles    grouprolerepo.PgRepository
	sessions authrepo.SessionRepository
	secret   []byte
	log      logger.Logger
	now      func() time.Time
}

// NewService builds channel use cases.
func NewService(
	channels channelrepo.PgRepository,
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
		channels: channels,
		groups:   groups,
		members:  members,
		roles:    roles,
		sessions: sessions,
		secret:   copied,
		log:      log,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

func (s service) ListGroupChannels(ctx context.Context, accessToken, groupPublicID string) ([]models.PublicChannel, []models.PublicChannelCategory, error) {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return nil, nil, err
	}

	publicID, err := validator.GroupPublicID(groupPublicID)
	if err != nil {
		return nil, nil, err
	}

	group, err := s.groups.FindGroupByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_id"})
		}

		return nil, nil, err
	}

	if _, err := s.members.FindGroupMember(ctx, group.ID, userID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, nil, exception.PermissionDenied(exception.MsgForbidden, nil)
		}

		return nil, nil, err
	}

	channels, err := s.channels.ListByGroupID(ctx, group.ID)
	if err != nil {
		return nil, nil, err
	}

	categories, err := s.channels.ListCategoriesByGroupID(ctx, group.ID)
	if err != nil {
		return nil, nil, err
	}

	outChannels := make([]models.PublicChannel, len(channels))
	for i, ch := range channels {
		outChannels[i] = ch.Public()
	}

	outCategories := make([]models.PublicChannelCategory, len(categories))
	for i, cat := range categories {
		outCategories[i] = cat.Public()
	}

	return outChannels, outCategories, nil
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

	cat, err := s.channels.CreateCategory(ctx, group.ID, trimmed, isPrivate, userID, s.now())
	if err != nil {
		return models.PublicChannelCategory{}, err
	}

	return cat.Public(), nil
}

func (s service) CreateGroupChannel(
	ctx context.Context,
	accessToken, groupPublicID, name, channelType, categoryPublicID string,
	isPrivate bool,
) (models.PublicChannel, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" || utf8.RuneCountInString(trimmed) > 64 {
		return models.PublicChannel{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "name"})
	}

	kind := strings.TrimSpace(channelType)
	switch kind {
	case constants.ChannelKindTextOnly, constants.ChannelKindVoiceOnly, constants.ChannelKindForum, constants.ChannelKindBroadcast:
	default:
		return models.PublicChannel{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "type"})
	}

	userID, group, err := s.resolveGroupWithManageChannels(ctx, accessToken, groupPublicID)
	if err != nil {
		return models.PublicChannel{}, err
	}

	var categoryUUID string
	if trimmedCat := strings.TrimSpace(categoryPublicID); trimmedCat != "" {
		resolved, err := s.resolveCategoryUUID(ctx, group.ID, trimmedCat)
		if err != nil {
			return models.PublicChannel{}, err
		}
		categoryUUID = resolved
	}

	ch, err := s.channels.CreateChannel(ctx, group.ID, categoryUUID, kind, trimmed, isPrivate, userID, s.now())
	if err != nil {
		return models.PublicChannel{}, err
	}

	return ch.Public(), nil
}

func (s service) resolveCategoryUUID(ctx context.Context, groupID, categoryPublicID string) (string, error) {
	validCatID, err := validator.ChannelPublicID(categoryPublicID)
	if err != nil {
		return "", err
	}
	categories, err := s.channels.ListCategoriesByGroupID(ctx, groupID)
	if err != nil {
		return "", err
	}
	for _, cat := range categories {
		if cat.PublicID == validCatID {
			return cat.ID, nil
		}
	}

	return "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "category_id"})
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

	categories, err := s.channels.ListCategoriesByGroupID(ctx, group.ID)
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
	orderedUUIDs, err := mapOrderedPublicIDs(categoryPublicIDs, byPublic, "category_ids")
	if err != nil {
		return err
	}

	return s.channels.SetCategoryPositions(ctx, group.ID, orderedUUIDs)
}

func (s service) ReorderGroupChannels(
	ctx context.Context,
	accessToken, groupPublicID, sidebarSection, categoryPublicID string,
	channelPublicIDs []string,
) error {
	_, group, err := s.resolveGroupWithManageChannels(ctx, accessToken, groupPublicID)
	if err != nil {
		return err
	}
	if len(channelPublicIDs) == 0 {
		return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "channel_ids"})
	}

	var targetSection *string
	if trimmedSec := strings.TrimSpace(sidebarSection); trimmedSec != "" {
		if trimmedSec != constants.SidebarSectionFavorites && trimmedSec != constants.SidebarSectionChannels {
			return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "sidebar_section"})
		}
		targetSection = &trimmedSec
	}

	var targetCategoryID *string
	if trimmedCat := strings.TrimSpace(categoryPublicID); trimmedCat != "" {
		id, err := s.resolveCategoryUUID(ctx, group.ID, trimmedCat)
		if err != nil {
			return err
		}
		targetCategoryID = &id
		if targetSection == nil {
			sec := constants.SidebarSectionChannels
			targetSection = &sec
		}
	}

	channels, err := s.channels.ListByGroupID(ctx, group.ID)
	if err != nil {
		return err
	}

	byPublic := make(map[string]string, len(channels))
	for _, ch := range channels {
		byPublic[ch.PublicID] = ch.ID
	}
	orderedUUIDs, err := mapOrderedPublicIDs(channelPublicIDs, byPublic, "channel_ids")
	if err != nil {
		return err
	}

	return s.channels.SetChannelPositions(ctx, group.ID, targetSection, targetCategoryID, orderedUUIDs)
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

func mapOrderedPublicIDs(rawIDs []string, byPublic map[string]string, field string) ([]string, error) {
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

func (s service) AssertGroupChannelAccess(ctx context.Context, accessToken, groupPublicID, channelPublicID string) (models.ChannelAccess, error) {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return models.ChannelAccess{}, err
	}

	groupID, err := validator.GroupPublicID(groupPublicID)
	if err != nil {
		return models.ChannelAccess{}, err
	}

	channelID, err := validator.ChannelPublicID(channelPublicID)
	if err != nil {
		return models.ChannelAccess{}, err
	}

	group, err := s.groups.FindGroupByPublicID(ctx, groupID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return models.ChannelAccess{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_id"})
		}

		return models.ChannelAccess{}, err
	}

	if _, err := s.members.FindGroupMember(ctx, group.ID, userID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return models.ChannelAccess{}, exception.PermissionDenied(exception.MsgForbidden, nil)
		}

		return models.ChannelAccess{}, err
	}

	ch, err := s.channels.FindByPublicIDInGroup(ctx, group.ID, channelID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return models.ChannelAccess{}, exception.PermissionDenied(exception.MsgForbidden, nil)
		}

		return models.ChannelAccess{}, err
	}

	var perms []string
	if s.roles != nil {
		p, err := s.roles.GetEffectivePermissions(ctx, group.ID, userID)
		if err == nil {
			perms = p
		}
	}

	return models.ChannelAccess{ChannelID: ch.ID, UserID: userID, Permissions: perms}, nil
}

func (s service) SeedDefaultChannels(ctx context.Context, groupID, ownerUserID string, createdAt time.Time) error {
	return s.channels.SeedDefaultChannels(ctx, groupID, ownerUserID, createdAt)
}
