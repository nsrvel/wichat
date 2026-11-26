package chatEvents

import (
	"time"

	"github.com/google/uuid"
	events "github.com/wichat/wichat/backend/wi-shared/chat-events"
)

func newEventID() string {
	return uuid.NewString()
}

func wireMessage(msg PublicMessageWire) events.Message {
	var reply *events.ReplyPreview
	if msg.ReplyTo != nil {
		reply = &events.ReplyPreview{
			ID:                msg.ReplyTo.ID,
			AuthorID:          msg.ReplyTo.AuthorID,
			AuthorDisplayName: msg.ReplyTo.AuthorDisplayName,
			BodyPreview:       msg.ReplyTo.BodyPreview,
		}
	}
	var reactions []events.ReactionSummary
	if len(msg.Reactions) > 0 {
		reactions = wireReactions(msg.Reactions)
	}
	var editedAt string
	if msg.EditedAt != nil {
		editedAt = msg.EditedAt.UTC().Format(time.RFC3339Nano)
	}

	return events.Message{
		ID:                msg.ID,
		ChannelSeq:        msg.ChannelSeq,
		ClientMsgID:       msg.ClientMsgID,
		AuthorID:          msg.AuthorID,
		AuthorDisplayName: msg.AuthorDisplayName,
		AuthorAvatarURL:   msg.AuthorAvatarURL,
		Body:              msg.Body,
		CreatedAt:         msg.CreatedAt.UTC().Format(time.RFC3339Nano),
		EditedAt:          editedAt,
		IsPinned:          msg.IsPinned,
		ReplyTo:           reply,
		Reactions:         reactions,
		MentionedUserIDs:  msg.MentionedUserPublicIDs,
	}
}

func wireReactions(in []ReactionSummaryWire) []events.ReactionSummary {
	out := make([]events.ReactionSummary, len(in))
	for i, r := range in {
		out[i] = events.ReactionSummary{
			Emoji:   r.Emoji,
			Count:   r.Count,
			UserIDs: r.UserIDs,
		}
	}

	return out
}

func encodeOutbox(partitionKey string, payload []byte, eventType string, createdAt time.Time) (OutboxInsert, error) {
	return OutboxInsert{
		EventID:      newEventID(),
		EventType:    eventType,
		PartitionKey: partitionKey,
		Payload:      payload,
		CreatedAt:    createdAt,
	}, nil
}
