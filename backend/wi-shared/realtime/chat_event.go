// Package realtime defines WebSocket chat fanout JSON.
package realtime

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	TypeMessageCreated          = "message.created"
	TypeMessageUpdated          = "message.updated"
	TypeMessageDeleted          = "message.deleted"
	TypeMessageReactionsUpdated = "message.reactions.updated"
	TypeVoiceOccupancyUpdated   = "voice.occupancy.updated"
	TypeVoiceRecordingUpdated   = "voice.recording.updated"
	TypeVoiceModerationApplied  = "voice.moderation.applied"
	TypeVoiceIdleWarning        = "voice.idle.warning"
)

// ReplyPreviewPayload mirrors REST reply preview JSON for fanout.
type ReplyPreviewPayload struct {
	ID                string `json:"id"`
	AuthorID          string `json:"author_id"`
	AuthorDisplayName string `json:"author_display_name"`
	BodyPreview       string `json:"body_preview"`
}

// ReactionPayload mirrors REST reaction summary JSON for fanout.
type ReactionPayload struct {
	Emoji   string   `json:"emoji"`
	Count   int32    `json:"count"`
	UserIDs []string `json:"user_ids"`
}

// MessagePayload mirrors REST message JSON for fanout.
type MessagePayload struct {
	ID                string               `json:"id"`
	AuthorID          string               `json:"author_id"`
	AuthorDisplayName string               `json:"author_display_name"`
	AuthorAvatarURL   string               `json:"author_avatar_url,omitempty"`
	Body              string               `json:"body"`
	CreatedAt         string               `json:"created_at"`
	EditedAt          string               `json:"edited_at,omitempty"`
	IsPinned          bool                 `json:"is_pinned,omitempty"`
	ReplyTo           *ReplyPreviewPayload `json:"reply_to,omitempty"`
	Reactions         []ReactionPayload    `json:"reactions,omitempty"`
	Mine              bool                 `json:"mine"`
}

// Envelope is a WebSocket fanout frame for message.created / message.updated.
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
