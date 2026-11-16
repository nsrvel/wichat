package mock

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/wichat/wichat/backend/ms-user/internal/channel/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/ms-user/pkg/token"
)

type pgRepositoryMock struct {
	st *usermock.State
}

// NewPgRepositoryMock is an in-memory channel repo backed by shared test state.
func NewPgRepositoryMock(st *usermock.State) repository.PgRepository {
	if st == nil {
		st = usermock.NewState()
	}

	return pgRepositoryMock{st: st}
}

func (m pgRepositoryMock) ListByGroupID(_ context.Context, groupID string) ([]models.Channel, error) {
	m.st.Lock()
	defer m.st.Unlock()

	var out []models.Channel
	for _, ch := range m.st.Channels {
		if ch.GroupID == groupID {
			out = append(out, ch)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Position != out[j].Position {
			return out[i].Position < out[j].Position
		}

		return out[i].Name < out[j].Name
	})

	return out, nil
}

func (m pgRepositoryMock) CreateChannel(
	_ context.Context,
	groupID, categoryID, channelType, name string,
	isPrivate bool,
	createdBy string,
	now time.Time,
) (models.Channel, error) {
	m.st.Lock()
	defer m.st.Unlock()

	var categoryPublicID string
	if categoryID != "" {
		found := false
		for _, cat := range m.st.ChannelCategories {
			if cat.GroupID == groupID && cat.ID == categoryID {
				categoryPublicID = cat.PublicID
				found = true
				break
			}
		}
		if !found {
			return models.Channel{}, constants.ErrNotFound
		}
	}

	maxPos := int32(-1)
	for _, ch := range m.st.Channels {
		if ch.GroupID == groupID && ch.CategoryID == categoryID && ch.Position > maxPos {
			maxPos = ch.Position
		}
	}

	publicID, err := token.NewPublicID()
	if err != nil {
		return models.Channel{}, err
	}
	id, err := token.NewID()
	if err != nil {
		return models.Channel{}, err
	}

	ch := models.Channel{
		ID:               id,
		PublicID:         publicID,
		GroupID:          groupID,
		CategoryID:       categoryID,
		CategoryPublicID: categoryPublicID,
		Type:             channelType,
		IsPrivate:        isPrivate,
		Name:             name,
		SidebarSection:   constants.SidebarSectionChannels,
		Position:         maxPos + 1,
		CreatedBy:        createdBy,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	m.st.Channels = append(m.st.Channels, ch)

	return ch, nil
}

func indexPositions(orderedIDs []string) map[string]int32 {
	pos := make(map[string]int32, len(orderedIDs))
	var idx int32
	for _, id := range orderedIDs {
		pos[id] = idx
		idx++
	}

	return pos
}

func (m pgRepositoryMock) SetChannelPositions(_ context.Context, groupID string, sidebarSection, categoryID *string, orderedChannelIDs []string) error {
	m.st.Lock()
	defer m.st.Unlock()

	var categoryPublicID string
	if categoryID != nil {
		for _, cat := range m.st.ChannelCategories {
			if cat.GroupID == groupID && cat.ID == *categoryID {
				categoryPublicID = cat.PublicID
				break
			}
		}
	}

	pos := indexPositions(orderedChannelIDs)
	updated := 0
	for i := range m.st.Channels {
		if m.st.Channels[i].GroupID == groupID {
			if p, ok := pos[m.st.Channels[i].ID]; ok {
				m.st.Channels[i].Position = p
				if sidebarSection != nil {
					m.st.Channels[i].SidebarSection = *sidebarSection
					if *sidebarSection == constants.SidebarSectionFavorites {
						m.st.Channels[i].CategoryID = ""
						m.st.Channels[i].CategoryPublicID = ""
					}
				}
				if categoryID != nil {
					m.st.Channels[i].CategoryID = *categoryID
					m.st.Channels[i].CategoryPublicID = categoryPublicID
				}
				updated++
			}
		}
	}
	if updated != len(orderedChannelIDs) {
		return constants.ErrNotFound
	}

	return nil
}

func (m pgRepositoryMock) FindByPublicIDInGroup(_ context.Context, groupID, publicID string) (models.Channel, error) {
	m.st.Lock()
	defer m.st.Unlock()

	publicID = strings.TrimSpace(strings.ToLower(publicID))
	for _, ch := range m.st.Channels {
		if ch.GroupID == groupID && strings.EqualFold(ch.PublicID, publicID) {
			return ch, nil
		}
	}

	return models.Channel{}, constants.ErrNotFound
}

func (m pgRepositoryMock) FindByID(_ context.Context, id string) (models.Channel, error) {
	m.st.Lock()
	defer m.st.Unlock()

	id = strings.TrimSpace(id)
	for _, ch := range m.st.Channels {
		if ch.ID == id {
			return ch, nil
		}
	}

	return models.Channel{}, constants.ErrNotFound
}

func (m pgRepositoryMock) SeedDefaultChannels(ctx context.Context, groupID, createdBy string, createdAt time.Time) error {
	return m.SeedDefaultChannelsWithTemplate(ctx, groupID, createdBy, "", createdAt)
}

func (m pgRepositoryMock) SeedDefaultChannelsWithTemplate(
	_ context.Context,
	groupID, createdBy, template string,
	createdAt time.Time,
) error {
	m.st.Lock()
	defer m.st.Unlock()

	for _, ch := range m.st.Channels {
		if ch.GroupID == groupID {
			return nil
		}
	}

	if template == "" {
		publicID, err := token.NewPublicID()
		if err != nil {
			return err
		}
		id, err := token.NewID()
		if err != nil {
			return err
		}
		m.st.Channels = append(m.st.Channels, models.Channel{
			ID:             id,
			PublicID:       publicID,
			GroupID:        groupID,
			Type:           constants.ChannelKindTextOnly,
			Name:           constants.DefaultChannelName,
			SidebarSection: constants.SidebarSectionChannels,
			CreatedBy:      createdBy,
			CreatedAt:      createdAt,
			UpdatedAt:      createdAt,
		})
		return nil
	}

	specs := repository.TemplateCategories(template)
	var catPos int32
	for _, catSpec := range specs {
		catPubID, err := token.NewPublicID()
		if err != nil {
			return err
		}
		catID, err := token.NewID()
		if err != nil {
			return err
		}
		m.st.ChannelCategories = append(m.st.ChannelCategories, models.ChannelCategory{
			ID:        catID,
			PublicID:  catPubID,
			GroupID:   groupID,
			Name:      catSpec.Name,
			Position:  catPos,
			CreatedBy: createdBy,
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		})
		catPos++
		var chPos int32
		for _, chSpec := range catSpec.Channels {
			chPubID, err := token.NewPublicID()
			if err != nil {
				return err
			}
			chID, err := token.NewID()
			if err != nil {
				return err
			}
			m.st.Channels = append(m.st.Channels, models.Channel{
				ID:               chID,
				PublicID:         chPubID,
				GroupID:          groupID,
				CategoryID:       catID,
				CategoryPublicID: catPubID,
				Type:             chSpec.Type,
				Name:             chSpec.Name,
				SidebarSection:   constants.SidebarSectionChannels,
				Position:         chPos,
				CreatedBy:        createdBy,
				CreatedAt:        createdAt,
				UpdatedAt:        createdAt,
			})
			chPos++
		}
	}

	return nil
}

func (m pgRepositoryMock) UpdateChannel(
	_ context.Context,
	groupID, channelPublicID string,
	input models.UpdateChannelInput,
	now time.Time,
) (models.Channel, error) {
	m.st.Lock()
	defer m.st.Unlock()

	pubID := strings.TrimSpace(strings.ToLower(channelPublicID))
	for i, ch := range m.st.Channels {
		if ch.GroupID == groupID && strings.EqualFold(ch.PublicID, pubID) {
			if input.Name != nil {
				ch.Name = *input.Name
			}
			if input.IsPrivate != nil {
				ch.IsPrivate = *input.IsPrivate
			}
			if input.SlowmodeSeconds != nil {
				ch.SlowmodeSeconds = *input.SlowmodeSeconds
			}
			if input.ThreadSlowmodeSeconds != nil {
				ch.ThreadSlowmodeSeconds = *input.ThreadSlowmodeSeconds
			}
			if input.ContentVisibility != nil {
				ch.ContentVisibility = *input.ContentVisibility
			}
			if input.ForumDefaultLayout != nil {
				ch.ForumDefaultLayout = *input.ForumDefaultLayout
			}
			if input.ForumSortOrder != nil {
				ch.ForumSortOrder = *input.ForumSortOrder
			}
			if input.ForumTagMatching != nil {
				ch.ForumTagMatching = *input.ForumTagMatching
			}
			if input.BitrateKbps != nil {
				ch.BitrateKbps = *input.BitrateKbps
			}
			if input.VideoQuality != nil {
				ch.VideoQuality = *input.VideoQuality
			}
			if input.UserLimit != nil {
				ch.UserLimit = *input.UserLimit
			}
			if input.RTCRegion != nil {
				ch.RTCRegion = *input.RTCRegion
			}
			ch.UpdatedAt = now
			m.st.Channels[i] = ch
			return ch, nil
		}
	}

	return models.Channel{}, constants.ErrNotFound
}

func (m pgRepositoryMock) DeleteChannel(_ context.Context, groupID, channelPublicID string) error {
	m.st.Lock()
	defer m.st.Unlock()

	pubID := strings.TrimSpace(strings.ToLower(channelPublicID))
	for i, ch := range m.st.Channels {
		if ch.GroupID == groupID && strings.EqualFold(ch.PublicID, pubID) {
			m.st.Channels = append(m.st.Channels[:i], m.st.Channels[i+1:]...)
			return nil
		}
	}

	return constants.ErrNotFound
}

func (m pgRepositoryMock) ListChannelPermissionOverrides(
	_ context.Context,
	groupID, channelPublicID string,
) ([]models.PublicChannelCategoryPermissionOverride, error) {
	m.st.Lock()
	defer m.st.Unlock()

	var chID string
	pubID := strings.TrimSpace(strings.ToLower(channelPublicID))
	for _, ch := range m.st.Channels {
		if ch.GroupID == groupID && strings.EqualFold(ch.PublicID, pubID) {
			chID = ch.ID
			break
		}
	}
	if chID == "" {
		return nil, constants.ErrNotFound
	}

	var out []models.PublicChannelCategoryPermissionOverride
	for _, o := range m.st.ChannelPermissionOverrides {
		if o.ChannelID != chID {
			continue
		}
		item := models.PublicChannelCategoryPermissionOverride{
			TargetType:       o.TargetType,
			AllowPermissions: o.AllowPermissions,
			DenyPermissions:  o.DenyPermissions,
		}
		if o.TargetType == "role" {
			for _, r := range m.st.Roles {
				if r.GroupID == groupID && r.ID == o.TargetID {
					item.TargetID = r.PublicID
					item.TargetName = r.Name
					item.TargetColor = r.Color
					item.IsEveryone = r.Slug == constants.RoleSlugEveryone
					break
				}
			}
		} else {
			for _, u := range m.st.Users {
				if u.ID == o.TargetID {
					item.TargetID = u.ID
					item.TargetName = u.DisplayName
					if item.TargetName == "" {
						item.TargetName = u.Username
					}
					item.TargetAvatarURL = u.AvatarURL
					break
				}
			}
		}
		out = append(out, item)
	}

	return out, nil
}

func (m pgRepositoryMock) UpsertChannelPermissionOverride(
	_ context.Context,
	groupID, channelPublicID, targetType, targetPublicID string,
	allow, deny []string,
	_ time.Time,
) (models.PublicChannelCategoryPermissionOverride, error) {
	m.st.Lock()
	defer m.st.Unlock()

	var chID string
	pubID := strings.TrimSpace(strings.ToLower(channelPublicID))
	for _, ch := range m.st.Channels {
		if ch.GroupID == groupID && strings.EqualFold(ch.PublicID, pubID) {
			chID = ch.ID
			break
		}
	}
	if chID == "" {
		return models.PublicChannelCategoryPermissionOverride{}, constants.ErrNotFound
	}

	var targetUUID, targetName, targetColor, targetAvatar string
	var isEveryone bool
	switch targetType {
	case "role":
		for _, r := range m.st.Roles {
			if r.GroupID == groupID && (r.PublicID == targetPublicID || r.ID == targetPublicID) {
				targetUUID = r.ID
				targetName = r.Name
				targetColor = r.Color
				isEveryone = r.Slug == constants.RoleSlugEveryone
				break
			}
		}
	case "member":
		for _, u := range m.st.Users {
			if u.ID == targetPublicID {
				targetUUID = u.ID
				targetName = u.DisplayName
				if targetName == "" {
					targetName = u.Username
				}
				targetAvatar = u.AvatarURL
				break
			}
		}
	}
	if targetUUID == "" {
		targetUUID = targetPublicID
		targetName = targetPublicID
	}

	record := models.ChannelPermissionOverride{
		ChannelID:        chID,
		TargetType:       targetType,
		TargetID:         targetUUID,
		IsEveryone:       isEveryone,
		AllowPermissions: allow,
		DenyPermissions:  deny,
	}
	replaced := false
	for i, existing := range m.st.ChannelPermissionOverrides {
		if existing.ChannelID == chID && existing.TargetType == targetType && existing.TargetID == targetUUID {
			m.st.ChannelPermissionOverrides[i] = record
			replaced = true
			break
		}
	}
	if !replaced {
		m.st.ChannelPermissionOverrides = append(m.st.ChannelPermissionOverrides, record)
	}

	return models.PublicChannelCategoryPermissionOverride{
		TargetType:       targetType,
		TargetID:         targetPublicID,
		TargetName:       targetName,
		TargetColor:      targetColor,
		TargetAvatarURL:  targetAvatar,
		IsEveryone:       isEveryone,
		AllowPermissions: allow,
		DenyPermissions:  deny,
	}, nil
}

func (m pgRepositoryMock) DeleteChannelPermissionOverride(
	_ context.Context,
	groupID, channelPublicID, targetType, targetPublicID string,
) error {
	m.st.Lock()
	defer m.st.Unlock()

	var chID string
	pubID := strings.TrimSpace(strings.ToLower(channelPublicID))
	for _, ch := range m.st.Channels {
		if ch.GroupID == groupID && strings.EqualFold(ch.PublicID, pubID) {
			chID = ch.ID
			break
		}
	}
	if chID == "" {
		return constants.ErrNotFound
	}

	var targetUUID string
	if targetType == "role" {
		for _, r := range m.st.Roles {
			if r.GroupID == groupID && (r.PublicID == targetPublicID || r.ID == targetPublicID) {
				targetUUID = r.ID
				break
			}
		}
	} else {
		for _, u := range m.st.Users {
			if u.ID == targetPublicID {
				targetUUID = u.ID
				break
			}
		}
	}

	for i, existing := range m.st.ChannelPermissionOverrides {
		if existing.ChannelID == chID && existing.TargetType == targetType && existing.TargetID == targetUUID {
			m.st.ChannelPermissionOverrides = append(
				m.st.ChannelPermissionOverrides[:i],
				m.st.ChannelPermissionOverrides[i+1:]...,
			)
			return nil
		}
	}

	return nil
}

func (m pgRepositoryMock) SyncChannelPermissionsWithCategory(
	_ context.Context,
	groupID, channelPublicID string,
	now time.Time,
) (models.Channel, error) {
	m.st.Lock()
	defer m.st.Unlock()

	pubID := strings.TrimSpace(strings.ToLower(channelPublicID))
	chIdx := -1
	for i, ch := range m.st.Channels {
		if ch.GroupID == groupID && strings.EqualFold(ch.PublicID, pubID) {
			chIdx = i
			break
		}
	}
	if chIdx == -1 || m.st.Channels[chIdx].CategoryID == "" {
		return models.Channel{}, constants.ErrNotFound
	}
	ch := m.st.Channels[chIdx]

	var cat models.ChannelCategory
	foundCat := false
	for _, c := range m.st.ChannelCategories {
		if c.GroupID == groupID && c.ID == ch.CategoryID {
			cat = c
			foundCat = true
			break
		}
	}
	if !foundCat {
		return models.Channel{}, constants.ErrNotFound
	}

	ch.IsPrivate = cat.IsPrivate
	ch.UpdatedAt = now
	m.st.Channels[chIdx] = ch

	filtered := make([]models.ChannelPermissionOverride, 0, len(m.st.ChannelPermissionOverrides))
	for _, o := range m.st.ChannelPermissionOverrides {
		if o.ChannelID != ch.ID {
			filtered = append(filtered, o)
		}
	}
	m.st.ChannelPermissionOverrides = filtered

	return ch, nil
}
