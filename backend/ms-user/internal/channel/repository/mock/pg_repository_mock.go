package mock

import (
	"context"
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

	return out, nil
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
		Type:           constants.ChannelKindTextAndVoice,
		Name:           constants.DefaultChannelName,
		SidebarSection: constants.SidebarSectionChannels,
		CreatedBy:      createdBy,
		CreatedAt:      createdAt,
		UpdatedAt:      createdAt,
	})

	return nil
}
