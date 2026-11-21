package service

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
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// Service owns search indexing use cases.
type Service interface {
	EnsureIndex(ctx context.Context) error
	IndexKafkaValue(ctx context.Context, value []byte) error
}

type service struct {
	repo searchrepo.EsRepository
	log  logger.Logger
}

// NewService builds search indexing use cases.
func NewService(repo searchrepo.EsRepository, log logger.Logger) Service {
	return service{repo: repo, log: log}
}

var errInvalidEventPayload = errors.New("invalid event payload")

func (s service) EnsureIndex(ctx context.Context) error {
	if s.repo == nil {
		return nil
	}

	return s.repo.EnsureIndex(ctx)
}

func (s service) IndexKafkaValue(ctx context.Context, value []byte) error {
	if s.repo == nil {
		return nil
	}

	var head struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(value, &head); err != nil {
		return fmt.Errorf("%w: %w", errInvalidEventPayload, err)
	}

	switch strings.TrimSpace(head.Type) {
	case chatEvents.TypeMessageCreated:
		var ev chatEvents.MessageCreated
		if err := json.Unmarshal(value, &ev); err != nil {
			return fmt.Errorf("%w: %w", errInvalidEventPayload, err)
		}

		return s.indexMessageDoc(ctx, ev.GroupID, ev.ChannelID, ev.Message)
	case chatEvents.TypeMessageUpdated:
		var ev chatEvents.MessageUpdated
		if err := json.Unmarshal(value, &ev); err != nil {
			return fmt.Errorf("%w: %w", errInvalidEventPayload, err)
		}

		return s.indexMessageDoc(ctx, ev.GroupID, ev.ChannelID, ev.Message)
	case chatEvents.TypeMessageDeleted:
		var ev chatEvents.MessageDeleted
		if err := json.Unmarshal(value, &ev); err != nil {
			return fmt.Errorf("%w: %w", errInvalidEventPayload, err)
		}
		if ev.MessageID == "" {
			return fmt.Errorf("missing message id in event")
		}

		return s.repo.DeleteMessage(ctx, ev.MessageID)
	default:
		return nil
	}
}

func (s service) indexMessageDoc(ctx context.Context, groupID, channelID string, msg chatEvents.Message) error {
	doc := elasticsearch.MessageDoc{
		MessageID:         msg.ID,
		GroupID:           groupID,
		ChannelID:         channelID,
		AuthorID:          msg.AuthorID,
		AuthorDisplayName: msg.AuthorDisplayName,
		Body:              msg.Body,
		CreatedAt:         msg.CreatedAt,
	}
	if doc.MessageID == "" {
		return fmt.Errorf("missing message id in event")
	}

	return s.repo.IndexMessageCreated(ctx, doc)
}

// InvalidEventPayload reports poison Kafka JSON for delivery to skip commit.
func InvalidEventPayload(err error) bool {
	return errors.Is(err, errInvalidEventPayload)
}

// RetryDelays are backoff steps for transient indexing failures.
func RetryDelays() []time.Duration {
	return []time.Duration{100 * time.Millisecond, 500 * time.Millisecond, 1500 * time.Millisecond}
}
