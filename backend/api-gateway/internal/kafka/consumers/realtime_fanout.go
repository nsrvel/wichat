package consumers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

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

var errInvalidFanoutPayload = errors.New("invalid fanout payload")

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

		if err := r.handleWithRetry(ctx, in); err != nil {
			if errors.Is(err, context.Canceled) || ctx.Err() != nil {
				return nil
			}
			if errors.Is(err, errInvalidFanoutPayload) {
				r.log.Error("realtime fanout poison payload", "err", err)
			} else {
				r.log.Error("realtime fanout handle failed after retries; skipping commit", "err", err)
				continue
			}
		}

		if err := r.consumer.Commit(ctx, in); err != nil {
			r.log.Error("realtime fanout commit", "err", err)
		}
	}
}

func fanoutRetryDelays() []time.Duration {
	return []time.Duration{50 * time.Millisecond, 200 * time.Millisecond, 500 * time.Millisecond}
}

func (r realtimeFanout) handleWithRetry(ctx context.Context, in wikafka.Incoming) error {
	backoffs := fanoutRetryDelays()
	var err error
	for attempt := 0; attempt <= len(backoffs); attempt++ {
		err = r.handle(in)
		if err == nil || errors.Is(err, errInvalidFanoutPayload) {
			return err
		}
		if attempt == len(backoffs) {
			break
		}
		timer := time.NewTimer(backoffs[attempt])
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}

	return err
}

func (r realtimeFanout) handle(in wikafka.Incoming) error {
	value, err := chatEvents.NormalizePayload(in.Value)
	if err != nil {
		return fmt.Errorf("%w: %w", errInvalidFanoutPayload, err)
	}

	var head struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(value, &head); err != nil {
		return fmt.Errorf("%w: %w", errInvalidFanoutPayload, err)
	}

	switch strings.TrimSpace(head.Type) {
	case chatEvents.TypeMessageCreated:
		var ev chatEvents.MessageCreated
		if err := json.Unmarshal(value, &ev); err != nil {
			return fmt.Errorf("%w: %w", errInvalidFanoutPayload, err)
		}
		return r.deliver(ev.GroupID, ev.ChannelID, func() ([]byte, error) {
			return realtime.MarshalMessageCreatedEvent(ev)
		})
	case chatEvents.TypeMessageUpdated:
		var ev chatEvents.MessageUpdated
		if err := json.Unmarshal(value, &ev); err != nil {
			return fmt.Errorf("%w: %w", errInvalidFanoutPayload, err)
		}
		return r.deliver(ev.GroupID, ev.ChannelID, func() ([]byte, error) {
			return realtime.MarshalMessageUpdatedEvent(ev)
		})
	case chatEvents.TypeMessageDeleted:
		var ev chatEvents.MessageDeleted
		if err := json.Unmarshal(value, &ev); err != nil {
			return fmt.Errorf("%w: %w", errInvalidFanoutPayload, err)
		}
		return r.deliver(ev.GroupID, ev.ChannelID, func() ([]byte, error) {
			return realtime.MarshalMessageDeletedEvent(ev)
		})
	case chatEvents.TypeMessageReactionsUpdated:
		var ev chatEvents.MessageReactionsUpdated
		if err := json.Unmarshal(value, &ev); err != nil {
			return fmt.Errorf("%w: %w", errInvalidFanoutPayload, err)
		}
		return r.deliver(ev.GroupID, ev.ChannelID, func() ([]byte, error) {
			return realtime.MarshalMessageReactionsUpdatedEvent(ev)
		})
	default:
		return nil
	}
}

func (r realtimeFanout) deliver(groupID, channelID string, marshalFn func() ([]byte, error)) error {
	if groupID == "" || channelID == "" {
		return fmt.Errorf("missing group or channel in event")
	}
	payload, err := marshalFn()
	if err != nil {
		return err
	}

	return r.hub.Deliver(groupID, channelID, payload)
}
