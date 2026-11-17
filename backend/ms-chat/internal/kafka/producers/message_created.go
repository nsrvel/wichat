package producers

import (
	"context"
	"time"

	"github.com/wichat/wichat/backend/ms-chat/internal/models"
	chatEvents "github.com/wichat/wichat/backend/wi-shared/chat-events"
	wikafka "github.com/wichat/wichat/backend/wi-shared/infra/kafka"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// MessageCreatedPublisher emits chat-events after Postgres commit.
type MessageCreatedPublisher interface {
	PublishMessageCreated(ctx context.Context, groupPublicID, channelPublicID string, msg models.PublicMessage) error
	PublishMessageUpdated(ctx context.Context, groupPublicID, channelPublicID string, msg models.PublicMessage) error
	PublishMessageDeleted(ctx context.Context, groupPublicID, channelPublicID, messagePublicID string) error
	PublishMessageReactionsUpdated(ctx context.Context, groupPublicID, channelPublicID, messagePublicID string, reactions []models.MessageReactionSummary) error
}

type messageCreatedPublisher struct {
	producer wikafka.Producer
	topic    string
	log      logger.Logger
}

// NewMessageCreated builds a Kafka publisher for chat-events.
func NewMessageCreated(producer wikafka.Producer, topic string, log logger.Logger) MessageCreatedPublisher {
	return messageCreatedPublisher{producer: producer, topic: topic, log: log}
}

func toChatEventMessage(msg models.PublicMessage) chatEvents.Message {
	var reply *chatEvents.ReplyPreview
	if msg.ReplyTo != nil {
		reply = &chatEvents.ReplyPreview{
			ID:                msg.ReplyTo.ID,
			AuthorID:          msg.ReplyTo.AuthorID,
			AuthorDisplayName: msg.ReplyTo.AuthorDisplayName,
			BodyPreview:       msg.ReplyTo.BodyPreview,
		}
	}
	var reactions []chatEvents.ReactionSummary
	if len(msg.Reactions) > 0 {
		reactions = toChatEventReactions(msg.Reactions)
	}
	var editedAt string
	if msg.EditedAt != nil {
		editedAt = msg.EditedAt.UTC().Format(time.RFC3339Nano)
	}

	return chatEvents.Message{
		ID:                msg.ID,
		AuthorID:          msg.AuthorID,
		AuthorDisplayName: msg.AuthorDisplayName,
		AuthorAvatarURL:   msg.AuthorAvatarURL,
		Body:              msg.Body,
		CreatedAt:         msg.CreatedAt.UTC().Format(time.RFC3339Nano),
		EditedAt:          editedAt,
		IsPinned:          msg.IsPinned,
		ReplyTo:           reply,
		Reactions:         reactions,
	}
}

func toChatEventReactions(in []models.MessageReactionSummary) []chatEvents.ReactionSummary {
	out := make([]chatEvents.ReactionSummary, len(in))
	for i, r := range in {
		out[i] = chatEvents.ReactionSummary{
			Emoji:   r.Emoji,
			Count:   r.Count,
			UserIDs: r.UserIDs,
		}
	}

	return out
}

func (p messageCreatedPublisher) publish(ctx context.Context, op, channelPublicID string, ev any) error {
	raw, err := chatEvents.Marshal(ev)
	if err != nil {
		return err
	}
	if err := p.producer.Publish(ctx, p.topic, []byte(channelPublicID), raw); err != nil {
		p.log.Error("kafka publish "+op, "topic", p.topic, "err", err)

		return err
	}

	return nil
}

func (p messageCreatedPublisher) PublishMessageCreated(ctx context.Context, groupPublicID, channelPublicID string, msg models.PublicMessage) error {
	return p.publish(ctx, chatEvents.TypeMessageCreated, channelPublicID, chatEvents.MessageCreated{
		Type:      chatEvents.TypeMessageCreated,
		GroupID:   groupPublicID,
		ChannelID: channelPublicID,
		Message:   toChatEventMessage(msg),
	})
}

func (p messageCreatedPublisher) PublishMessageUpdated(ctx context.Context, groupPublicID, channelPublicID string, msg models.PublicMessage) error {
	return p.publish(ctx, chatEvents.TypeMessageUpdated, channelPublicID, chatEvents.MessageUpdated{
		Type:      chatEvents.TypeMessageUpdated,
		GroupID:   groupPublicID,
		ChannelID: channelPublicID,
		Message:   toChatEventMessage(msg),
	})
}

func (p messageCreatedPublisher) PublishMessageDeleted(ctx context.Context, groupPublicID, channelPublicID, messagePublicID string) error {
	return p.publish(ctx, chatEvents.TypeMessageDeleted, channelPublicID, chatEvents.MessageDeleted{
		Type:      chatEvents.TypeMessageDeleted,
		GroupID:   groupPublicID,
		ChannelID: channelPublicID,
		MessageID: messagePublicID,
	})
}

func (p messageCreatedPublisher) PublishMessageReactionsUpdated(
	ctx context.Context,
	groupPublicID, channelPublicID, messagePublicID string,
	reactions []models.MessageReactionSummary,
) error {
	return p.publish(ctx, chatEvents.TypeMessageReactionsUpdated, channelPublicID, chatEvents.MessageReactionsUpdated{
		Type:      chatEvents.TypeMessageReactionsUpdated,
		GroupID:   groupPublicID,
		ChannelID: channelPublicID,
		MessageID: messagePublicID,
		Reactions: toChatEventReactions(reactions),
	})
}
