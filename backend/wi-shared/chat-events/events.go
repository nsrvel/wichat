// Package chatEvents defines Kafka chat-events payloads (durable backbone).
package chatEvents

import (
	"encoding/json"
	"time"
)

const TypeMessageCreated = "message.created"

// Message is channel message fields for indexing and downstream consumers.
type Message struct {
	ID                string `json:"id"`
	AuthorID          string `json:"author_id"`
	AuthorDisplayName string `json:"author_display_name"`
	Body              string `json:"body"`
	CreatedAt         string `json:"created_at"`
}

// MessageCreated is emitted to Kafka topic chat-events after Postgres commit.
type MessageCreated struct {
	Type      string  `json:"type"`
	GroupID   string  `json:"group_id"`
	ChannelID string  `json:"channel_id"`
	Message   Message `json:"message"`
}

// NewMessageCreated builds a chat-events payload.
func NewMessageCreated(groupID, channelID, id, authorID, authorDisplayName, body string, createdAt time.Time) MessageCreated {
	return MessageCreated{
		Type:      TypeMessageCreated,
		GroupID:   groupID,
		ChannelID: channelID,
		Message: Message{
			ID:                id,
			AuthorID:          authorID,
			AuthorDisplayName: authorDisplayName,
			Body:              body,
			CreatedAt:         createdAt.UTC().Format(time.RFC3339Nano),
		},
	}
}

// Marshal JSON-encodes an event for Kafka.
func Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}
