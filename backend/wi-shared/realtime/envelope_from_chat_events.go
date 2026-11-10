package realtime

import (
	chatEvents "github.com/wichat/wichat/backend/wi-shared/chat-events"
)

// EnvelopeFromMessageCreated maps a Kafka chat-events payload to a WebSocket frame.
func EnvelopeFromMessageCreated(ev chatEvents.MessageCreated) Envelope {
	return Envelope{
		Type: TypeMessageCreated,
		Message: MessagePayload{
			ID:                ev.Message.ID,
			AuthorID:          ev.Message.AuthorID,
			AuthorDisplayName: ev.Message.AuthorDisplayName,
			Body:              ev.Message.Body,
			CreatedAt:         ev.Message.CreatedAt,
			Mine:              false,
		},
	}
}

// MarshalMessageCreatedEvent JSON-encodes a chat-events record for WebSocket subscribers.
func MarshalMessageCreatedEvent(ev chatEvents.MessageCreated) ([]byte, error) {
	return MarshalEnvelope(EnvelopeFromMessageCreated(ev))
}
