package consumers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/wichat/wichat/backend/api-gateway/internal/realtime/hub"
	chatEvents "github.com/wichat/wichat/backend/wi-shared/chat-events"
	wikafka "github.com/wichat/wichat/backend/wi-shared/infra/kafka"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/realtime"
)

// RealtimeFanout consumes chat-events and pushes to local WebSocket subscribers.
type RealtimeFanout interface {
	Run(ctx context.Context) error
}

type realtimeFanout struct {
	consumer wikafka.Consumer
	hub      *hub.Hub
	log      logger.Logger
}

// NewRealtimeFanout wires the gateway Kafka → WebSocket fanout.
func NewRealtimeFanout(consumer wikafka.Consumer, h *hub.Hub, log logger.Logger) RealtimeFanout {
	return realtimeFanout{consumer: consumer, hub: h, log: log}
}

func (r realtimeFanout) Run(ctx context.Context) error {
	r.log.Info("realtime fanout running")

	for {
		in, err := r.consumer.Read(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || ctx.Err() != nil {
				return nil
			}
			return err
		}

		if err := r.handle(in); err != nil {
			r.log.Error("realtime fanout handle", "err", err)
		}

		if err := r.consumer.Commit(ctx, in); err != nil {
			r.log.Error("realtime fanout commit", "err", err)
		}
	}
}

func (r realtimeFanout) handle(in wikafka.Incoming) error {
	var head struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(in.Value, &head); err != nil {
		return err
	}

	switch strings.TrimSpace(head.Type) {
	case chatEvents.TypeMessageCreated:
		var ev chatEvents.MessageCreated
		if err := json.Unmarshal(in.Value, &ev); err != nil {
			return err
		}
		return r.deliver(ev.GroupID, ev.ChannelID, func() ([]byte, error) {
			return realtime.MarshalMessageCreatedEvent(ev)
		})
	case chatEvents.TypeMessageUpdated:
		var ev chatEvents.MessageUpdated
		if err := json.Unmarshal(in.Value, &ev); err != nil {
			return err
		}
		return r.deliver(ev.GroupID, ev.ChannelID, func() ([]byte, error) {
			return realtime.MarshalMessageUpdatedEvent(ev)
		})
	case chatEvents.TypeMessageDeleted:
		var ev chatEvents.MessageDeleted
		if err := json.Unmarshal(in.Value, &ev); err != nil {
			return err
		}
		return r.deliver(ev.GroupID, ev.ChannelID, func() ([]byte, error) {
			return realtime.MarshalMessageDeletedEvent(ev)
		})
	case chatEvents.TypeMessageReactionsUpdated:
		var ev chatEvents.MessageReactionsUpdated
		if err := json.Unmarshal(in.Value, &ev); err != nil {
			return err
		}
		return r.deliver(ev.GroupID, ev.ChannelID, func() ([]byte, error) {
			return realtime.MarshalMessageReactionsUpdatedEvent(ev)
		})
	case chatEvents.TypeVoiceOccupancyUpdated:
		var ev chatEvents.VoiceOccupancyUpdated
		if err := json.Unmarshal(in.Value, &ev); err != nil {
			return err
		}
		return r.deliverGroupAndChannel(ev.GroupID, ev.ChannelID, func() ([]byte, error) {
			return realtime.MarshalVoiceOccupancyUpdatedEvent(ev)
		})
	case chatEvents.TypeVoiceRecordingUpdated:
		var ev chatEvents.VoiceRecordingUpdated
		if err := json.Unmarshal(in.Value, &ev); err != nil {
			return err
		}
		return r.deliverGroupAndChannel(ev.GroupID, ev.ChannelID, func() ([]byte, error) {
			return realtime.MarshalVoiceRecordingUpdatedEvent(ev)
		})
	case chatEvents.TypeVoiceModerationApplied:
		var ev chatEvents.VoiceModerationApplied
		if err := json.Unmarshal(in.Value, &ev); err != nil {
			return err
		}
		return r.deliverGroupAndChannel(ev.GroupID, ev.ChannelID, func() ([]byte, error) {
			return realtime.MarshalVoiceModerationAppliedEvent(ev)
		})
	case chatEvents.TypeVoiceIdleWarning:
		var ev chatEvents.VoiceIdleWarning
		if err := json.Unmarshal(in.Value, &ev); err != nil {
			return err
		}
		return r.deliverGroupAndChannel(ev.GroupID, ev.ChannelID, func() ([]byte, error) {
			return realtime.MarshalVoiceIdleWarningEvent(ev)
		})
	default:
		return nil
	}
}

func (r realtimeFanout) deliverGroupAndChannel(groupID, channelID string, marshalFn func() ([]byte, error)) error {
	payload, err := marshalFn()
	if err != nil {
		return err
	}
	if groupID != "" {
		r.hub.DeliverGroup(groupID, payload)
	}
	if groupID != "" && channelID != "" {
		r.hub.Deliver(groupID, channelID, payload)
	}
	return nil
}

func (r realtimeFanout) deliver(groupID, channelID string, marshalFn func() ([]byte, error)) error {
	if groupID == "" || channelID == "" {
		return fmt.Errorf("missing group or channel in event")
	}
	payload, err := marshalFn()
	if err != nil {
		return err
	}
	r.hub.Deliver(groupID, channelID, payload)

	return nil
}
