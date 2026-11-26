package models

import (
	"time"

	"github.com/uptrace/bun"
)

// MessageMentionRow maps message_mentions.
type MessageMentionRow struct {
	bun.BaseModel   `bun:"table:message_mentions"`
	MessageID       string     `bun:"message_id,pk"`
	GroupID         string     `bun:"group_id"`
	ChannelID       string     `bun:"channel_id"`
	MentionedUserID string     `bun:"mentioned_user_id,pk"`
	AuthorID        string     `bun:"author_id"`
	CreatedAt       time.Time  `bun:"created_at"`
	ReadAt          *time.Time `bun:"read_at"`
}

// GroupMentionHit is one inbox row for the caller.
type GroupMentionHit struct {
	MessageID         string
	ChannelID         string
	ChannelName       string
	ChannelType       string
	AuthorDisplayName string
	BodyPreview       string
	CreatedAt         time.Time
	IsUnread          bool
}

// MarkGroupMentionsReadInput marks mention inbox rows read for the token bearer.
type MarkGroupMentionsReadInput struct {
	AccessToken      string
	GroupPublicID    string
	MessagePublicIDs []string
}

// ListGroupMentionsInput lists @mention inbox for the token bearer in a group.
type ListGroupMentionsInput struct {
	AccessToken   string
	GroupPublicID string
	CallerUserID  string
	Limit         int
	Cursor        string
	Kind          string
}
