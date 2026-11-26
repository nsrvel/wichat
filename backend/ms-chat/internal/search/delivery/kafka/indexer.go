package kafka

import (
	"context"
	"errors"
	"fmt"
	"time"

	searchindexservice "github.com/wichat/wichat/backend/ms-chat/internal/search-index/service"
	wikafka "github.com/wichat/wichat/backend/wi-shared/infra/kafka"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// SearchIndexer consumes chat-events and indexes via search service.
type SearchIndexer interface {
	Run(ctx context.Context) error
}

type searchIndexer struct {
	consumer wikafka.Consumer
	svc      searchindexservice.Service
	log      logger.Logger
}

// NewSearchIndexer wires Kafka delivery for search indexing.
func NewSearchIndexer(consumer wikafka.Consumer, svc searchindexservice.Service, log logger.Logger) SearchIndexer {
	return searchIndexer{consumer: consumer, svc: svc, log: log}
}

func (s searchIndexer) Run(ctx context.Context) error {
	if err := s.svc.EnsureIndex(ctx); err != nil {
		return fmt.Errorf("search index: %w", err)
	}

	s.log.Info("search indexer running")

	for {
		in, err := s.consumer.Read(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || ctx.Err() != nil {
				return nil
			}
			return err
		}

		if err := s.handleWithRetry(ctx, in); err != nil {
			if errors.Is(err, context.Canceled) || ctx.Err() != nil {
				return nil
			}
			if searchindexservice.InvalidEventPayload(err) {
				s.log.Error("search indexer poison payload", "err", err)
			} else {
				s.log.Error("search indexer handle failed after retries; skipping commit", "err", err)
				continue
			}
		}

		if err := s.consumer.Commit(ctx, in); err != nil {
			s.log.Error("search indexer commit", "err", err)
		}
	}
}

func (s searchIndexer) handleWithRetry(ctx context.Context, in wikafka.Incoming) error {
	backoffs := searchindexservice.RetryDelays()
	var err error
	for attempt := 0; attempt <= len(backoffs); attempt++ {
		err = s.svc.IndexKafkaValue(ctx, in.Value)
		if err == nil || searchindexservice.InvalidEventPayload(err) {
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
