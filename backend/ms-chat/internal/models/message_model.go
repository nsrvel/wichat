package models

import "time"

// MessageReplyPreview is a compact quoted parent message snippet.
type MessageReplyPreview struct {
	ID                string
	AuthorID          string
	AuthorDisplayName string
	BodyPreview       string
}

// MessageReactionSummary aggregates reactions for one emoji on a message.
type MessageReactionSummary struct {
	Emoji   string
	Count   int32
	UserIDs []string
}

// Message is a stored channel message.
type Message struct {
	ID        string
	PublicID  string
	ChannelID string
	AuthorID  string
	ReplyToID *string
	Body      string
	CreatedAt time.Time
	EditedAt  *time.Time
	IsPinned  bool
	PinnedAt  *time.Time
	PinnedBy  *string
}

// MessageWithAuthor is a message row joined with author and reply metadata.
type MessageWithAuthor struct {
	Message
	AuthorDisplayName string
	AuthorAvatarURL   string
	ReplyTo           *MessageReplyPreview
	Reactions         []MessageReactionSummary
}

// PublicMessage is returned through gRPC and REST.
type PublicMessage struct {
	ID                string
	AuthorID          string
	AuthorDisplayName string
	AuthorAvatarURL   string
	Body              string
	CreatedAt         time.Time
	EditedAt          *time.Time
	IsPinned          bool
	ReplyTo           *MessageReplyPreview
	Reactions         []MessageReactionSummary
	Mine              bool
}

// ChannelAccessInfo holds resolved channel ids and effective caller permissions/settings.
type ChannelAccessInfo struct {
	ChannelUUID     string
	UserUUID        string
	Permissions     []string
	ChannelType     string
	SlowmodeSeconds int32
}

// ListChannelMessagesInput lists messages in a channel after access resolution.
type ListChannelMessagesInput struct {
	AccessToken       string
	GroupPublicID     string
	ChannelPublicID   string
	CallerUserID      string
	ResolvedChannelID string
	CallerPermissions []string
	Limit             int
	BeforeCreatedAt   *time.Time
}

// CreateChannelMessageInput creates a message after access resolution.
type CreateChannelMessageInput struct {
	AccessToken       string
	GroupPublicID     string
	ChannelPublicID   string
	CallerUserID      string
	ResolvedChannelID string
	CallerPermissions []string
	ChannelType       string
	SlowmodeSeconds   int32
	Body              string
	ReplyToPublicID   string
}

// UpdateChannelMessageInput edits an existing message body.
type UpdateChannelMessageInput struct {
	GroupPublicID     string
	ChannelPublicID   string
	CallerUserID      string
	ResolvedChannelID string
	MessagePublicID   string
	Body              string
}

// DeleteChannelMessageInput deletes an existing message.
type DeleteChannelMessageInput struct {
	GroupPublicID     string
	ChannelPublicID   string
	CallerUserID      string
	ResolvedChannelID string
	CallerPermissions []string
	MessagePublicID   string
}

// ToggleMessageReactionInput toggles an emoji reaction on a message.
type ToggleMessageReactionInput struct {
	GroupPublicID     string
	ChannelPublicID   string
	CallerUserID      string
	ResolvedChannelID string
	CallerPermissions []string
	MessagePublicID   string
	Emoji             string
}

// SetChannelMessagePinnedInput pins or unpins a message in a channel.
type SetChannelMessagePinnedInput struct {
	GroupPublicID     string
	ChannelPublicID   string
	CallerUserID      string
	ResolvedChannelID string
	CallerPermissions []string
	MessagePublicID   string
	Pinned            bool
}

// ListPinnedChannelMessagesInput lists pinned messages in a channel.
type ListPinnedChannelMessagesInput struct {
	CallerUserID      string
	ResolvedChannelID string
	CallerPermissions []string
}

// MessageAttachment is a file or image uploaded to a channel.
type MessageAttachment struct {
	ID          string
	URL         string
	Filename    string
	ContentType string
	SizeBytes   int64
}

// UploadChannelAttachmentInput uploads a file or image attachment to a channel.
type UploadChannelAttachmentInput struct {
	GroupPublicID     string
	ChannelPublicID   string
	CallerUserID      string
	ResolvedChannelID string
	CallerPermissions []string
	ChannelType       string
	Filename          string
	ContentType       string
	Data              []byte
}
