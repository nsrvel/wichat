package mock

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/internal/user-activity-log/repository"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

// State holds in-memory user activity logs for unit tests.
type State struct {
	mu   sync.Mutex
	Logs []models.UserActivityLog
}

// NewState creates an empty mock store for user activity logs.
func NewState() *State {
	return &State{}
}

type pgRepositoryMock struct {
	st *State
}

// NewPgRepositoryMock returns an in-memory user activity log repository.
func NewPgRepositoryMock(st *State) repository.PgRepository {
	return &pgRepositoryMock{st: st}
}

func (m *pgRepositoryMock) Insert(_ context.Context, entry models.UserActivityLog) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	if entry.Changes == nil {
		entry.Changes = map[string]models.ActivityFieldChange{}
	}
	if entry.Metadata == nil {
		entry.Metadata = map[string]any{}
	}
	m.st.Logs = append(m.st.Logs, entry)
	return nil
}

func (m *pgRepositoryMock) ListByUser(
	_ context.Context,
	userID, category, action string,
	limit int,
	beforeCreatedAt *time.Time,
	beforeID string,
) ([]models.UserActivityLog, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	if limit <= 0 || limit > constants.MaxActivityLogLimit {
		limit = constants.DefaultActivityLogLimit
	}

	filtered := make([]models.UserActivityLog, 0)
	for _, entry := range m.st.Logs {
		if entry.UserID != userID {
			continue
		}
		if category != "" && entry.Category != category {
			continue
		}
		if action != "" && entry.Action != action {
			continue
		}
		if beforeCreatedAt != nil && beforeID != "" {
			if entry.CreatedAt.After(*beforeCreatedAt) {
				continue
			}
			if entry.CreatedAt.Equal(*beforeCreatedAt) && entry.ID >= beforeID {
				continue
			}
		}
		filtered = append(filtered, entry)
	}

	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].CreatedAt.Equal(filtered[j].CreatedAt) {
			return filtered[i].ID > filtered[j].ID
		}
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	if len(filtered) > limit {
		filtered = filtered[:limit]
	}
	return filtered, nil
}
