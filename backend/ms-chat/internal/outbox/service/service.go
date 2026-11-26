package service

import (
	"context"
	"time"

	outboxrepo "github.com/wichat/wichat/backend/ms-chat/internal/outbox/repository"
	"github.com/wichat/wichat/backend/ms-chat/pkg/constants"
	chatEvents "github.com/wichat/wichat/backend/wi-shared/chat-events"
	wikafka "github.com/wichat/wichat/backend/wi-shared/infra/kafka"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// Service runs the outbox relay loop (poll → Kafka).
type Service interface {
	RunRelay(ctx context.Context) error
}

type service struct {
	outbox   outboxrepo.PgRepository
	producer wikafka.Producer
	topic    string
	interval time.Duration
	log      logger.Logger
}

// NewService wires outbox polling to Kafka publish.
func NewService(outbox outboxrepo.PgRepository, producer wikafka.Producer, topic string, log logger.Logger) Service {
	return service{
		outbox:   outbox,
		producer: producer,
		topic:    topic,
		interval: constants.OutboxPollInterval,
		log:      log,
	}
}

func (s service) RunRelay(ctx context.Context) error {
	s.log.Info("outbox relay running", "topic", s.topic)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			s.flushBatch(ctx)
		}
	}
}

func (s service) flushBatch(ctx context.Context) {
	rows, err := s.outbox.ClaimPending(ctx, constants.OutboxPublishBatchSize)
	if err != nil {
		if ctx.Err() != nil {
			return
		}
		s.log.Error("outbox claim", "err", err)

		return
	}
	if len(rows) == 0 {
		return
	}

	for _, row := range rows {
		payload, normErr := chatEvents.NormalizePayload(row.Payload)
		if normErr != nil {
			s.log.Error("outbox payload normalize", "outbox_id", row.ID, "event_type", row.EventType, "err", normErr)
			if recordErr := s.outbox.RecordPublishFailure(ctx, row.ID, normErr.Error()); recordErr != nil && ctx.Err() == nil {
				s.log.Error("outbox record failure", "outbox_id", row.ID, "err", recordErr)
			}
			continue
		}

		pubCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		err := s.producer.Publish(pubCtx, s.topic, []byte(row.PartitionKey), payload)
		cancel()
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			s.log.Error("outbox publish", "outbox_id", row.ID, "event_type", row.EventType, "err", err)
			if recordErr := s.outbox.RecordPublishFailure(ctx, row.ID, err.Error()); recordErr != nil && ctx.Err() == nil {
				s.log.Error("outbox record failure", "outbox_id", row.ID, "err", recordErr)
			}

			continue
		}

		if err := s.outbox.MarkPublished(ctx, row.ID, time.Now().UTC()); err != nil && ctx.Err() == nil {
			s.log.Error("outbox mark published", "outbox_id", row.ID, "err", err)
		}
	}
}
