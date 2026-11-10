package chatevents_test

import (
	"testing"
	"time"

	"github.com/wichat/wichat/backend/wi-shared/chatevents"
)

func TestNewMessageCreatedJSON(t *testing.T) {
	at := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	ev := chatevents.NewMessageCreated("g1", "c1", "m1", "u1", "Alice", "hi", at)
	raw, err := chatevents.Marshal(ev)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) == 0 {
		t.Fatal("expected json")
	}
	if ev.Type != chatevents.TypeMessageCreated {
		t.Fatalf("type: %s", ev.Type)
	}
}
