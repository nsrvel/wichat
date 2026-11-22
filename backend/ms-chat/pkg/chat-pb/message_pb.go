// Package chatPb maps message domain models to chat v1 protobuf messages.
package chatPb

import (
	"time"

	"github.com/wichat/wichat/backend/ms-chat/internal/models"
	chatv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/chat/v1"
)

// MessageSummary maps a public message to protobuf.
func MessageSummary(msg models.PublicMessage) *chatv1.MessageSummary {
	var editedAt string
	if msg.EditedAt != nil {
		editedAt = msg.EditedAt.UTC().Format(time.RFC3339Nano)
	}
	var reply *chatv1.MessageReplyPreview
	if msg.ReplyTo != nil {
		reply = &chatv1.MessageReplyPreview{
			Id:                msg.ReplyTo.ID,
			AuthorId:          msg.ReplyTo.AuthorID,
			AuthorDisplayName: msg.ReplyTo.AuthorDisplayName,
			BodyPreview:       msg.ReplyTo.BodyPreview,
		}
	}

	return &chatv1.MessageSummary{
		Id:                msg.ID,
		ChannelSeq:        msg.ChannelSeq,
		ClientMsgId:       msg.ClientMsgID,
		AuthorId:          msg.AuthorID,
		AuthorDisplayName: msg.AuthorDisplayName,
		AuthorAvatarUrl:   msg.AuthorAvatarURL,
		Body:              msg.Body,
		CreatedAt:         msg.CreatedAt.UTC().Format(time.RFC3339Nano),
		EditedAt:          editedAt,
		IsPinned:          msg.IsPinned,
		ReplyTo:           reply,
		Reactions:         ReactionSummaries(msg.Reactions),
		Mine:              msg.Mine,
	}
}

// MessageSummaries maps a slice of public messages.
func MessageSummaries(msgs []models.PublicMessage) []*chatv1.MessageSummary {
	out := make([]*chatv1.MessageSummary, len(msgs))
	for i, msg := range msgs {
		out[i] = MessageSummary(msg)
	}

	return out
}

// ReactionSummaries maps reaction aggregates to protobuf.
func ReactionSummaries(reactions []models.MessageReactionSummary) []*chatv1.MessageReactionSummary {
	if len(reactions) == 0 {
		return nil
	}
	out := make([]*chatv1.MessageReactionSummary, len(reactions))
	for i, r := range reactions {
		out[i] = &chatv1.MessageReactionSummary{
			Emoji:   r.Emoji,
			Count:   r.Count,
			UserIds: r.UserIDs,
		}
	}

	return out
}

// MessageAttachment maps an attachment DTO to protobuf.
func MessageAttachment(att models.MessageAttachment) *chatv1.MessageAttachment {
	return &chatv1.MessageAttachment{
		Id:          att.ID,
		Url:         att.URL,
		Filename:    att.Filename,
		ContentType: att.ContentType,
		SizeBytes:   att.SizeBytes,
	}
}
