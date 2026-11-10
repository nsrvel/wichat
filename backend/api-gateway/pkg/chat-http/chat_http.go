// Package chatHttp maps message JSON bodies and protobuf summaries for HTTP responses.
package chatHttp

import (
	chatv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/chat/v1"
)

// MessageBody is one channel message in JSON.
type MessageBody struct {
	ID                string `json:"id"`
	AuthorID          string `json:"author_id"`
	AuthorDisplayName string `json:"author_display_name"`
	Body              string `json:"body"`
	CreatedAt         string `json:"created_at"`
	Mine              bool   `json:"mine"`
}

// MessagesListBody wraps a channel message list.
type MessagesListBody struct {
	Messages []MessageBody `json:"messages"`
}

// CreateMessageBody is the JSON body for POST channel messages.
type CreateMessageBody struct {
	Body string `json:"body"`
}

// CreateMessageResponseBody is the created message payload.
type CreateMessageResponseBody struct {
	Message MessageBody `json:"message"`
}

// MessagesFromPB maps message summary protos to JSON bodies.
func MessagesFromPB(messages []*chatv1.MessageSummary) []MessageBody {
	out := make([]MessageBody, len(messages))
	for i, msg := range messages {
		out[i] = MessageFromPB(msg)
	}

	return out
}

// MessageFromPB maps one message summary proto to JSON.
func MessageFromPB(msg *chatv1.MessageSummary) MessageBody {
	if msg == nil {
		return MessageBody{}
	}

	return MessageBody{
		ID:                msg.GetId(),
		AuthorID:          msg.GetAuthorId(),
		AuthorDisplayName: msg.GetAuthorDisplayName(),
		Body:              msg.GetBody(),
		CreatedAt:         msg.GetCreatedAt(),
		Mine:              msg.GetMine(),
	}
}
