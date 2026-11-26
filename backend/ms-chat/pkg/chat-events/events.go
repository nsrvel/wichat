package chatEvents

import (
	"time"

	events "github.com/wichat/wichat/backend/wi-shared/chat-events"
)

// EncodeMessageCreated builds Kafka JSON for message.created outbox rows.
func EncodeMessageCreated(groupPublicID, channelPublicID string, msg PublicMessageWire, createdAt time.Time) (OutboxInsert, error) {
	ev := events.MessageCreated{
		Type:      events.TypeMessageCreated,
		GroupID:   groupPublicID,
		ChannelID: channelPublicID,
		Message:   wireMessage(msg),
	}
	payload, err := events.Marshal(ev)
	if err != nil {
		return OutboxInsert{}, err
	}

	return encodeOutbox(channelPublicID, payload, events.TypeMessageCreated, createdAt)
}

// EncodeMessageUpdated builds Kafka JSON for message.updated outbox rows.
func EncodeMessageUpdated(groupPublicID, channelPublicID string, msg PublicMessageWire, at time.Time) (OutboxInsert, error) {
	ev := events.MessageUpdated{
		Type:      events.TypeMessageUpdated,
		GroupID:   groupPublicID,
		ChannelID: channelPublicID,
		Message:   wireMessage(msg),
	}
	payload, err := events.Marshal(ev)
	if err != nil {
		return OutboxInsert{}, err
	}

	return encodeOutbox(channelPublicID, payload, events.TypeMessageUpdated, at)
}

// EncodeMessageDeleted builds Kafka JSON for message.deleted outbox rows.
func EncodeMessageDeleted(groupPublicID, channelPublicID, messagePublicID string, at time.Time) (OutboxInsert, error) {
	ev := events.MessageDeleted{
		Type:      events.TypeMessageDeleted,
		GroupID:   groupPublicID,
		ChannelID: channelPublicID,
		MessageID: messagePublicID,
	}
	payload, err := events.Marshal(ev)
	if err != nil {
		return OutboxInsert{}, err
	}

	return encodeOutbox(channelPublicID, payload, events.TypeMessageDeleted, at)
}

// EncodeAttachmentCreated builds Kafka JSON for attachment.created outbox rows.
func EncodeAttachmentCreated(
	groupPublicID, channelPublicID, attachmentID, filename, contentType string,
	createdAt time.Time,
) (OutboxInsert, error) {
	ev := events.AttachmentCreated{
		Type:         events.TypeAttachmentCreated,
		GroupID:      groupPublicID,
		ChannelID:    channelPublicID,
		AttachmentID: attachmentID,
		Filename:     filename,
		ContentType:  contentType,
		CreatedAt:    createdAt.UTC().Format(time.RFC3339Nano),
	}
	payload, err := events.Marshal(ev)
	if err != nil {
		return OutboxInsert{}, err
	}

	return encodeOutbox(channelPublicID, payload, events.TypeAttachmentCreated, createdAt)
}

// EncodeMessageReactionsUpdated builds Kafka JSON for message.reactions.updated outbox rows.
func EncodeMessageReactionsUpdated(
	groupPublicID, channelPublicID, messagePublicID string,
	reactions []ReactionSummaryWire,
	at time.Time,
) (OutboxInsert, error) {
	ev := events.MessageReactionsUpdated{
		Type:      events.TypeMessageReactionsUpdated,
		GroupID:   groupPublicID,
		ChannelID: channelPublicID,
		MessageID: messagePublicID,
		Reactions: wireReactions(reactions),
	}
	payload, err := events.Marshal(ev)
	if err != nil {
		return OutboxInsert{}, err
	}

	return encodeOutbox(channelPublicID, payload, events.TypeMessageReactionsUpdated, at)
}
