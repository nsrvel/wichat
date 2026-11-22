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
	UpdateChannelMessage(ctx context.Context, in *chatv1.UpdateChannelMessageRequest) (*chatv1.UpdateChannelMessageResponse, error)
	DeleteChannelMessage(ctx context.Context, in *chatv1.DeleteChannelMessageRequest) (*chatv1.DeleteChannelMessageResponse, error)
	ToggleMessageReaction(ctx context.Context, in *chatv1.ToggleMessageReactionRequest) (*chatv1.ToggleMessageReactionResponse, error)
	SetChannelMessagePinned(ctx context.Context, in *chatv1.SetChannelMessagePinnedRequest) (*chatv1.SetChannelMessagePinnedResponse, error)
	ListPinnedChannelMessages(ctx context.Context, in *chatv1.ListPinnedChannelMessagesRequest) (*chatv1.ListPinnedChannelMessagesResponse, error)
	UploadChannelAttachment(ctx context.Context, in *chatv1.UploadChannelAttachmentRequest) (*chatv1.UploadChannelAttachmentResponse, error)
	ResolveMessageChannelSeq(ctx context.Context, in *chatv1.ResolveMessageChannelSeqRequest) (*chatv1.ResolveMessageChannelSeqResponse, error)
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

func (c *client) UpdateChannelMessage(ctx context.Context, in *chatv1.UpdateChannelMessageRequest) (*chatv1.UpdateChannelMessageResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "chat.MessageService/UpdateChannelMessage", func(callCtx context.Context) (*chatv1.UpdateChannelMessageResponse, error) {
		return c.stub.UpdateChannelMessage(callCtx, in)
	})
}

func (c *client) DeleteChannelMessage(ctx context.Context, in *chatv1.DeleteChannelMessageRequest) (*chatv1.DeleteChannelMessageResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "chat.MessageService/DeleteChannelMessage", func(callCtx context.Context) (*chatv1.DeleteChannelMessageResponse, error) {
		return c.stub.DeleteChannelMessage(callCtx, in)
	})
}

func (c *client) ToggleMessageReaction(ctx context.Context, in *chatv1.ToggleMessageReactionRequest) (*chatv1.ToggleMessageReactionResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "chat.MessageService/ToggleMessageReaction", func(callCtx context.Context) (*chatv1.ToggleMessageReactionResponse, error) {
		return c.stub.ToggleMessageReaction(callCtx, in)
	})
}

func (c *client) SetChannelMessagePinned(ctx context.Context, in *chatv1.SetChannelMessagePinnedRequest) (*chatv1.SetChannelMessagePinnedResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "chat.MessageService/SetChannelMessagePinned", func(callCtx context.Context) (*chatv1.SetChannelMessagePinnedResponse, error) {
		return c.stub.SetChannelMessagePinned(callCtx, in)
	})
}

func (c *client) ListPinnedChannelMessages(ctx context.Context, in *chatv1.ListPinnedChannelMessagesRequest) (*chatv1.ListPinnedChannelMessagesResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "chat.MessageService/ListPinnedChannelMessages", func(callCtx context.Context) (*chatv1.ListPinnedChannelMessagesResponse, error) {
		return c.stub.ListPinnedChannelMessages(callCtx, in)
	})
}

func (c *client) UploadChannelAttachment(ctx context.Context, in *chatv1.UploadChannelAttachmentRequest) (*chatv1.UploadChannelAttachmentResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "chat.MessageService/UploadChannelAttachment", func(callCtx context.Context) (*chatv1.UploadChannelAttachmentResponse, error) {
		return c.stub.UploadChannelAttachment(callCtx, in, grpc.MaxCallSendMsgSize(32<<20))
	})
}

func (c *client) ResolveMessageChannelSeq(ctx context.Context, in *chatv1.ResolveMessageChannelSeqRequest) (*chatv1.ResolveMessageChannelSeqResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "chat.MessageService/ResolveMessageChannelSeq", func(callCtx context.Context) (*chatv1.ResolveMessageChannelSeqResponse, error) {
		return c.stub.ResolveMessageChannelSeq(callCtx, in)
	})
}
