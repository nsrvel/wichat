package realtime

import (
	"encoding/json"

	chatEvents "github.com/wichat/wichat/backend/wi-shared/chat-events"
)

func messagePayloadFromEvent(m chatEvents.Message) MessagePayload {
	var reply *ReplyPreviewPayload
	if m.ReplyTo != nil {
		reply = &ReplyPreviewPayload{
			ID:                m.ReplyTo.ID,
			AuthorID:          m.ReplyTo.AuthorID,
			AuthorDisplayName: m.ReplyTo.AuthorDisplayName,
			BodyPreview:       m.ReplyTo.BodyPreview,
		}
	}
	var reactions []ReactionPayload
	if len(m.Reactions) > 0 {
		reactions = make([]ReactionPayload, len(m.Reactions))
		for i, r := range m.Reactions {
			reactions[i] = ReactionPayload{
				Emoji:   r.Emoji,
				Count:   r.Count,
				UserIDs: r.UserIDs,
			}
		}
	}

	return MessagePayload{
		ID:                m.ID,
		ChannelSeq:        m.ChannelSeq,
		ClientMsgID:       m.ClientMsgID,
		AuthorID:          m.AuthorID,
		AuthorDisplayName: m.AuthorDisplayName,
		AuthorAvatarURL:   m.AuthorAvatarURL,
		Body:              m.Body,
		CreatedAt:         m.CreatedAt,
		EditedAt:          m.EditedAt,
		IsPinned:          m.IsPinned,
		ReplyTo:           reply,
		Reactions:         reactions,
		Mine:              false,
	}
}

// EnvelopeFromMessageCreated maps a Kafka chat-events payload to a WebSocket frame.
func EnvelopeFromMessageCreated(ev chatEvents.MessageCreated) Envelope {
	return Envelope{
		Type:    TypeMessageCreated,
		Message: messagePayloadFromEvent(ev.Message),
	}
}

// MarshalMessageCreatedEvent JSON-encodes a chat-events record for WebSocket subscribers.
func MarshalMessageCreatedEvent(ev chatEvents.MessageCreated) ([]byte, error) {
	return MarshalEnvelope(EnvelopeFromMessageCreated(ev))
}

// MarshalMessageUpdatedEvent JSON-encodes a message.updated event for WebSocket subscribers.
func MarshalMessageUpdatedEvent(ev chatEvents.MessageUpdated) ([]byte, error) {
	return MarshalEnvelope(Envelope{
		Type:    TypeMessageUpdated,
		Message: messagePayloadFromEvent(ev.Message),
	})
}

// MarshalMessageDeletedEvent JSON-encodes a message.deleted event for WebSocket subscribers.
func MarshalMessageDeletedEvent(ev chatEvents.MessageDeleted) ([]byte, error) {
	return json.Marshal(struct {
		Type      string `json:"type"`
		MessageID string `json:"message_id"`
	}{
		Type:      TypeMessageDeleted,
		MessageID: ev.MessageID,
	})
}

// MarshalMessageReactionsUpdatedEvent JSON-encodes a message.reactions.updated event for WebSocket subscribers.
func MarshalMessageReactionsUpdatedEvent(ev chatEvents.MessageReactionsUpdated) ([]byte, error) {
	reactions := make([]ReactionPayload, len(ev.Reactions))
	for i, r := range ev.Reactions {
		reactions[i] = ReactionPayload{
			Emoji:   r.Emoji,
			Count:   r.Count,
			UserIDs: r.UserIDs,
		}
	}

	return json.Marshal(struct {
		Type      string            `json:"type"`
		MessageID string            `json:"message_id"`
		Reactions []ReactionPayload `json:"reactions"`
	}{
		Type:      TypeMessageReactionsUpdated,
		MessageID: ev.MessageID,
		Reactions: reactions,
	})
}
