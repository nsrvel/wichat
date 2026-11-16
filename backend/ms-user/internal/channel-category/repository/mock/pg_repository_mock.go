package mock

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/wichat/wichat/backend/ms-user/internal/channel-category/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/ms-user/pkg/token"
)

type pgRepositoryMock struct {
	st *usermock.State
}

// NewPgRepositoryMock wraps shared in-memory state as a channel category repository.
func NewPgRepositoryMock(st *usermock.State) repository.PgRepository {
	return pgRepositoryMock{st: st}
}

func (m pgRepositoryMock) ListCategoriesByGroupID(_ context.Context, groupID string) ([]models.ChannelCategory, error) {
	m.st.Lock()
	defer m.st.Unlock()

	var out []models.ChannelCategory
	for _, cat := range m.st.ChannelCategories {
		if cat.GroupID == groupID {
			out = append(out, cat)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Position != out[j].Position {
			return out[i].Position < out[j].Position
		}

		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})

	return out, nil
}

func (m pgRepositoryMock) CreateCategory(
	_ context.Context,
	groupID, name string,
	isPrivate bool,
	createdBy string,
	now time.Time,
) (models.ChannelCategory, error) {
	m.st.Lock()
	defer m.st.Unlock()

	maxPos := int32(-1)
	for _, cat := range m.st.ChannelCategories {
		if cat.GroupID == groupID && cat.Position > maxPos {
			maxPos = cat.Position
		}
	}

	publicID, err := token.NewPublicID()
	if err != nil {
		return models.ChannelCategory{}, err
	}
	id, err := token.NewID()
	if err != nil {
		return models.ChannelCategory{}, err
	}

	cat := models.ChannelCategory{
		ID: id, PublicID: publicID, GroupID: groupID, Name: name,
		IsPrivate: isPrivate, Position: maxPos + 1,
		CreatedBy: createdBy, CreatedAt: now, UpdatedAt: now,
	}
	m.st.ChannelCategories = append(m.st.ChannelCategories, cat)

	return cat, nil
}

func (m pgRepositoryMock) UpdateCategory(
	_ context.Context,
	groupID, categoryPublicID string,
	name *string,
	isPrivate *bool,
	now time.Time,
) (models.ChannelCategory, error) {
	m.st.Lock()
	defer m.st.Unlock()

	pubID := strings.TrimSpace(strings.ToLower(categoryPublicID))
	for i, cat := range m.st.ChannelCategories {
		if cat.GroupID == groupID && strings.EqualFold(cat.PublicID, pubID) {
			if name != nil {
				cat.Name = *name
			}
			if isPrivate != nil {
				cat.IsPrivate = *isPrivate
			}
			cat.UpdatedAt = now
			m.st.ChannelCategories[i] = cat

			return cat, nil
		}
	}

	return models.ChannelCategory{}, constants.ErrNotFound
}

func (m pgRepositoryMock) DeleteCategory(_ context.Context, groupID, categoryPublicID string) error {
	m.st.Lock()
	defer m.st.Unlock()

	pubID := strings.TrimSpace(strings.ToLower(categoryPublicID))
	var catID string
	for i, cat := range m.st.ChannelCategories {
		if cat.GroupID == groupID && strings.EqualFold(cat.PublicID, pubID) {
			catID = cat.ID
			m.st.ChannelCategories = append(m.st.ChannelCategories[:i], m.st.ChannelCategories[i+1:]...)
			break
		}
	}
	if catID == "" {
		return constants.ErrNotFound
	}
	for i := range m.st.Channels {
		if m.st.Channels[i].GroupID == groupID && m.st.Channels[i].CategoryID == catID {
			m.st.Channels[i].CategoryID = ""
			m.st.Channels[i].CategoryPublicID = ""
		}
	}

	return nil
}

func (m pgRepositoryMock) ListCategoryPermissionOverrides(
	_ context.Context,
	groupID, categoryPublicID string,
) ([]models.PublicChannelCategoryPermissionOverride, error) {
	m.st.Lock()
	defer m.st.Unlock()

	var catID string
	pubID := strings.TrimSpace(strings.ToLower(categoryPublicID))
	for _, cat := range m.st.ChannelCategories {
		if cat.GroupID == groupID && strings.EqualFold(cat.PublicID, pubID) {
			catID = cat.ID
			break
		}
	}
	if catID == "" {
		return nil, constants.ErrNotFound
	}

	var out []models.PublicChannelCategoryPermissionOverride
	for _, o := range m.st.CategoryPermissionOverrides {
		if o.CategoryID != catID {
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

func (m pgRepositoryMock) ListGroupCategoryPermissionOverrides(
	_ context.Context,
	groupID string,
) ([]models.ChannelCategoryPermissionOverride, error) {
	m.st.Lock()
	defer m.st.Unlock()

	catIDs := make(map[string]struct{})
	for _, cat := range m.st.ChannelCategories {
		if cat.GroupID == groupID {
			catIDs[cat.ID] = struct{}{}
		}
	}
	var out []models.ChannelCategoryPermissionOverride
	for _, o := range m.st.CategoryPermissionOverrides {
		if _, ok := catIDs[o.CategoryID]; ok {
			out = append(out, o)
		}
	}

	return out, nil
}

func (m pgRepositoryMock) UpsertCategoryPermissionOverride(
	_ context.Context,
	groupID, categoryPublicID, targetType, targetPublicID string,
	allow, deny []string,
	_ time.Time,
) (models.PublicChannelCategoryPermissionOverride, error) {
	m.st.Lock()
	defer m.st.Unlock()

	var catID string
	pubID := strings.TrimSpace(strings.ToLower(categoryPublicID))
	for _, cat := range m.st.ChannelCategories {
		if cat.GroupID == groupID && strings.EqualFold(cat.PublicID, pubID) {
			catID = cat.ID
			break
		}
	}
	if catID == "" {
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

	record := models.ChannelCategoryPermissionOverride{
		CategoryID:       catID,
		TargetType:       targetType,
		TargetID:         targetUUID,
		IsEveryone:       isEveryone,
		AllowPermissions: allow,
		DenyPermissions:  deny,
	}
	replaced := false
	for i, existing := range m.st.CategoryPermissionOverrides {
		if existing.CategoryID == catID && existing.TargetType == targetType && existing.TargetID == targetUUID {
			m.st.CategoryPermissionOverrides[i] = record
			replaced = true
			break
		}
	}
	if !replaced {
		m.st.CategoryPermissionOverrides = append(m.st.CategoryPermissionOverrides, record)
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

func (m pgRepositoryMock) DeleteCategoryPermissionOverride(
	_ context.Context,
	groupID, categoryPublicID, targetType, targetPublicID string,
) error {
	m.st.Lock()
	defer m.st.Unlock()

	var catID string
	pubID := strings.TrimSpace(strings.ToLower(categoryPublicID))
	for _, cat := range m.st.ChannelCategories {
		if cat.GroupID == groupID && strings.EqualFold(cat.PublicID, pubID) {
			catID = cat.ID
			break
		}
	}
	if catID == "" {
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

	for i, existing := range m.st.CategoryPermissionOverrides {
		if existing.CategoryID == catID && existing.TargetType == targetType && existing.TargetID == targetUUID {
			m.st.CategoryPermissionOverrides = append(
				m.st.CategoryPermissionOverrides[:i],
				m.st.CategoryPermissionOverrides[i+1:]...,
			)

			return nil
		}
	}

	return nil
}

func (m pgRepositoryMock) SetCategoryPositions(_ context.Context, groupID string, orderedCategoryIDs []string) error {
	m.st.Lock()
	defer m.st.Unlock()

	pos := make(map[string]int32, len(orderedCategoryIDs))
	var idx int32
	for _, id := range orderedCategoryIDs {
		pos[id] = idx
		idx++
	}
	updated := 0
	for i := range m.st.ChannelCategories {
		if m.st.ChannelCategories[i].GroupID == groupID {
			if p, ok := pos[m.st.ChannelCategories[i].ID]; ok {
				m.st.ChannelCategories[i].Position = p
				updated++
			}
		}
	}
	if updated != len(orderedCategoryIDs) {
		return constants.ErrNotFound
	}

	return nil
}
