// Package realtime defines WebSocket chat fanout JSON (v0).
package realtime

import (
	"encoding/json"
	"fmt"
	"time"
)

const TypeMessageCreated = "message.created"

// MessagePayload mirrors REST message JSON for fanout.
type MessagePayload struct {
	ID                string `json:"id"`
	AuthorID          string `json:"author_id"`
	AuthorDisplayName string `json:"author_display_name"`
	Body              string `json:"body"`
	CreatedAt         string `json:"created_at"`
	Mine              bool   `json:"mine"`
}

// Envelope is a WebSocket fanout frame.
type Envelope struct {
	Type    string         `json:"type"`
	Message MessagePayload `json:"message"`
}

// ChannelKey is the subscription key for a group channel.
func ChannelKey(groupPublicID, channelPublicID string) string {
	return fmt.Sprintf("chat:g:%s:c:%s", groupPublicID, channelPublicID)
}

// NewMessageCreated builds a fanout envelope (mine is false for subscribers).
func NewMessageCreated(id, authorID, authorDisplayName, body string, createdAt time.Time) Envelope {
	return Envelope{
		Type: TypeMessageCreated,
		Message: MessagePayload{
			ID:                id,
			AuthorID:          authorID,
			AuthorDisplayName: authorDisplayName,
			Body:              body,
			CreatedAt:         createdAt.UTC().Format(time.RFC3339Nano),
			Mine:              false,
		},
	}
}

// MarshalEnvelope JSON-encodes an envelope for WebSocket delivery.
func MarshalEnvelope(env Envelope) ([]byte, error) {
	return json.Marshal(env)
}
