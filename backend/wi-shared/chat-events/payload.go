package chatEvents

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
)

var errInvalidEventPayloadShape = errors.New("invalid chat-events payload shape")

// NormalizePayload returns JSON object bytes for a Kafka record or outbox row value.
// Legacy ms-chat outbox stored []byte via bun jsonb as a base64 string wrapper.
func NormalizePayload(raw []byte) ([]byte, error) {
	raw = trimBytes(raw)
	if len(raw) == 0 {
		return nil, fmt.Errorf("%w: empty", errInvalidEventPayloadShape)
	}
	if raw[0] == '{' {
		if !json.Valid(raw) {
			return nil, fmt.Errorf("%w: invalid object json", errInvalidEventPayloadShape)
		}
		return raw, nil
	}
	if raw[0] != '"' {
		return nil, fmt.Errorf("%w: unexpected prefix", errInvalidEventPayloadShape)
	}
	var wrapped string
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return nil, fmt.Errorf("%w: %w", errInvalidEventPayloadShape, err)
	}
	if decoded, err := base64.StdEncoding.DecodeString(wrapped); err == nil {
		decoded = trimBytes(decoded)
		if len(decoded) > 0 && decoded[0] == '{' && json.Valid(decoded) {
			return decoded, nil
		}
	}
	inner := trimBytes([]byte(wrapped))
	if len(inner) > 0 && inner[0] == '{' && json.Valid(inner) {
		return inner, nil
	}
	return nil, fmt.Errorf("%w: cannot unwrap legacy payload", errInvalidEventPayloadShape)
}

func trimBytes(b []byte) []byte {
	for len(b) > 0 && (b[0] == ' ' || b[0] == '\n' || b[0] == '\r' || b[0] == '\t') {
		b = b[1:]
	}
	for len(b) > 0 {
		c := b[len(b)-1]
		if c != ' ' && c != '\n' && c != '\r' && c != '\t' {
			break
		}
		b = b[:len(b)-1]
	}
	return b
}
