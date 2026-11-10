package mock

import (
	"context"
	"strings"
	"time"

	"github.com/wichat/wichat/backend/ms-user/internal/invite/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

type pgRepositoryMock struct {
	st *usermock.State
}

// NewPgRepositoryMock is an in-memory invite store backed by shared test state.
func NewPgRepositoryMock(st *usermock.State) repository.PgRepository {
	return pgRepositoryMock{st: st}
}

func (m pgRepositoryMock) InsertInvite(_ context.Context, row models.GroupInviteRow) error {
	m.st.Lock()
	defer m.st.Unlock()

	for _, existing := range m.st.Invites {
		if existing.Token == row.Token {
			return constants.ErrConflict
		}
	}
	m.st.Invites = append(m.st.Invites, row)

	return nil
}

func (m pgRepositoryMock) FindOpenShareLink(_ context.Context, groupID string, channelID *string, now time.Time) (models.GroupInviteRow, error) {
	m.st.Lock()
	defer m.st.Unlock()

	var best models.GroupInviteRow
	found := false
	for _, row := range m.st.Invites {
		if row.GroupID != groupID || row.Kind != constants.GroupInviteKindShareLink {
			continue
		}
		if row.RedeemedAt != nil || !row.ExpiresAt.After(now) {
			continue
		}
		if len(row.Token) != constants.GroupInviteTokenLen {
			continue
		}
		if channelID == nil || strings.TrimSpace(*channelID) == "" {
			if row.ChannelID != nil && strings.TrimSpace(*row.ChannelID) != "" {
				continue
			}
		} else if row.ChannelID == nil || *row.ChannelID != *channelID {
			continue
		}
		if !found || row.CreatedAt.After(best.CreatedAt) {
			best = row
			found = true
		}
	}
	if !found {
		return models.GroupInviteRow{}, constants.ErrNotFound
	}

	return best, nil
}

func (m pgRepositoryMock) FindByToken(_ context.Context, token string) (models.GroupInviteRow, error) {
	m.st.Lock()
	defer m.st.Unlock()

	token = strings.TrimSpace(token)
	for _, row := range m.st.Invites {
		if row.Token == token {
			return row, nil
		}
	}

	return models.GroupInviteRow{}, constants.ErrNotFound
}

func (m pgRepositoryMock) MarkRedeemed(_ context.Context, id, userID string, now time.Time) error {
	m.st.Lock()
	defer m.st.Unlock()

	for i := range m.st.Invites {
		if m.st.Invites[i].ID != id {
			continue
		}
		if m.st.Invites[i].RedeemedAt != nil {
			return constants.ErrNotFound
		}
		m.st.Invites[i].RedeemedAt = &now
		uid := userID
		m.st.Invites[i].RedeemedByUserID = &uid

		return nil
	}

	return constants.ErrNotFound
}
