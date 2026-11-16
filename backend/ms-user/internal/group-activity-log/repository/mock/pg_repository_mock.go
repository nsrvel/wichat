package mock

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/wichat/wichat/backend/ms-user/internal/group-activity-log/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

// PgRepositoryMock is an in-memory mock for group activity log repository tests.
type PgRepositoryMock struct {
	mu   sync.RWMutex
	Logs []models.GroupActivityLog
}

var _ repository.PgRepository = (*PgRepositoryMock)(nil)

// NewPgRepositoryMock constructs an empty in-memory group activity log store.
func NewPgRepositoryMock() *PgRepositoryMock {
	return &PgRepositoryMock{}
}

func (m *PgRepositoryMock) Insert(_ context.Context, entry models.GroupActivityLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Logs = append(m.Logs, entry)
	return nil
}

func (m *PgRepositoryMock) ListByGroup(
	_ context.Context,
	filter models.GroupActivityLogFilter,
	limit int,
	beforeCreatedAt *time.Time,
	beforeID string,
) ([]models.GroupActivityLog, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > constants.MaxActivityLogLimit {
		limit = constants.DefaultActivityLogLimit
	}

	filtered := make([]models.GroupActivityLog, 0, len(m.Logs))
	for _, item := range m.Logs {
		if item.GroupID != filter.GroupID {
			continue
		}
		if filter.ActorID != "" && item.ActorID != filter.ActorID {
			continue
		}
		if filter.Category != "" && item.Category != filter.Category {
			continue
		}
		if filter.Action != "" && item.Action != filter.Action {
			continue
		}
		if beforeCreatedAt != nil && beforeID != "" {
			if item.CreatedAt.After(*beforeCreatedAt) || (item.CreatedAt.Equal(*beforeCreatedAt) && item.ID >= beforeID) {
				continue
			}
		}
		filtered = append(filtered, item)
	}

	sort.Slice(filtered, func(i, j int) bool {
		if !filtered[i].CreatedAt.Equal(filtered[j].CreatedAt) {
			return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
		}
		return filtered[i].ID > filtered[j].ID
	})

	if len(filtered) > limit {
		filtered = filtered[:limit]
	}
	return filtered, nil
}
