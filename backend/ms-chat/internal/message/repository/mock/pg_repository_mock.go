package mock

import (
	"context"
	"fmt"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/wichat/wichat/backend/ms-chat/internal/message/repository"
	"github.com/wichat/wichat/backend/ms-chat/internal/models"
	"github.com/wichat/wichat/backend/ms-chat/pkg/constants"
)

type reactionRecord struct {
	MessageID string
	UserID    string
	Emoji     string
}

// State is shared in-memory storage for message repository tests.
type State struct {
	mu        sync.Mutex
	seq       int
	Messages  []models.Message
	Authors   map[string]string
	Avatars   map[string]string
	Reactions []reactionRecord
}

// NewState returns empty message mock storage.
func NewState() *State {
	return &State{
		Authors: map[string]string{},
		Avatars: map[string]string{},
	}
}

type pgRepositoryMock struct {
	st *State
}

// NewPgRepositoryMock constructs an in-memory PgRepository.
func NewPgRepositoryMock(st *State) repository.PgRepository {
	return pgRepositoryMock{st: st}
}

func (m pgRepositoryMock) buildMessageWithAuthorLocked(msg models.Message) models.MessageWithAuthor {
	var reply *models.MessageReplyPreview
	if msg.ReplyToID != nil {
		for _, parent := range m.st.Messages {
			if parent.ID == *msg.ReplyToID {
				preview := parent.Body
				if utf8.RuneCountInString(preview) > 120 {
					preview = string([]rune(preview)[:120]) + "…"
				}
				reply = &models.MessageReplyPreview{
					ID:                parent.PublicID,
					AuthorID:          parent.AuthorID,
					AuthorDisplayName: m.st.Authors[parent.AuthorID],
					BodyPreview:       preview,
				}
				break
			}
		}
	}

	return models.MessageWithAuthor{
		Message:           msg,
		AuthorDisplayName: m.st.Authors[msg.AuthorID],
		AuthorAvatarURL:   m.st.Avatars[msg.AuthorID],
		ReplyTo:           reply,
		Reactions:         m.reactionsForMessageLocked(msg.ID),
	}
}

func (m pgRepositoryMock) reactionsForMessageLocked(messageID string) []models.MessageReactionSummary {
	byEmoji := make(map[string]*models.MessageReactionSummary)
	var order []string
	for _, r := range m.st.Reactions {
		if r.MessageID != messageID {
			continue
		}
		item, ok := byEmoji[r.Emoji]
		if !ok {
			item = &models.MessageReactionSummary{Emoji: r.Emoji}
			byEmoji[r.Emoji] = item
			order = append(order, r.Emoji)
		}
		item.Count++
		item.UserIDs = append(item.UserIDs, r.UserID)
	}
	out := make([]models.MessageReactionSummary, 0, len(order))
	for _, emoji := range order {
		out = append(out, *byEmoji[emoji])
	}

	return out
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
		out = append(out, m.buildMessageWithAuthorLocked(msg))
		if len(out) >= limit {
			break
		}
	}

	return out, nil
}

func (m pgRepositoryMock) ListPinnedByChannel(_ context.Context, channelID string) ([]models.MessageWithAuthor, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	var out []models.MessageWithAuthor
	for _, msg := range m.st.Messages {
		if msg.ChannelID == channelID && msg.IsPinned {
			out = append(out, m.buildMessageWithAuthorLocked(msg))
		}
	}

	return out, nil
}

func (m pgRepositoryMock) FindByPublicIDInChannel(_ context.Context, channelID, publicID string) (models.MessageWithAuthor, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for _, msg := range m.st.Messages {
		if msg.ChannelID == channelID && msg.PublicID == publicID {
			return m.buildMessageWithAuthorLocked(msg), nil
		}
	}

	return models.MessageWithAuthor{}, constants.ErrNotFound
}

func (m pgRepositoryMock) Insert(_ context.Context, channelID, authorID, body string, replyToID *string, createdAt time.Time) (models.Message, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	m.st.seq++
	pubID := fmt.Sprintf("%024x", m.st.seq)
	msg := models.Message{
		ID:        "msg-" + pubID,
		PublicID:  pubID,
		ChannelID: channelID,
		AuthorID:  authorID,
		ReplyToID: replyToID,
		Body:      body,
		CreatedAt: createdAt,
	}
	m.st.Messages = append(m.st.Messages, msg)

	return msg, nil
}

func (m pgRepositoryMock) UpdateBody(_ context.Context, channelID, publicID, authorID, body string, editedAt time.Time) (models.MessageWithAuthor, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for i, msg := range m.st.Messages {
		if msg.ChannelID == channelID && msg.PublicID == publicID && msg.AuthorID == authorID {
			msg.Body = body
			msg.EditedAt = &editedAt
			m.st.Messages[i] = msg

			return m.buildMessageWithAuthorLocked(msg), nil
		}
	}

	return models.MessageWithAuthor{}, constants.ErrNotFound
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

func (m pgRepositoryMock) DeleteByPublicID(_ context.Context, channelID, publicID string) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for i, msg := range m.st.Messages {
		if msg.ChannelID == channelID && msg.PublicID == publicID {
			m.st.Messages = append(m.st.Messages[:i], m.st.Messages[i+1:]...)

			return nil
		}
	}

	return constants.ErrNotFound
}

func (m pgRepositoryMock) ToggleReaction(_ context.Context, messageID, userID, emoji string, _ time.Time) ([]models.MessageReactionSummary, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for i, r := range m.st.Reactions {
		if r.MessageID == messageID && r.UserID == userID && r.Emoji == emoji {
			m.st.Reactions = append(m.st.Reactions[:i], m.st.Reactions[i+1:]...)

			return m.reactionsForMessageLocked(messageID), nil
		}
	}
	m.st.Reactions = append(m.st.Reactions, reactionRecord{
		MessageID: messageID,
		UserID:    userID,
		Emoji:     emoji,
	})

	return m.reactionsForMessageLocked(messageID), nil
}

func (m pgRepositoryMock) SetPinned(_ context.Context, channelID, publicID, userID string, pinned bool, now time.Time) (models.MessageWithAuthor, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for i, msg := range m.st.Messages {
		if msg.ChannelID == channelID && msg.PublicID == publicID {
			msg.IsPinned = pinned
			if pinned {
				msg.PinnedAt = &now
				msg.PinnedBy = &userID
			} else {
				msg.PinnedAt = nil
				msg.PinnedBy = nil
			}
			m.st.Messages[i] = msg

			return m.buildMessageWithAuthorLocked(msg), nil
		}
	}

	return models.MessageWithAuthor{}, constants.ErrNotFound
}

func (m pgRepositoryMock) FindUserDisplayName(_ context.Context, userID string) (string, error) {
	name, _, err := m.FindUserProfile(context.Background(), userID)

	return name, err
}

func (m pgRepositoryMock) FindUserProfile(_ context.Context, userID string) (string, string, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	if name, ok := m.st.Authors[userID]; ok {
		return name, m.st.Avatars[userID], nil
	}

	return "", "", constants.ErrNotFound
}

func (m pgRepositoryMock) FindLastMessageTimeByAuthor(_ context.Context, channelID, authorID string) (*time.Time, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	var latest *time.Time
	for _, msg := range m.st.Messages {
		if msg.ChannelID == channelID && msg.AuthorID == authorID {
			t := msg.CreatedAt
			if latest == nil || t.After(*latest) {
				latest = &t
			}
		}
	}

	return latest, nil
}
