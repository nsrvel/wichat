package kafka

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// Producer publishes bytes to a Kafka topic.
type Producer interface {
	Publish(ctx context.Context, topic string, key, value []byte) error
	Close() error
}

type producer struct {
	writer *kafka.Writer
	log    logger.Logger
}

// NewProducer returns a Producer. Empty brokers yields a nil Producer and no error.
func NewProducer(cfg Config, log logger.Logger) (Producer, error) {
	if len(cfg.Brokers) == 0 {
		return nil, nil
	}

	w := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
	}
	return producer{writer: w, log: log}, nil
}

func (p producer) Publish(ctx context.Context, topic string, key, value []byte) error {
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return fmt.Errorf("kafka: topic required")
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
	})
	if err != nil {
		p.log.Error("kafka publish", "topic", topic, "err", err)
		return err
	}

	return nil
}

func (p producer) Close() error {
	return p.writer.Close()
}
