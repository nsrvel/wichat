package models

import (
	"time"

	"github.com/uptrace/bun"
)

// MessageRow maps the messages table.
type MessageRow struct {
	bun.BaseModel `bun:"table:messages,alias:m"`
	ID            string    `bun:"id,pk"`
	PublicID      string    `bun:"public_id"`
	ChannelID     string    `bun:"channel_id"`
	AuthorID      string    `bun:"author_id"`
	Body          string    `bun:"body"`
	CreatedAt     time.Time `bun:"created_at"`
}

// MessageRowFrom builds a bun row from a domain message.
func MessageRowFrom(m Message) MessageRow {
	return MessageRow{
		ID: m.ID, PublicID: m.PublicID, ChannelID: m.ChannelID, AuthorID: m.AuthorID,
		Body: m.Body, CreatedAt: m.CreatedAt,
	}
}

// ToModel converts a bun row to a domain message.
func (r MessageRow) ToModel() Message {
	return Message{
		ID: r.ID, PublicID: r.PublicID, ChannelID: r.ChannelID, AuthorID: r.AuthorID,
		Body: r.Body, CreatedAt: r.CreatedAt,
	}
}

// MessageListRow is a list query projection with author display name.
type MessageListRow struct {
	bun.BaseModel     `bun:"table:messages,alias:m"`
	ID                string    `bun:"id,pk,type:uuid"`
	PublicID          string    `bun:"public_id"`
	ChannelID         string    `bun:"channel_id,type:uuid"`
	AuthorID          string    `bun:"author_id,type:uuid"`
	Body              string    `bun:"body"`
	CreatedAt         time.Time `bun:"created_at"`
	AuthorDisplayName string    `bun:"author_display_name"`
}

// ToMessageWithAuthor converts a list projection to the domain type.
func (r MessageListRow) ToMessageWithAuthor() MessageWithAuthor {
	return MessageWithAuthor{
		Message: Message{
			ID: r.ID, PublicID: r.PublicID, ChannelID: r.ChannelID, AuthorID: r.AuthorID,
			Body: r.Body, CreatedAt: r.CreatedAt,
		},
		AuthorDisplayName: r.AuthorDisplayName,
	}
}
