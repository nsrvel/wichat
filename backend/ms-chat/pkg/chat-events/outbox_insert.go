package chatEvents

import "time"

// OutboxInsert is a chat event row for the outbox table (Kafka relay).
type OutboxInsert struct {
	EventID      string
	EventType    string
	PartitionKey string
	Payload      []byte
	CreatedAt    time.Time
}
