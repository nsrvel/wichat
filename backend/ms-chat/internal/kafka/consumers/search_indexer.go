package consumers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	searchrepo "github.com/wichat/wichat/backend/ms-chat/internal/search/repository"
	chatEvents "github.com/wichat/wichat/backend/wi-shared/chat-events"
	"github.com/wichat/wichat/backend/wi-shared/infra/elasticsearch"
	wikafka "github.com/wichat/wichat/backend/wi-shared/infra/kafka"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// SearchIndexer runs the Kafka chat-events → Elasticsearch indexer.
type SearchIndexer interface {
	Run(ctx context.Context) error
}

type searchIndexer struct {
	consumer wikafka.Consumer
	repo     searchrepo.EsRepository
	log      logger.Logger
}

// NewSearchIndexer wires the search indexer consumer.
func NewSearchIndexer(consumer wikafka.Consumer, repo searchrepo.EsRepository, log logger.Logger) SearchIndexer {
	return searchIndexer{consumer: consumer, repo: repo, log: log}
}

var errInvalidEventPayload = errors.New("invalid event payload")

func (s searchIndexer) Run(ctx context.Context) error {
	if err := s.repo.EnsureIndex(ctx); err != nil {
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
			if errors.Is(err, errInvalidEventPayload) {
				// Poison-pill JSON: log and commit offset so the partition is not wedged.
				s.log.Error("search indexer poison payload", "err", err)
			} else {
				// Transient ES failure: skip commit so Kafka redelivers instead of dropping the event.
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
	backoffs := []time.Duration{100 * time.Millisecond, 500 * time.Millisecond, 1500 * time.Millisecond}
	var err error
	for attempt := 0; attempt <= len(backoffs); attempt++ {
		err = s.handle(ctx, in)
		if err == nil || errors.Is(err, errInvalidEventPayload) {
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

func (s searchIndexer) handle(ctx context.Context, in wikafka.Incoming) error {
	var head struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(in.Value, &head); err != nil {
		return fmt.Errorf("%w: %w", errInvalidEventPayload, err)
	}
	if strings.TrimSpace(head.Type) != chatEvents.TypeMessageCreated {
		return nil
	}

	var ev chatEvents.MessageCreated
	if err := json.Unmarshal(in.Value, &ev); err != nil {
		return fmt.Errorf("%w: %w", errInvalidEventPayload, err)
	}

	doc := elasticsearch.MessageDoc{
		MessageID:         ev.Message.ID,
		GroupID:           ev.GroupID,
		ChannelID:         ev.ChannelID,
		AuthorID:          ev.Message.AuthorID,
		AuthorDisplayName: ev.Message.AuthorDisplayName,
		Body:              ev.Message.Body,
		CreatedAt:         ev.Message.CreatedAt,
	}
	if doc.MessageID == "" {
		return fmt.Errorf("missing message id in event")
	}

	return s.repo.IndexMessageCreated(ctx, doc)
}
