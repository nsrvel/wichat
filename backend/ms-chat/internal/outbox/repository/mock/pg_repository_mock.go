package mock

import (
	"context"
	"sync"
	"time"

	"github.com/uptrace/bun"
	"github.com/wichat/wichat/backend/ms-chat/internal/models"
	"github.com/wichat/wichat/backend/ms-chat/internal/outbox/repository"
	"github.com/wichat/wichat/backend/ms-chat/pkg/constants"
)

type record struct {
	models.EventOutboxRow
	PublishedAt *time.Time
}

// State is in-memory outbox storage for tests.
type State struct {
	mu      sync.Mutex
	seq     int
	Records []record
}

// NewState returns empty outbox mock storage.
func NewState() *State {
	return &State{}
}

type pgRepositoryMock struct {
	st *State
}

// NewPgRepositoryMock constructs an in-memory outbox store.
func NewPgRepositoryMock(st *State) repository.PgRepository {
	return pgRepositoryMock{st: st}
}

func (m pgRepositoryMock) Insert(_ context.Context, in models.EventOutboxInsert) error {
	return m.insertLocked(in)
}

func (m pgRepositoryMock) InsertInTx(_ context.Context, _ bun.Tx, in models.EventOutboxInsert) error {
	return m.insertLocked(in)
}

func (m pgRepositoryMock) insertLocked(in models.EventOutboxInsert) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	m.st.seq++
	m.st.Records = append(m.st.Records, record{
		EventOutboxRow: models.EventOutboxRow{
			ID:              "outbox-" + in.EventID,
			EventID:         in.EventID,
			EventType:       in.EventType,
			PartitionKey:    in.PartitionKey,
			Payload:         in.Payload,
			CreatedAt:       in.CreatedAt,
			PublishAttempts: 0,
		},
	})

	return nil
}

func (m pgRepositoryMock) ClaimPending(_ context.Context, limit int) ([]models.EventOutboxRow, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	if limit <= 0 {
		limit = constants.OutboxPublishBatchSize
	}

	var out []models.EventOutboxRow
	for _, rec := range m.st.Records {
		if rec.PublishedAt != nil {
			continue
		}
		out = append(out, rec.EventOutboxRow)
		if len(out) >= limit {
			break
		}
	}

	return out, nil
}

func (m pgRepositoryMock) MarkPublished(_ context.Context, id string, publishedAt time.Time) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for i, rec := range m.st.Records {
		if rec.ID == id {
			if rec.PublishedAt != nil {
				return constants.ErrNotFound
			}
			t := publishedAt
			rec.PublishedAt = &t
			m.st.Records[i] = rec

			return nil
		}
	}

	return constants.ErrNotFound
}

func (m pgRepositoryMock) RecordPublishFailure(_ context.Context, id, errMsg string) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for i, rec := range m.st.Records {
		if rec.ID == id {
			rec.PublishAttempts++
			m.st.Records[i] = rec

			return nil
		}
	}

	return constants.ErrNotFound
}
