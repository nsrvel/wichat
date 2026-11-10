// Package message is the gRPC client for ms-chat MessageService.
package message

import (
	"context"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	chatv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/chat/v1"
	"google.golang.org/grpc"
)

// Client calls MessageService with shared timeouts and retries.
type Client interface {
	ListChannelMessages(ctx context.Context, in *chatv1.ListChannelMessagesRequest) (*chatv1.ListChannelMessagesResponse, error)
	CreateChannelMessage(ctx context.Context, in *chatv1.CreateChannelMessageRequest) (*chatv1.CreateChannelMessageResponse, error)
}

type client struct {
	stub chatv1.MessageServiceClient
	log  logger.Logger
	read utils.CallConfig
	none utils.CallConfig
}

// New wraps an existing gRPC connection.
func New(conn *grpc.ClientConn, settings utils.Settings, log logger.Logger) Client {
	breaker := utils.NewBreaker(settings)
	read := utils.CallConfigFromSettings(settings, utils.RetryIdempotent)
	read.Breaker = breaker
	none := utils.CallConfigFromSettings(settings, utils.RetryNone)
	none.Breaker = breaker

	return &client{
		stub: chatv1.NewMessageServiceClient(conn),
		log:  log,
		read: read,
		none: none,
	}
}

func (c *client) ListChannelMessages(ctx context.Context, in *chatv1.ListChannelMessagesRequest) (*chatv1.ListChannelMessagesResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "chat.MessageService/ListChannelMessages", func(callCtx context.Context) (*chatv1.ListChannelMessagesResponse, error) {
		return c.stub.ListChannelMessages(callCtx, in)
	})
}

func (c *client) CreateChannelMessage(ctx context.Context, in *chatv1.CreateChannelMessageRequest) (*chatv1.CreateChannelMessageResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "chat.MessageService/CreateChannelMessage", func(callCtx context.Context) (*chatv1.CreateChannelMessageResponse, error) {
		return c.stub.CreateChannelMessage(callCtx, in)
	})
}
