package producers

import (
	"context"
	"encoding/json"

	chatEvents "github.com/wichat/wichat/backend/wi-shared/chat-events"
	wikafka "github.com/wichat/wichat/backend/wi-shared/infra/kafka"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// VoiceEventsPublisher emits voice.* events to Kafka chat-events.
type VoiceEventsPublisher interface {
	PublishOccupancyUpdated(ctx context.Context, ev chatEvents.VoiceOccupancyUpdated) error
	PublishRecordingUpdated(ctx context.Context, ev chatEvents.VoiceRecordingUpdated) error
	PublishModerationApplied(ctx context.Context, ev chatEvents.VoiceModerationApplied) error
	PublishIdleWarning(ctx context.Context, ev chatEvents.VoiceIdleWarning) error
}

type voiceEventsPublisher struct {
	producer wikafka.Producer
	topic    string
	log      logger.Logger
}

// NewVoiceEventsPublisher constructs a Kafka publisher for voice.* events.
func NewVoiceEventsPublisher(producer wikafka.Producer, topic string, log logger.Logger) VoiceEventsPublisher {
	return voiceEventsPublisher{producer: producer, topic: topic, log: log}
}

func (p voiceEventsPublisher) publish(ctx context.Context, key string, payload any) error {
	if p.producer == nil {
		return nil
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if err := p.producer.Publish(ctx, p.topic, []byte(key), raw); err != nil {
		if p.log != nil {
			p.log.Warn("voice event publish failed", "topic", p.topic, "key", key, "err", err)
		}
		return err
	}
	return nil
}

func (p voiceEventsPublisher) PublishOccupancyUpdated(ctx context.Context, ev chatEvents.VoiceOccupancyUpdated) error {
	ev.Type = chatEvents.TypeVoiceOccupancyUpdated
	return p.publish(ctx, ev.GroupID, ev)
}

func (p voiceEventsPublisher) PublishRecordingUpdated(ctx context.Context, ev chatEvents.VoiceRecordingUpdated) error {
	ev.Type = chatEvents.TypeVoiceRecordingUpdated
	return p.publish(ctx, ev.GroupID, ev)
}

func (p voiceEventsPublisher) PublishModerationApplied(ctx context.Context, ev chatEvents.VoiceModerationApplied) error {
	ev.Type = chatEvents.TypeVoiceModerationApplied
	return p.publish(ctx, ev.GroupID, ev)
}

func (p voiceEventsPublisher) PublishIdleWarning(ctx context.Context, ev chatEvents.VoiceIdleWarning) error {
	ev.Type = chatEvents.TypeVoiceIdleWarning
	return p.publish(ctx, ev.GroupID, ev)
}
