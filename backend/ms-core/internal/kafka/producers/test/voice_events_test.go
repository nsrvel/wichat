package producers_test

import (
	"context"
	"testing"

	"github.com/wichat/wichat/backend/ms-core/internal/kafka/producers"
	chatEvents "github.com/wichat/wichat/backend/wi-shared/chat-events"
)

func TestVoiceEventsPublisherNilProducerNoop(t *testing.T) {
	pub := producers.NewVoiceEventsPublisher(nil, "chat-events", nil)
	ctx := context.Background()

	if err := pub.PublishOccupancyUpdated(ctx, chatEvents.VoiceOccupancyUpdated{GroupID: "g1", ChannelID: "c1"}); err != nil {
		t.Fatalf("PublishOccupancyUpdated() error = %v", err)
	}
	if err := pub.PublishRecordingUpdated(ctx, chatEvents.VoiceRecordingUpdated{GroupID: "g1", ChannelID: "c1"}); err != nil {
		t.Fatalf("PublishRecordingUpdated() error = %v", err)
	}
	if err := pub.PublishModerationApplied(ctx, chatEvents.VoiceModerationApplied{GroupID: "g1", ChannelID: "c1"}); err != nil {
		t.Fatalf("PublishModerationApplied() error = %v", err)
	}
	if err := pub.PublishIdleWarning(ctx, chatEvents.VoiceIdleWarning{GroupID: "g1", ChannelID: "c1"}); err != nil {
		t.Fatalf("PublishIdleWarning() error = %v", err)
	}
}
