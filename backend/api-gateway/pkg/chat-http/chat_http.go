// Package chatHttp maps message JSON bodies and protobuf summaries for HTTP responses.
package chatHttp

import (
	"strconv"

	chatv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/chat/v1"
)

// ReplyPreviewBody is a compact parent message snippet in JSON.
type ReplyPreviewBody struct {
	ID                string `json:"id"`
	AuthorID          string `json:"author_id"`
	AuthorDisplayName string `json:"author_display_name"`
	BodyPreview       string `json:"body_preview"`
}

// ReactionSummaryBody aggregates reactions for one emoji on a message.
type ReactionSummaryBody struct {
	Emoji   string   `json:"emoji"`
	Count   int32    `json:"count"`
	UserIDs []string `json:"user_ids"`
}

// MessageBody is one channel message in JSON.
type MessageBody struct {
	ID                string                `json:"id"`
	ChannelSeq        string                `json:"channel_seq,omitempty"`
	AuthorID          string                `json:"author_id"`
	AuthorDisplayName string                `json:"author_display_name"`
	AuthorAvatarURL   string                `json:"author_avatar_url,omitempty"`
	Body              string                `json:"body"`
	CreatedAt         string                `json:"created_at"`
	EditedAt          string                `json:"edited_at,omitempty"`
	IsPinned          bool                  `json:"is_pinned,omitempty"`
	ReplyTo           *ReplyPreviewBody     `json:"reply_to,omitempty"`
	Reactions         []ReactionSummaryBody `json:"reactions,omitempty"`
	Mine              bool                  `json:"mine"`
}

// MessagesListBody wraps a channel message list.
type MessagesListBody struct {
	Messages []MessageBody `json:"messages"`
	HasMore  bool          `json:"has_more"`
}

// CreateMessageBody is the JSON body for POST channel messages.
type CreateMessageBody struct {
	Body      string `json:"body"`
	ReplyToID string `json:"reply_to_id,omitempty"`
}

// UpdateMessageBody is the JSON body for PATCH channel messages.
type UpdateMessageBody struct {
	Body string `json:"body"`
}

// ToggleReactionBody is the JSON body for POST message reactions.
type ToggleReactionBody struct {
	Emoji string `json:"emoji"`
}

// CreateMessageResponseBody is the created/updated message payload.
type CreateMessageResponseBody struct {
	Message MessageBody `json:"message"`
}

// ReactionsResponseBody wraps updated message reactions.
type ReactionsResponseBody struct {
	Reactions []ReactionSummaryBody `json:"reactions"`
}

// AttachmentBody is an uploaded channel file/image attachment in JSON.
type AttachmentBody struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
}

// UploadAttachmentResponseBody wraps a newly uploaded channel attachment.
type UploadAttachmentResponseBody struct {
	Attachment AttachmentBody `json:"attachment"`
}

// MessagesFromPB maps message summary protos to JSON bodies.
func MessagesFromPB(messages []*chatv1.MessageSummary) []MessageBody {
	out := make([]MessageBody, len(messages))
	for i, msg := range messages {
		out[i] = MessageFromPB(msg)
	}

	return out
}

// ReactionsFromPB maps reaction summary protos to JSON bodies.
func ReactionsFromPB(reactions []*chatv1.MessageReactionSummary) []ReactionSummaryBody {
	if len(reactions) == 0 {
		return nil
	}
	out := make([]ReactionSummaryBody, len(reactions))
	for i, r := range reactions {
		out[i] = ReactionSummaryBody{
			Emoji:   r.GetEmoji(),
			Count:   r.GetCount(),
			UserIDs: r.GetUserIds(),
		}
	}

	return out
}

// MessageFromPB maps one message summary proto to JSON.
func MessageFromPB(msg *chatv1.MessageSummary) MessageBody {
	if msg == nil {
		return MessageBody{}
	}

	var reply *ReplyPreviewBody
	if rp := msg.GetReplyTo(); rp != nil {
		reply = &ReplyPreviewBody{
			ID:                rp.GetId(),
			AuthorID:          rp.GetAuthorId(),
			AuthorDisplayName: rp.GetAuthorDisplayName(),
			BodyPreview:       rp.GetBodyPreview(),
		}
	}

	var channelSeq string
	if seq := msg.GetChannelSeq(); seq > 0 {
		channelSeq = strconv.FormatInt(seq, 10)
	}

	return MessageBody{
		ID:                msg.GetId(),
		ChannelSeq:        channelSeq,
		AuthorID:          msg.GetAuthorId(),
		AuthorDisplayName: msg.GetAuthorDisplayName(),
		AuthorAvatarURL:   msg.GetAuthorAvatarUrl(),
		Body:              msg.GetBody(),
		CreatedAt:         msg.GetCreatedAt(),
		EditedAt:          msg.GetEditedAt(),
		IsPinned:          msg.GetIsPinned(),
		ReplyTo:           reply,
		Reactions:         ReactionsFromPB(msg.GetReactions()),
		Mine:              msg.GetMine(),
	}
}

// AttachmentFromPB maps an attachment proto to JSON.
func AttachmentFromPB(att *chatv1.MessageAttachment) AttachmentBody {
	if att == nil {
		return AttachmentBody{}
	}

	return AttachmentBody{
		ID:          att.GetId(),
		URL:         att.GetUrl(),
		Filename:    att.GetFilename(),
		ContentType: att.GetContentType(),
		SizeBytes:   att.GetSizeBytes(),
	}
}
