package realtime_test

import (
	"encoding/json"
	"testing"

	"github.com/wichat/wichat/backend/wi-shared/realtime"
)

func TestMarshalChannelsUpdated(t *testing.T) {
	t.Parallel()

	raw, err := realtime.MarshalChannelsUpdated("g1", realtime.ChannelsUpdateReasonDmCreated, "ch1")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var ev realtime.ChannelsUpdatedEvent
	if err := json.Unmarshal(raw, &ev); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ev.Type != realtime.TypeChannelsUpdated || ev.GroupID != "g1" || ev.Reason != realtime.ChannelsUpdateReasonDmCreated || ev.ChannelID != "ch1" {
		t.Fatalf("event %+v", ev)
	}
}