package models

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/uptrace/bun"
)

// EventOutboxTable maps the event_outbox table.
type EventOutboxTable struct {
	bun.BaseModel   `bun:"table:event_outbox,alias:ob"`
	ID              string     `bun:"id,pk,type:uuid"`
	EventID         string     `bun:"event_id,type:uuid,notnull"`
	EventType       string     `bun:"event_type,notnull"`
	PartitionKey    string     `bun:"partition_key,notnull"`
	Payload         []byte     `bun:"payload,type:jsonb,notnull"`
	CreatedAt       time.Time  `bun:"created_at,notnull"`
	PublishedAt     *time.Time `bun:"published_at"`
	PublishAttempts int        `bun:"publish_attempts,notnull"`
	LastError       *string    `bun:"last_error"`
}

// EventOutboxRowFromInsert builds a bun row from a domain insert DTO.
func EventOutboxRowFromInsert(id string, in EventOutboxInsert) (EventOutboxTable, error) {
	if !json.Valid(in.Payload) {
		return EventOutboxTable{}, errors.New("outbox payload: invalid json")
	}

	return EventOutboxTable{
		ID:           id,
		EventID:      in.EventID,
		EventType:    in.EventType,
		PartitionKey: in.PartitionKey,
		Payload:      in.Payload,
		CreatedAt:    in.CreatedAt,
	}, nil
}

// ToDomain converts a table row to a relay domain row.
func (r EventOutboxTable) ToDomain() EventOutboxRow {
	return EventOutboxRow{
		ID:              r.ID,
		EventID:         r.EventID,
		EventType:       r.EventType,
		PartitionKey:    r.PartitionKey,
		Payload:         r.Payload,
		CreatedAt:       r.CreatedAt,
		PublishAttempts: r.PublishAttempts,
	}
}
