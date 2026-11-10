package producers

import (
	"context"

	"github.com/wichat/wichat/backend/ms-chat/internal/models"
	chatEvents "github.com/wichat/wichat/backend/wi-shared/chat-events"
	wikafka "github.com/wichat/wichat/backend/wi-shared/infra/kafka"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// MessageCreatedPublisher emits chat-events after Postgres commit.
type MessageCreatedPublisher interface {
	PublishMessageCreated(ctx context.Context, groupPublicID, channelPublicID string, msg models.PublicMessage) error
}

type messageCreatedPublisher struct {
	producer wikafka.Producer
	topic    string
	log      logger.Logger
}

// NewMessageCreated builds a Kafka publisher for message.created events.
func NewMessageCreated(producer wikafka.Producer, topic string, log logger.Logger) MessageCreatedPublisher {
	return messageCreatedPublisher{producer: producer, topic: topic, log: log}
}

func (p messageCreatedPublisher) PublishMessageCreated(ctx context.Context, groupPublicID, channelPublicID string, msg models.PublicMessage) error {
	ev := chatEvents.NewMessageCreated(
		groupPublicID, channelPublicID,
		msg.ID, msg.AuthorID, msg.AuthorDisplayName, msg.Body, msg.CreatedAt,
	)
	raw, err := chatEvents.Marshal(ev)
	if err != nil {
		return err
	}

	if err := p.producer.Publish(ctx, p.topic, []byte(msg.ID), raw); err != nil {
		p.log.Error("kafka publish message.created", "topic", p.topic, "err", err)
		return err
	}

	return nil
}
