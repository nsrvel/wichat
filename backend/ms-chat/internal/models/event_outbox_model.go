package models

import "time"

// EventOutboxInsert is a row to insert inside the message create transaction.
type EventOutboxInsert struct {
	EventID      string
	EventType    string
	PartitionKey string
	Payload      []byte
	CreatedAt    time.Time
}

// InsertMessageWithOutboxInput bundles message persistence and outbox enqueue.
type InsertMessageWithOutboxInput struct {
	ID        string
	PublicID  string
	ChannelID string
	AuthorID  string
	Body      string
	ReplyToID *string
	CreatedAt time.Time
	Outbox    EventOutboxInsert
}

// EventOutboxRow is a claimed unpublished outbox row.
type EventOutboxRow struct {
	ID              string
	EventID         string
	EventType       string
	PartitionKey    string
	Payload         []byte
	CreatedAt       time.Time
	PublishAttempts int
}
