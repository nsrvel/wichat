package mock

import (
	"context"
	"sync"
	"time"

	"github.com/wichat/wichat/backend/ms-chat/internal/message/repository"
	"github.com/wichat/wichat/backend/ms-chat/internal/models"
	"github.com/wichat/wichat/backend/ms-chat/pkg/constants"
)

// State is shared in-memory storage for message repository tests.
type State struct {
	mu       sync.Mutex
	Messages []models.Message
	Authors  map[string]string
}

// NewState returns empty message mock storage.
func NewState() *State {
	return &State{Authors: map[string]string{}}
}

type pgRepositoryMock struct {
	st *State
}

// NewPgRepositoryMock constructs an in-memory PgRepository.
func NewPgRepositoryMock(st *State) repository.PgRepository {
	return pgRepositoryMock{st: st}
}

func (m pgRepositoryMock) ListByChannel(_ context.Context, channelID string, limit int, before *time.Time) ([]models.MessageWithAuthor, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	if limit <= 0 {
		limit = constants.DefaultMessageListLimit
	}
	var out []models.MessageWithAuthor
	for _, msg := range m.st.Messages {
		if msg.ChannelID != channelID {
			continue
		}
		if before != nil && !msg.CreatedAt.Before(*before) {
			continue
		}
		out = append(out, models.MessageWithAuthor{
			Message:           msg,
			AuthorDisplayName: m.st.Authors[msg.AuthorID],
		})
		if len(out) >= limit {
			break
		}
	}

	return out, nil
}

func (m pgRepositoryMock) Insert(_ context.Context, channelID, authorID, body string, createdAt time.Time) (models.Message, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	msg := models.Message{
		ID:        "msg-" + time.Now().Format("150405.000000000"),
		ChannelID: channelID,
		AuthorID:  authorID,
		Body:      body,
		CreatedAt: createdAt,
	}
	m.st.Messages = append(m.st.Messages, msg)

	return msg, nil
}

func (m pgRepositoryMock) DeleteByID(_ context.Context, id string) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for i, msg := range m.st.Messages {
		if msg.ID == id {
			m.st.Messages = append(m.st.Messages[:i], m.st.Messages[i+1:]...)

			return nil
		}
	}

	return constants.ErrNotFound
}

func (m pgRepositoryMock) FindUserDisplayName(_ context.Context, userID string) (string, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	if name, ok := m.st.Authors[userID]; ok {
		return name, nil
	}

	return "", constants.ErrNotFound
}
