package realtime_test

import (
	"encoding/json"
	"testing"
	"time"

	chatEvents "github.com/wichat/wichat/backend/wi-shared/chat-events"
	"github.com/wichat/wichat/backend/wi-shared/realtime"
)

func TestMarshalMessageCreatedEvent_matchesWSContract(t *testing.T) {
	createdAt, err := time.Parse(time.RFC3339Nano, "2026-01-15T10:00:00.000000000Z")
	if err != nil {
		t.Fatal(err)
	}

	ev := chatEvents.NewMessageCreated(
		"aaaaaaaaaaaaaaaaaaaaaaaa",
		"bbbbbbbbbbbbbbbbbbbbbbbb",
		"cccccccccccccccccccccccc",
		"dddddddddddddddddddddddd",
		"Author",
		"hello",
		createdAt,
	)

	raw, err := realtime.MarshalMessageCreatedEvent(ev)
	if err != nil {
		t.Fatal(err)
	}

	var decoded struct {
		Type    string `json:"type"`
		Message struct {
			ID   string `json:"id"`
			Mine bool   `json:"mine"`
		} `json:"message"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Type != realtime.TypeMessageCreated {
		t.Fatalf("type: got %q", decoded.Type)
	}
	if decoded.Message.ID != ev.Message.ID {
		t.Fatalf("message id: got %q", decoded.Message.ID)
	}
	if decoded.Message.Mine {
		t.Fatal("mine should be false for subscribers")
	}
}
