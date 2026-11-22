package models

import (
	"time"
	"unicode/utf8"

	"github.com/uptrace/bun"
)

const replyBodyPreviewMaxRunes = 120

// MessageRow maps the messages table.
type MessageRow struct {
	bun.BaseModel `bun:"table:messages,alias:m"`
	ID            string     `bun:"id,pk"`
	PublicID      string     `bun:"public_id"`
	ChannelID     string     `bun:"channel_id"`
	AuthorID      string     `bun:"author_id"`
	ClientMsgID   *string    `bun:"client_msg_id,type:uuid"`
	ReplyToID     *string    `bun:"reply_to_id,type:uuid"`
	ChannelSeq    int64      `bun:"channel_seq,notnull"`
	Body          string     `bun:"body"`
	CreatedAt     time.Time  `bun:"created_at"`
	EditedAt      *time.Time `bun:"edited_at"`
	IsPinned      bool       `bun:"is_pinned"`
	PinnedAt      *time.Time `bun:"pinned_at"`
	PinnedBy      *string    `bun:"pinned_by,type:uuid"`
}

// MessageRowFrom builds a bun row from a domain message.
func MessageRowFrom(m Message) MessageRow {
	return MessageRow{
		ID: m.ID, PublicID: m.PublicID, ChannelID: m.ChannelID, AuthorID: m.AuthorID,
		ClientMsgID: m.ClientMsgID, ReplyToID: m.ReplyToID, ChannelSeq: m.ChannelSeq, Body: m.Body, CreatedAt: m.CreatedAt,
		EditedAt: m.EditedAt, IsPinned: m.IsPinned, PinnedAt: m.PinnedAt, PinnedBy: m.PinnedBy,
	}
}

// ToModel converts a bun row to a domain message.
func (r MessageRow) ToModel() Message {
	return Message{
		ID: r.ID, PublicID: r.PublicID, ChannelID: r.ChannelID, AuthorID: r.AuthorID,
		ClientMsgID: r.ClientMsgID, ReplyToID: r.ReplyToID, ChannelSeq: r.ChannelSeq, Body: r.Body, CreatedAt: r.CreatedAt,
		EditedAt: r.EditedAt, IsPinned: r.IsPinned, PinnedAt: r.PinnedAt, PinnedBy: r.PinnedBy,
	}
}

// MessageReactionRow maps the message_reactions table.
type MessageReactionRow struct {
	bun.BaseModel `bun:"table:message_reactions,alias:mr"`
	ID            string    `bun:"id,pk,type:uuid"`
	MessageID     string    `bun:"message_id,type:uuid"`
	UserID        string    `bun:"user_id,type:uuid"`
	Emoji         string    `bun:"emoji"`
	CreatedAt     time.Time `bun:"created_at"`
}

// MessageListRow is a list query projection with author and reply preview metadata.
type MessageListRow struct {
	bun.BaseModel     `bun:"table:messages,alias:m"`
	ID                string     `bun:"id,pk,type:uuid"`
	PublicID          string     `bun:"public_id"`
	ChannelID         string     `bun:"channel_id,type:uuid"`
	AuthorID          string     `bun:"author_id,type:uuid"`
	ClientMsgID       *string    `bun:"client_msg_id,type:uuid"`
	ReplyToID         *string    `bun:"reply_to_id,type:uuid"`
	ChannelSeq        int64      `bun:"channel_seq"`
	Body              string     `bun:"body"`
	CreatedAt         time.Time  `bun:"created_at"`
	EditedAt          *time.Time `bun:"edited_at"`
	IsPinned          bool       `bun:"is_pinned"`
	PinnedAt          *time.Time `bun:"pinned_at"`
	PinnedBy          *string    `bun:"pinned_by,type:uuid"`
	AuthorDisplayName string     `bun:"author_display_name"`
	AuthorAvatarURL   string     `bun:"author_avatar_url"`
	ReplyPublicID     *string    `bun:"reply_public_id"`
	ReplyAuthorID     *string    `bun:"reply_author_id"`
	ReplyAuthorName   *string    `bun:"reply_author_name"`
	ReplyBody         *string    `bun:"reply_body"`
}

func truncateReplyPreview(s string) string {
	if utf8.RuneCountInString(s) <= replyBodyPreviewMaxRunes {
		return s
	}
	runes := []rune(s)

	return string(runes[:replyBodyPreviewMaxRunes]) + "…"
}

// ToMessageWithAuthor converts a list projection to the domain type.
func (r MessageListRow) ToMessageWithAuthor() MessageWithAuthor {
	var reply *MessageReplyPreview
	if r.ReplyPublicID != nil && *r.ReplyPublicID != "" {
		var authorID, authorName, bodyPreview string
		if r.ReplyAuthorID != nil {
			authorID = *r.ReplyAuthorID
		}
		if r.ReplyAuthorName != nil {
			authorName = *r.ReplyAuthorName
		}
		if r.ReplyBody != nil {
			bodyPreview = truncateReplyPreview(*r.ReplyBody)
		}
		reply = &MessageReplyPreview{
			ID:                *r.ReplyPublicID,
			AuthorID:          authorID,
			AuthorDisplayName: authorName,
			BodyPreview:       bodyPreview,
		}
	}

	return MessageWithAuthor{
		Message: Message{
			ID: r.ID, PublicID: r.PublicID, ChannelID: r.ChannelID, AuthorID: r.AuthorID,
			ClientMsgID: r.ClientMsgID, ReplyToID: r.ReplyToID, ChannelSeq: r.ChannelSeq, Body: r.Body, CreatedAt: r.CreatedAt,
			EditedAt: r.EditedAt, IsPinned: r.IsPinned, PinnedAt: r.PinnedAt, PinnedBy: r.PinnedBy,
		},
		AuthorDisplayName: r.AuthorDisplayName,
		AuthorAvatarURL:   r.AuthorAvatarURL,
		ReplyTo:           reply,
	}
}
