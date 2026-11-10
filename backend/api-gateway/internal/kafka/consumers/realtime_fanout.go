package consumers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/wichat/wichat/backend/api-gateway/internal/realtime/hub"
	"github.com/wichat/wichat/backend/wi-shared/chatevents"
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
	if strings.TrimSpace(head.Type) != chatevents.TypeMessageCreated {
		return nil
	}

	var ev chatevents.MessageCreated
	if err := json.Unmarshal(in.Value, &ev); err != nil {
		return err
	}
	if ev.GroupID == "" || ev.ChannelID == "" {
		return fmt.Errorf("missing group or channel in event")
	}

	payload, err := realtime.MarshalMessageCreatedEvent(ev)
	if err != nil {
		return err
	}

	r.hub.Deliver(ev.GroupID, ev.ChannelID, payload)

	return nil
}
