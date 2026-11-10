package realtime_test

import (
	"testing"
	"time"

	"github.com/wichat/wichat/backend/wi-shared/realtime"
)

func TestChannelKeyAndDirectEnvelope(t *testing.T) {
	t.Parallel()

	if got := realtime.ChannelKey("g1", "c1"); got != "chat:g:g1:c:c1" {
		t.Fatalf("key %q", got)
	}

	createdAt := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	env := realtime.NewMessageCreated("id", "u", "Ada", "hi", createdAt)
	raw, err := realtime.MarshalEnvelope(env)
	if err != nil || len(raw) == 0 {
		t.Fatalf("marshal: %v", err)
	}
}
