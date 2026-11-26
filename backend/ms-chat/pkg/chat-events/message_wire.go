package chatEvents

import "time"

// ReplyPreviewWire is a quoted parent snippet for event encoding.
type ReplyPreviewWire struct {
	ID                string
	AuthorID          string
	AuthorDisplayName string
	BodyPreview       string
}

// ReactionSummaryWire aggregates one emoji on a message.
type ReactionSummaryWire struct {
	Emoji   string
	Count   int32
	UserIDs []string
}

// PublicMessageWire is the message shape for Kafka chat-events (no internal/models).
type PublicMessageWire struct {
	ID                     string
	ChannelSeq             int64
	ClientMsgID            string
	AuthorID               string
	AuthorDisplayName      string
	AuthorAvatarURL        string
	Body                   string
	CreatedAt              time.Time
	EditedAt               *time.Time
	IsPinned               bool
	ReplyTo                *ReplyPreviewWire
	Reactions              []ReactionSummaryWire
	MentionedUserPublicIDs []string
}
