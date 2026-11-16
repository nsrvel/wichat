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

func (m pgRepositoryMock) SetCategoryPositions(_ context.Context, groupID string, orderedCategoryIDs []string) error {
	m.st.Lock()
	defer m.st.Unlock()

	pos := indexPositions(orderedCategoryIDs)
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

func (m pgRepositoryMock) SeedDefaultChannels(_ context.Context, groupID, createdBy string, createdAt time.Time) error {
	m.st.Lock()
	defer m.st.Unlock()

	for _, ch := range m.st.Channels {
		if ch.GroupID == groupID && ch.Name == constants.DefaultChannelName {
			return nil
		}
	}

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
