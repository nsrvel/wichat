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
	channelrepo "github.com/wichat/wichat/backend/ms-user/internal/channel/repository"
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

// Service is channel metadata for a group workspace.
type Service interface {
	ListGroupChannels(ctx context.Context, accessToken, groupPublicID string) ([]models.PublicChannel, []models.PublicChannelCategory, error)
	CreateGroupChannel(ctx context.Context, accessToken, groupPublicID, name, channelType, categoryPublicID string, isPrivate bool) (models.PublicChannel, error)
	UpdateGroupChannel(ctx context.Context, accessToken, groupPublicID, channelPublicID string, input models.UpdateChannelInput) (models.PublicChannel, error)
	DeleteGroupChannel(ctx context.Context, accessToken, groupPublicID, channelPublicID string) error
	ListChannelPermissions(ctx context.Context, accessToken, groupPublicID, channelPublicID string) ([]models.PublicChannelCategoryPermissionOverride, error)
	UpsertChannelPermission(ctx context.Context, accessToken, groupPublicID, channelPublicID, targetType, targetPublicID string, allow, deny []string) (models.PublicChannelCategoryPermissionOverride, error)
	DeleteChannelPermission(ctx context.Context, accessToken, groupPublicID, channelPublicID, targetType, targetPublicID string) error
	SyncChannelPermissions(ctx context.Context, accessToken, groupPublicID, channelPublicID string) (models.PublicChannel, error)
	ReorderGroupChannels(ctx context.Context, accessToken, groupPublicID, sidebarSection, categoryPublicID string, channelPublicIDs []string) error
	AssertGroupChannelAccess(ctx context.Context, accessToken, groupPublicID, channelPublicID string) (models.ChannelAccess, error)
	SeedDefaultChannels(ctx context.Context, groupID, ownerUserID string, createdAt time.Time) error
	SeedDefaultChannelsWithTemplate(ctx context.Context, groupID, ownerUserID, template string, createdAt time.Time) error
	WithGroupActivityLog(r groupactivitysvc.Recorder) Service
}

type service struct {
	channels      channelrepo.PgRepository
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

// NewService builds channel use cases.
func NewService(
	channels channelrepo.PgRepository,
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
		channels:   channels,
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

type permOverride struct {
	TargetType       string
	TargetID         string
	IsEveryone       bool
	AllowPermissions []string
	DenyPermissions  []string
}

func categoryOverridesToRules(overrides []models.ChannelCategoryPermissionOverride) []permOverride {
	out := make([]permOverride, len(overrides))
	for i, o := range overrides {
		out[i] = permOverride{
			TargetType:       o.TargetType,
			TargetID:         o.TargetID,
			IsEveryone:       o.IsEveryone,
			AllowPermissions: o.AllowPermissions,
			DenyPermissions:  o.DenyPermissions,
		}
	}

	return out
}

func channelOverridesToRules(overrides []models.ChannelPermissionOverride) []permOverride {
	out := make([]permOverride, len(overrides))
	for i, o := range overrides {
		out[i] = permOverride{
			TargetType:       o.TargetType,
			TargetID:         o.TargetID,
			IsEveryone:       o.IsEveryone,
			AllowPermissions: o.AllowPermissions,
			DenyPermissions:  o.DenyPermissions,
		}
	}

	return out
}

func resolveEffectivePermissionsWithOverrides(
	basePerms []string,
	isPrivate bool,
	overrides []permOverride,
	memberRoleIDs []string,
	userID string,
) []string {
	permSet := make(map[string]struct{}, len(basePerms))
	for _, p := range basePerms {
		permSet[p] = struct{}{}
	}
	if isPrivate {
		delete(permSet, constants.PermViewChannels)
	}

	// 1. @everyone override
	for _, o := range overrides {
		if o.TargetType == "role" && o.IsEveryone {
			for _, d := range o.DenyPermissions {
				delete(permSet, d)
			}
			for _, a := range o.AllowPermissions {
				permSet[a] = struct{}{}
			}
		}
	}

	// 2. Assigned custom roles (union of denies, then union of allows)
	roleDenies := make(map[string]struct{})
	roleAllows := make(map[string]struct{})
	for _, o := range overrides {
		if o.TargetType == "role" && !o.IsEveryone && slices.Contains(memberRoleIDs, o.TargetID) {
			for _, d := range o.DenyPermissions {
				roleDenies[d] = struct{}{}
			}
			for _, a := range o.AllowPermissions {
				roleAllows[a] = struct{}{}
			}
		}
	}
	for d := range roleDenies {
		delete(permSet, d)
	}
	for a := range roleAllows {
		permSet[a] = struct{}{}
	}

	// 3. Specific member override (highest precedence)
	for _, o := range overrides {
		if o.TargetType == "member" && o.TargetID == userID {
			for _, d := range o.DenyPermissions {
				delete(permSet, d)
			}
			for _, a := range o.AllowPermissions {
				permSet[a] = struct{}{}
			}
		}
	}

	out := make([]string, 0, len(permSet))
	for _, p := range constants.AllKnownPermissions {
		if _, ok := permSet[p]; ok {
			out = append(out, p)
		}
	}
	return out
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

	member, err := s.members.FindGroupMember(ctx, group.ID, userID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, nil, exception.PermissionDenied(exception.MsgForbidden, nil)
		}

		return nil, nil, err
	}

	channels, err := s.channels.ListByGroupID(ctx, group.ID)
	if err != nil {
		return nil, nil, err
	}

	categories, err := s.categories.ListCategoriesByGroupID(ctx, group.ID)
	if err != nil {
		return nil, nil, err
	}

	bypass := member.Role == constants.GroupRoleOwner
	var basePerms []string
	var roleIDs []string
	overridesByCat := make(map[string][]models.ChannelCategoryPermissionOverride)
	overridesByCh := make(map[string][]models.ChannelPermissionOverride)
	if !bypass && s.roles != nil {
		basePerms, err = s.roles.GetEffectivePermissions(ctx, group.ID, userID)
		if err != nil {
			return nil, nil, err
		}
		if slices.Contains(basePerms, constants.PermAdministrator) {
			bypass = true
		} else {
			roleIDs, err = s.roles.ListUserRoleIDs(ctx, group.ID, userID)
			if err != nil {
				return nil, nil, err
			}
			allCatOverrides, err := s.categories.ListGroupCategoryPermissionOverrides(ctx, group.ID)
			if err != nil {
				return nil, nil, err
			}
			for _, o := range allCatOverrides {
				overridesByCat[o.CategoryID] = append(overridesByCat[o.CategoryID], o)
			}
			allChOverrides, err := s.channels.ListGroupChannelPermissionOverrides(ctx, group.ID)
			if err != nil {
				return nil, nil, err
			}
			for _, o := range allChOverrides {
				overridesByCh[o.ChannelID] = append(overridesByCh[o.ChannelID], o)
			}
		}
	}

	visibleCats := make(map[string]bool, len(categories))
	catEffectivePerms := make(map[string][]string, len(categories))
	outCategories := make([]models.PublicChannelCategory, 0, len(categories))
	for _, cat := range categories {
		if bypass || s.roles == nil {
			visibleCats[cat.ID] = true
			outCategories = append(outCategories, cat.Public())
			continue
		}
		effective := resolveEffectivePermissionsWithOverrides(
			basePerms,
			cat.IsPrivate,
			categoryOverridesToRules(overridesByCat[cat.ID]),
			roleIDs,
			userID,
		)
		catEffectivePerms[cat.ID] = effective
		if slices.Contains(effective, constants.PermViewChannels) {
			visibleCats[cat.ID] = true
			outCategories = append(outCategories, cat.Public())
		}
	}

	outChannels := make([]models.PublicChannel, 0, len(channels))
	for _, ch := range channels {
		if ch.CategoryID != "" && !visibleCats[ch.CategoryID] {
			continue
		}
		pub := ch.Public()
		if bypass || s.roles == nil {
			pub.CallerPermissions = append([]string(nil), constants.AllKnownPermissions...)
			outChannels = append(outChannels, pub)
			continue
		}
		startPerms := basePerms
		if ch.CategoryID != "" {
			startPerms = catEffectivePerms[ch.CategoryID]
		}
		effective := resolveEffectivePermissionsWithOverrides(
			startPerms,
			ch.IsPrivate,
			channelOverridesToRules(overridesByCh[ch.ID]),
			roleIDs,
			userID,
		)
		if !slices.Contains(effective, constants.PermViewChannels) {
			continue
		}
		pub.CallerPermissions = effective
		outChannels = append(outChannels, pub)
	}

	return outChannels, outCategories, nil
}

func validateUpdateChannelInput(input models.UpdateChannelInput) (models.UpdateChannelInput, error) {
	out := input
	hasAny := false
	if input.Name != nil {
		trimmed := strings.TrimSpace(*input.Name)
		if trimmed == "" || utf8.RuneCountInString(trimmed) > 64 {
			return models.UpdateChannelInput{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "name"})
		}
		out.Name = &trimmed
		hasAny = true
	}
	if input.IsPrivate != nil {
		hasAny = true
	}
	if input.SlowmodeSeconds != nil {
		if *input.SlowmodeSeconds < 0 || *input.SlowmodeSeconds > 21600 {
			return models.UpdateChannelInput{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "slowmode_seconds"})
		}
		hasAny = true
	}
	if input.ThreadSlowmodeSeconds != nil {
		if *input.ThreadSlowmodeSeconds < 0 || *input.ThreadSlowmodeSeconds > 21600 {
			return models.UpdateChannelInput{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "thread_slowmode_seconds"})
		}
		hasAny = true
	}
	if input.ContentVisibility != nil {
		v := strings.TrimSpace(*input.ContentVisibility)
		if v != "default" && v != "spoiler" && v != "age_restricted" {
			return models.UpdateChannelInput{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "content_visibility"})
		}
		out.ContentVisibility = &v
		hasAny = true
	}
	if input.ForumDefaultLayout != nil {
		v := strings.TrimSpace(*input.ForumDefaultLayout)
		if v != "list" && v != "gallery" {
			return models.UpdateChannelInput{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "forum_default_layout"})
		}
		out.ForumDefaultLayout = &v
		hasAny = true
	}
	if input.ForumSortOrder != nil {
		v := strings.TrimSpace(*input.ForumSortOrder)
		if v != "recent_activity" && v != "creation_time" {
			return models.UpdateChannelInput{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "forum_sort_order"})
		}
		out.ForumSortOrder = &v
		hasAny = true
	}
	if input.ForumTagMatching != nil {
		v := strings.TrimSpace(*input.ForumTagMatching)
		if v != "match_some" && v != "match_all" {
			return models.UpdateChannelInput{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "forum_tag_matching"})
		}
		out.ForumTagMatching = &v
		hasAny = true
	}
	if input.BitrateKbps != nil {
		if *input.BitrateKbps < 8 || *input.BitrateKbps > 96 {
			return models.UpdateChannelInput{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "bitrate_kbps"})
		}
		hasAny = true
	}
	if input.VideoQuality != nil {
		v := strings.TrimSpace(*input.VideoQuality)
		if v != "auto" && v != "720p" {
			return models.UpdateChannelInput{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "video_quality"})
		}
		out.VideoQuality = &v
		hasAny = true
	}
	if input.UserLimit != nil {
		if *input.UserLimit < 0 || *input.UserLimit > 99 {
			return models.UpdateChannelInput{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "user_limit"})
		}
		hasAny = true
	}
	if input.RTCRegion != nil {
		v := strings.TrimSpace(*input.RTCRegion)
		if v == "" || len(v) > 64 {
			return models.UpdateChannelInput{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "rtc_region"})
		}
		out.RTCRegion = &v
		hasAny = true
	}
	if !hasAny {
		return models.UpdateChannelInput{}, exception.InvalidInput(exception.MsgInvalidInput, nil)
	}

	return out, nil
}

func (s service) UpdateGroupChannel(
	ctx context.Context,
	accessToken, groupPublicID, channelPublicID string,
	input models.UpdateChannelInput,
) (models.PublicChannel, error) {
	chPubID := strings.TrimSpace(channelPublicID)
	if chPubID == "" {
		return models.PublicChannel{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "channel_id"})
	}

	cleanInput, err := validateUpdateChannelInput(input)
	if err != nil {
		return models.PublicChannel{}, err
	}

	userID, group, err := s.resolveGroupWithManageChannels(ctx, accessToken, groupPublicID)
	if err != nil {
		return models.PublicChannel{}, err
	}

	ch, err := s.channels.UpdateChannel(ctx, group.ID, chPubID, cleanInput, s.now())
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return models.PublicChannel{}, exception.NotFound(exception.MsgInvalidInput, map[string]any{"field": "channel_id"})
		}
		return models.PublicChannel{}, err
	}

	s.recordGroupActivity(ctx, models.GroupActivityLogEntry{
		GroupID:    group.ID,
		ActorID:    userID,
		Category:   "channel",
		Action:     "channel.update",
		TargetType: "channel",
		TargetID:   ch.PublicID,
		Metadata:   map[string]any{"name": ch.Name, "is_private": ch.IsPrivate},
	})

	return ch.Public(), nil
}

func (s service) DeleteGroupChannel(
	ctx context.Context,
	accessToken, groupPublicID, channelPublicID string,
) error {
	chPubID := strings.TrimSpace(channelPublicID)
	if chPubID == "" {
		return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "channel_id"})
	}

	userID, group, err := s.resolveGroupWithManageChannels(ctx, accessToken, groupPublicID)
	if err != nil {
		return err
	}

	if err := s.channels.DeleteChannel(ctx, group.ID, chPubID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return exception.NotFound(exception.MsgInvalidInput, map[string]any{"field": "channel_id"})
		}
		return err
	}

	s.recordGroupActivity(ctx, models.GroupActivityLogEntry{
		GroupID:    group.ID,
		ActorID:    userID,
		Category:   "channel",
		Action:     "channel.delete",
		TargetType: "channel",
		TargetID:   chPubID,
	})

	return nil
}

func (s service) ListChannelPermissions(
	ctx context.Context,
	accessToken, groupPublicID, channelPublicID string,
) ([]models.PublicChannelCategoryPermissionOverride, error) {
	chPubID := strings.TrimSpace(channelPublicID)
	if chPubID == "" {
		return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "channel_id"})
	}

	_, group, err := s.resolveGroupWithManageChannels(ctx, accessToken, groupPublicID)
	if err != nil {
		return nil, err
	}

	overrides, err := s.channels.ListChannelPermissionOverrides(ctx, group.ID, chPubID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, exception.NotFound(exception.MsgInvalidInput, map[string]any{"field": "channel_id"})
		}
		return nil, err
	}

	return overrides, nil
}

func sanitizeChannelPermList(perms []string) ([]string, error) {
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

func (s service) UpsertChannelPermission(
	ctx context.Context,
	accessToken, groupPublicID, channelPublicID, targetType, targetPublicID string,
	allow, deny []string,
) (models.PublicChannelCategoryPermissionOverride, error) {
	chPubID := strings.TrimSpace(channelPublicID)
	tType := strings.TrimSpace(targetType)
	tID := strings.TrimSpace(targetPublicID)
	if chPubID == "" || tID == "" || (tType != "role" && tType != "member") {
		return models.PublicChannelCategoryPermissionOverride{}, exception.InvalidInput(exception.MsgInvalidInput, nil)
	}

	cleanAllow, err := sanitizeChannelPermList(allow)
	if err != nil {
		return models.PublicChannelCategoryPermissionOverride{}, err
	}
	cleanDeny, err := sanitizeChannelPermList(deny)
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

	override, err := s.channels.UpsertChannelPermissionOverride(ctx, group.ID, chPubID, tType, tID, cleanAllow, cleanDeny, s.now())
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
		Action:     "channel.permissions.update",
		TargetType: "channel",
		TargetID:   chPubID,
		Metadata:   map[string]any{"target_type": tType, "target_id": tID},
	})

	return override, nil
}

func (s service) DeleteChannelPermission(
	ctx context.Context,
	accessToken, groupPublicID, channelPublicID, targetType, targetPublicID string,
) error {
	chPubID := strings.TrimSpace(channelPublicID)
	tType := strings.TrimSpace(targetType)
	tID := strings.TrimSpace(targetPublicID)
	if chPubID == "" || tID == "" || (tType != "role" && tType != "member") {
		return exception.InvalidInput(exception.MsgInvalidInput, nil)
	}

	userID, group, err := s.resolveGroupWithManageChannels(ctx, accessToken, groupPublicID)
	if err != nil {
		return err
	}

	if err := s.channels.DeleteChannelPermissionOverride(ctx, group.ID, chPubID, tType, tID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return exception.NotFound(exception.MsgInvalidInput, nil)
		}
		return err
	}

	s.recordGroupActivity(ctx, models.GroupActivityLogEntry{
		GroupID:    group.ID,
		ActorID:    userID,
		Category:   "channel",
		Action:     "channel.permissions.delete",
		TargetType: "channel",
		TargetID:   chPubID,
		Metadata:   map[string]any{"target_type": tType, "target_id": tID},
	})

	return nil
}

func (s service) SyncChannelPermissions(
	ctx context.Context,
	accessToken, groupPublicID, channelPublicID string,
) (models.PublicChannel, error) {
	chPubID := strings.TrimSpace(channelPublicID)
	if chPubID == "" {
		return models.PublicChannel{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "channel_id"})
	}

	userID, group, err := s.resolveGroupWithManageChannels(ctx, accessToken, groupPublicID)
	if err != nil {
		return models.PublicChannel{}, err
	}

	ch, err := s.channels.SyncChannelPermissionsWithCategory(ctx, group.ID, chPubID, s.now())
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return models.PublicChannel{}, exception.NotFound(exception.MsgInvalidInput, map[string]any{"field": "channel_id"})
		}
		return models.PublicChannel{}, err
	}

	s.recordGroupActivity(ctx, models.GroupActivityLogEntry{
		GroupID:    group.ID,
		ActorID:    userID,
		Category:   "channel",
		Action:     "channel.permissions.sync",
		TargetType: "channel",
		TargetID:   ch.PublicID,
		Metadata:   map[string]any{"category_id": ch.CategoryPublicID},
	})

	return ch.Public(), nil
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
	categories, err := s.categories.ListCategoriesByGroupID(ctx, groupID)
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

	member, err := s.members.FindGroupMember(ctx, group.ID, userID)
	if err != nil {
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
	if member.Role == constants.GroupRoleOwner {
		perms = append([]string(nil), constants.AllKnownPermissions...)
	} else if s.roles != nil {
		p, err := s.roles.GetEffectivePermissions(ctx, group.ID, userID)
		if err != nil {
			return models.ChannelAccess{}, err
		}
		perms = p

		if !slices.Contains(perms, constants.PermAdministrator) {
			roleIDs, err := s.roles.ListUserRoleIDs(ctx, group.ID, userID)
			if err != nil {
				return models.ChannelAccess{}, err
			}
			if ch.CategoryID != "" {
				categories, err := s.categories.ListCategoriesByGroupID(ctx, group.ID)
				if err != nil {
					return models.ChannelAccess{}, err
				}
				var cat *models.ChannelCategory
				for i := range categories {
					if categories[i].ID == ch.CategoryID {
						cat = &categories[i]
						break
					}
				}
				if cat != nil {
					allOverrides, err := s.categories.ListGroupCategoryPermissionOverrides(ctx, group.ID)
					if err != nil {
						return models.ChannelAccess{}, err
					}
					var catOverrides []models.ChannelCategoryPermissionOverride
					for _, o := range allOverrides {
						if o.CategoryID == cat.ID {
							catOverrides = append(catOverrides, o)
						}
					}
					perms = resolveEffectivePermissionsWithOverrides(
						perms,
						cat.IsPrivate,
						categoryOverridesToRules(catOverrides),
						roleIDs,
						userID,
					)
					if !slices.Contains(perms, constants.PermViewChannels) {
						return models.ChannelAccess{}, exception.PermissionDenied(exception.MsgForbidden, nil)
					}
				}
			}

			allChOverrides, err := s.channels.ListGroupChannelPermissionOverrides(ctx, group.ID)
			if err != nil {
				return models.ChannelAccess{}, err
			}
			var chOverrides []models.ChannelPermissionOverride
			for _, o := range allChOverrides {
				if o.ChannelID == ch.ID {
					chOverrides = append(chOverrides, o)
				}
			}
			perms = resolveEffectivePermissionsWithOverrides(
				perms,
				ch.IsPrivate,
				channelOverridesToRules(chOverrides),
				roleIDs,
				userID,
			)
			if !slices.Contains(perms, constants.PermViewChannels) {
				return models.ChannelAccess{}, exception.PermissionDenied(exception.MsgForbidden, nil)
			}
		}
	}

	pub := ch.Public()
	return models.ChannelAccess{
		ChannelID:         ch.ID,
		UserID:            userID,
		Permissions:       perms,
		ChannelType:       pub.Type,
		SlowmodeSeconds:   pub.SlowmodeSeconds,
		ContentVisibility: pub.ContentVisibility,
	}, nil
}

func (s service) SeedDefaultChannels(ctx context.Context, groupID, ownerUserID string, createdAt time.Time) error {
	return s.channels.SeedDefaultChannels(ctx, groupID, ownerUserID, createdAt)
}

func (s service) SeedDefaultChannelsWithTemplate(ctx context.Context, groupID, ownerUserID, template string, createdAt time.Time) error {
	return s.channels.SeedDefaultChannelsWithTemplate(ctx, groupID, ownerUserID, template, createdAt)
}
