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
	mu         sync.Mutex
	seq        int
	Messages   []models.Message
	Authors    map[string]string
	Avatars    map[string]string
	Reactions  []reactionRecord
	ChannelSeq map[string]int64
	Outbox     []models.EventOutboxInsert
	// FailOutbox makes WithOutbox repository methods fail at outbox insert (message unchanged).
	FailOutbox bool
}

// NewState returns empty message mock storage.
func NewState() *State {
	return &State{
		Authors:    map[string]string{},
		Avatars:    map[string]string{},
		ChannelSeq: map[string]int64{},
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

func (m pgRepositoryMock) ListByChannel(
	_ context.Context,
	channelID string,
	limit int,
	before *time.Time,
	beforePublicID string,
	beforeChannelSeq, afterChannelSeq int64,
) ([]models.MessageWithAuthor, bool, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	var beforeInternalID string
	if beforePublicID != "" {
		for _, msg := range m.st.Messages {
			if msg.ChannelID == channelID && msg.PublicID == beforePublicID {
				beforeInternalID = msg.ID
				break
			}
		}
	}

	var matching []models.MessageWithAuthor
	for _, msg := range m.st.Messages {
		if msg.ChannelID != channelID {
			continue
		}
		if before != nil {
			if msg.CreatedAt.After(*before) {
				continue
			}
			if msg.CreatedAt.Equal(*before) && (beforeInternalID == "" || msg.ID >= beforeInternalID) {
				continue
			}
		}
		matching = append(matching, m.buildMessageWithAuthorLocked(msg))
	}

	hasMore := len(matching) > limit
	if hasMore {
		matching = matching[len(matching)-limit:]
	}

	return matching, hasMore, nil
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

func (m pgRepositoryMock) InsertWithOutbox(_ context.Context, in *models.InsertMessageWithOutboxInput) (models.Message, error) {
	if in == nil {
		return models.Message{}, fmt.Errorf("insert message input required")
	}

	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	m.st.ChannelSeq[in.ChannelID]++
	channelSeq := m.st.ChannelSeq[in.ChannelID]
	msg := models.Message{
		ID:         in.ID,
		PublicID:   in.PublicID,
		ChannelID:  in.ChannelID,
		AuthorID:   in.AuthorID,
		ReplyToID:  in.ReplyToID,
		ChannelSeq: channelSeq,
		Body:       in.Body,
		CreatedAt:  in.CreatedAt,
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

func (m pgRepositoryMock) UpdateBodyWithOutbox(_ context.Context, in *models.UpdateMessageBodyWithOutboxInput) (models.MessageWithAuthor, error) {
	if in == nil {
		return models.MessageWithAuthor{}, fmt.Errorf("update message input required")
	}

	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for i, msg := range m.st.Messages {
		if msg.ChannelID == in.ChannelID && msg.PublicID == in.PublicID && msg.AuthorID == in.AuthorID {
			if m.st.FailOutbox {
				return models.MessageWithAuthor{}, fmt.Errorf("outbox insert failed")
			}
			msg.Body = in.Body
			msg.EditedAt = &in.EditedAt
			m.st.Messages[i] = msg
			m.st.Outbox = append(m.st.Outbox, in.Outbox)

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

func (m pgRepositoryMock) DeleteByPublicIDWithOutbox(_ context.Context, in *models.DeleteMessageWithOutboxInput) error {
	if in == nil {
		return fmt.Errorf("delete message input required")
	}

	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for i, msg := range m.st.Messages {
		if msg.ChannelID == in.ChannelID && msg.PublicID == in.PublicID {
			if m.st.FailOutbox {
				return fmt.Errorf("outbox insert failed")
			}
			m.st.Messages = append(m.st.Messages[:i], m.st.Messages[i+1:]...)
			m.st.Outbox = append(m.st.Outbox, in.Outbox)

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
