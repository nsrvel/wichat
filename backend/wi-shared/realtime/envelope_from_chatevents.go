package realtime

import (
	"github.com/wichat/wichat/backend/wi-shared/chatevents"
)

// EnvelopeFromMessageCreated maps a Kafka chat-events payload to a WebSocket frame.
func EnvelopeFromMessageCreated(ev chatevents.MessageCreated) Envelope {
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
func MarshalMessageCreatedEvent(ev chatevents.MessageCreated) ([]byte, error) {
	return MarshalEnvelope(EnvelopeFromMessageCreated(ev))
}
