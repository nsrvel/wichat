package mock

import (
	"context"
	"sync"
	"time"

	"github.com/uptrace/bun"
	"github.com/wichat/wichat/backend/ms-chat/internal/mention/repository"
	"github.com/wichat/wichat/backend/ms-chat/internal/models"
)

// State is in-memory mention storage for tests.
type State struct {
	mu           sync.Mutex
	Mentions     []models.MessageMentionRow
	Hits         []models.GroupMentionHit
	UnreadByUser map[string]int
	ReadAt       map[string]time.Time
}

// NewState returns empty mention mock storage.
func NewState() *State {
	return &State{
		UnreadByUser: map[string]int{},
		ReadAt:       map[string]time.Time{},
	}
}

type pgRepositoryMock struct {
	st *State
}

// NewPgRepositoryMock constructs an in-memory PgRepository.
func NewPgRepositoryMock(st *State) repository.PgRepository {
	return pgRepositoryMock{st: st}
}

func (m pgRepositoryMock) InsertMentions(_ context.Context, _ bun.Tx, rows []models.MessageMentionRow) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()
	m.st.Mentions = append(m.st.Mentions, rows...)
	return nil
}

func (pgRepositoryMock) ResolveUserPublicIDsForChannel(context.Context, string, []string) ([]string, error) {
	return nil, nil
}

func (pgRepositoryMock) LookupGroupInternalID(context.Context, string) (string, error) {
	return "group-int", nil
}

func (pgRepositoryMock) LookupUserInternalID(context.Context, string) (string, error) {
	return "user-int", nil
}

func (m pgRepositoryMock) ListForMentionedUser(
	_ context.Context,
	_, mentionedUserID string,
	limit int,
	_ time.Time,
	_ string,
	_ bool,
) ([]models.GroupMentionHit, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()
	if limit <= 0 || limit > len(m.st.Hits) {
		limit = len(m.st.Hits)
	}
	if limit == 0 {
		return nil, nil
	}
	_ = mentionedUserID
	out := make([]models.GroupMentionHit, 0, limit)
	for _, h := range m.st.Hits {
		out = append(out, h)
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (m pgRepositoryMock) MarkRead(_ context.Context, _, mentionedUserID string, messagePublicIDs []string, at time.Time) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()
	for _, id := range messagePublicIDs {
		m.st.ReadAt[id] = at
	}
	if m.st.UnreadByUser[mentionedUserID] > 0 {
		m.st.UnreadByUser[mentionedUserID]--
	}
	return nil
}

func (m pgRepositoryMock) CountUnread(_ context.Context, _, mentionedUserID string) (int, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()
	return m.st.UnreadByUser[mentionedUserID], nil
}

func (pgRepositoryMock) MentionedPublicUserIDsByMessageIDs(context.Context, []string) (map[string][]string, error) {
	return map[string][]string{}, nil
}