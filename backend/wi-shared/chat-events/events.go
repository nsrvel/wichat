// Package chatEvents defines Kafka chat-events payloads (durable backbone).
package chatEvents

import (
	"encoding/json"
	"time"
)

const (
	TypeMessageCreated          = "message.created"
	TypeMessageUpdated          = "message.updated"
	TypeMessageDeleted          = "message.deleted"
	TypeMessageReactionsUpdated = "message.reactions.updated"
)

// ReplyPreview is a compact quoted parent message snippet.
type ReplyPreview struct {
	ID                string `json:"id"`
	AuthorID          string `json:"author_id"`
	AuthorDisplayName string `json:"author_display_name"`
	BodyPreview       string `json:"body_preview"`
}

// ReactionSummary aggregates reactions for one emoji on a message.
type ReactionSummary struct {
	Emoji   string   `json:"emoji"`
	Count   int32    `json:"count"`
	UserIDs []string `json:"user_ids"`
}

// Message is channel message fields for indexing and downstream consumers.
type Message struct {
	ID                string            `json:"id"`
	AuthorID          string            `json:"author_id"`
	AuthorDisplayName string            `json:"author_display_name"`
	AuthorAvatarURL   string            `json:"author_avatar_url,omitempty"`
	Body              string            `json:"body"`
	CreatedAt         string            `json:"created_at"`
	EditedAt          string            `json:"edited_at,omitempty"`
	IsPinned          bool              `json:"is_pinned,omitempty"`
	ReplyTo           *ReplyPreview     `json:"reply_to,omitempty"`
	Reactions         []ReactionSummary `json:"reactions,omitempty"`
}

// MessageCreated is emitted to Kafka topic chat-events after Postgres commit.
type MessageCreated struct {
	Type      string  `json:"type"`
	GroupID   string  `json:"group_id"`
	ChannelID string  `json:"channel_id"`
	Message   Message `json:"message"`
}

// MessageUpdated is emitted when a message body or pin state changes.
type MessageUpdated struct {
	Type      string  `json:"type"`
	GroupID   string  `json:"group_id"`
	ChannelID string  `json:"channel_id"`
	Message   Message `json:"message"`
}

// MessageDeleted is emitted when a message is deleted.
type MessageDeleted struct {
	Type      string `json:"type"`
	GroupID   string `json:"group_id"`
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
}

// MessageReactionsUpdated is emitted when message reactions change.
type MessageReactionsUpdated struct {
	Type      string            `json:"type"`
	GroupID   string            `json:"group_id"`
	ChannelID string            `json:"channel_id"`
	MessageID string            `json:"message_id"`
	Reactions []ReactionSummary `json:"reactions"`
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
