package models

import "time"

// Message is a stored channel message.
type Message struct {
	ID        string
	PublicID  string
	ChannelID string
	AuthorID  string
	Body      string
	CreatedAt time.Time
}

// MessageWithAuthor is a message row joined with author display metadata.
type MessageWithAuthor struct {
	Message
	AuthorDisplayName string
}

// PublicMessage is returned through gRPC and REST.
type PublicMessage struct {
	ID                string
	AuthorID          string
	AuthorDisplayName string
	Body              string
	CreatedAt         time.Time
	Mine              bool
}

// ListChannelMessagesInput lists messages in a channel after access resolution.
type ListChannelMessagesInput struct {
	AccessToken       string
	GroupPublicID     string
	ChannelPublicID   string
	CallerUserID      string
	ResolvedChannelID string
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
	Body              string
}
