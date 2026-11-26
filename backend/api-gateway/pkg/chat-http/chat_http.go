// Package chatHttp maps message JSON bodies and protobuf summaries for HTTP responses.
package chatHttp

import (
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
	ChannelSeq        int64                 `json:"channel_seq,omitempty"`
	ClientMsgID       string                `json:"client_msg_id,omitempty"`
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
	MentionedUserIDs  []string              `json:"mentioned_user_ids,omitempty"`
}

// MessagesListBody wraps a channel message list.
type MessagesListBody struct {
	Messages []MessageBody `json:"messages"`
	HasMore  bool          `json:"has_more"`
}

// MessageChannelSeqBody resolves a message id to channel_seq.
type MessageChannelSeqBody struct {
	MessageID  string `json:"message_id"`
	GroupID    string `json:"group_id"`
	ChannelID  string `json:"channel_id"`
	ChannelSeq int64  `json:"channel_seq"`
}

// CreateMessageBody is the JSON body for POST channel messages.
type CreateMessageBody struct {
	Body             string   `json:"body"`
	ReplyToID        string   `json:"reply_to_id,omitempty"`
	ClientMsgID      string   `json:"client_msg_id,omitempty"`
	MentionedUserIDs []string `json:"mentioned_user_ids,omitempty"`
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

	return MessageBody{
		ID:                msg.GetId(),
		ChannelSeq:        msg.GetChannelSeq(),
		ClientMsgID:       msg.GetClientMsgId(),
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
		MentionedUserIDs:  msg.GetMentionedUserIds(),
	}
}

// GroupMentionHitBody is one @mention inbox row in JSON.
type GroupMentionHitBody struct {
	MessageID         string `json:"message_id"`
	ChannelID         string `json:"channel_id"`
	ChannelName       string `json:"channel_name"`
	ChannelType       string `json:"channel_type"`
	AuthorDisplayName string `json:"author_display_name"`
	BodyPreview       string `json:"body_preview"`
	CreatedAt         string `json:"created_at"`
	IsUnread          bool   `json:"is_unread"`
}

// MarkMentionsReadBody is POST body to mark mention inbox rows read.
type MarkMentionsReadBody struct {
	MessageIDs []string `json:"message_ids"`
}

// MentionsUnreadCountBody is unread mention count for a group.
type MentionsUnreadCountBody struct {
	Count int32 `json:"count"`
}

// GroupMentionsListBody wraps mention inbox results.
type GroupMentionsListBody struct {
	Hits       []GroupMentionHitBody `json:"hits"`
	NextCursor string                `json:"next_cursor,omitempty"`
}

// GroupMentionsFromPB maps proto mention hits to JSON.
func GroupMentionsFromPB(hits []*chatv1.GroupMentionHit) []GroupMentionHitBody {
	if len(hits) == 0 {
		return []GroupMentionHitBody{}
	}
	out := make([]GroupMentionHitBody, len(hits))
	for i, h := range hits {
		out[i] = GroupMentionHitBody{
			MessageID:         h.GetMessageId(),
			ChannelID:         h.GetChannelId(),
			ChannelName:       h.GetChannelName(),
			ChannelType:       h.GetChannelType(),
			AuthorDisplayName: h.GetAuthorDisplayName(),
			BodyPreview:       h.GetBodyPreview(),
			CreatedAt:         h.GetCreatedAt(),
			IsUnread:          h.GetIsUnread(),
		}
	}
	return out
}

// MessageSearchHitBody is one message search hit in JSON.
type MessageSearchHitBody struct {
	MessageID         string `json:"message_id"`
	GroupID           string `json:"group_id"`
	ChannelID         string `json:"channel_id"`
	BodyPreview       string `json:"body_preview"`
	AuthorDisplayName string `json:"author_display_name"`
	CreatedAt         string `json:"created_at"`
}

// MessageSearchHitsBody wraps message search results.
type MessageSearchHitsBody struct {
	Hits []MessageSearchHitBody `json:"hits"`
}

// MessageSearchHitsFromPB maps proto search hits to JSON.
func MessageSearchHitsFromPB(hits []*chatv1.MessageSearchHit) []MessageSearchHitBody {
	if len(hits) == 0 {
		return []MessageSearchHitBody{}
	}
	out := make([]MessageSearchHitBody, 0, len(hits))
	for _, hit := range hits {
		if hit == nil {
			continue
		}
		out = append(out, MessageSearchHitBody{
			MessageID:         hit.GetMessageId(),
			GroupID:           hit.GetGroupId(),
			ChannelID:         hit.GetChannelId(),
			BodyPreview:       hit.GetBodyPreview(),
			AuthorDisplayName: hit.GetAuthorDisplayName(),
			CreatedAt:         hit.GetCreatedAt(),
		})
	}
	return out
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
