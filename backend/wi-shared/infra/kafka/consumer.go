package kafka

import (
	"context"
	"fmt"
	"strings"

	"github.com/segmentio/kafka-go"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// Incoming is one consumed record.
type Incoming struct {
	Topic string
	Key   []byte
	Value []byte
	raw   kafka.Message
}

// Consumer reads and commits Kafka records.
type Consumer interface {
	Read(ctx context.Context) (Incoming, error)
	Commit(ctx context.Context, msg Incoming) error
	Close() error
}

type consumer struct {
	reader *kafka.Reader
	log    logger.Logger
}

// NewConsumer returns a Consumer. Empty brokers or topic yields nil.
func NewConsumer(cfg Config, log logger.Logger) (Consumer, error) {
	if len(cfg.Brokers) == 0 {
		return nil, nil
	}
	topic := strings.TrimSpace(cfg.Topic)
	if topic == "" {
		return nil, fmt.Errorf("kafka: topic required for consumer")
	}
	groupID := strings.TrimSpace(cfg.GroupID)
	if groupID == "" {
		return nil, fmt.Errorf("kafka: group id required for consumer")
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.Brokers,
		Topic:       topic,
		GroupID:     groupID,
		MinBytes:    1,
		MaxBytes:    1e6,
		StartOffset: kafka.LastOffset,
	})
	return consumer{reader: r, log: log}, nil
}

func (c consumer) Read(ctx context.Context) (Incoming, error) {
	m, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return Incoming{}, err
	}

	return Incoming{
		Topic: m.Topic,
		Key:   m.Key,
		Value: m.Value,
		raw:   m,
	}, nil
}

func (c consumer) Commit(ctx context.Context, msg Incoming) error {
	return c.reader.CommitMessages(ctx, msg.raw)
}

func (c consumer) Close() error {
	return c.reader.Close()
}
