// Package search is the gRPC client for ms-chat SearchService.
package search

import (
	"context"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	chatv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/chat/v1"
	"google.golang.org/grpc"
)

// Client calls SearchService with shared timeouts and retries.
type Client interface {
	SearchMessages(ctx context.Context, in *chatv1.SearchMessagesRequest) (*chatv1.SearchMessagesResponse, error)
	SearchFiles(ctx context.Context, in *chatv1.SearchFilesRequest) (*chatv1.SearchFilesResponse, error)
	ListGroupMentions(ctx context.Context, in *chatv1.ListGroupMentionsRequest) (*chatv1.ListGroupMentionsResponse, error)
	MarkGroupMentionsRead(ctx context.Context, in *chatv1.MarkGroupMentionsReadRequest) (*chatv1.MarkGroupMentionsReadResponse, error)
	CountGroupMentionsUnread(ctx context.Context, in *chatv1.CountGroupMentionsUnreadRequest) (*chatv1.CountGroupMentionsUnreadResponse, error)
}

type client struct {
	stub chatv1.SearchServiceClient
	log  logger.Logger
	read utils.CallConfig
}

// New wraps an existing gRPC connection.
func New(conn *grpc.ClientConn, settings utils.Settings, log logger.Logger) Client {
	breaker := utils.NewBreaker(settings)
	read := utils.CallConfigFromSettings(settings, utils.RetryIdempotent)
	read.Breaker = breaker

	return &client{
		stub: chatv1.NewSearchServiceClient(conn),
		log:  log,
		read: read,
	}
}

func (c *client) SearchMessages(ctx context.Context, in *chatv1.SearchMessagesRequest) (*chatv1.SearchMessagesResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "chat.SearchService/SearchMessages", func(callCtx context.Context) (*chatv1.SearchMessagesResponse, error) {
		return c.stub.SearchMessages(callCtx, in)
	})
}

func (c *client) SearchFiles(ctx context.Context, in *chatv1.SearchFilesRequest) (*chatv1.SearchFilesResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "chat.SearchService/SearchFiles", func(callCtx context.Context) (*chatv1.SearchFilesResponse, error) {
		return c.stub.SearchFiles(callCtx, in)
	})
}

func (c *client) ListGroupMentions(ctx context.Context, in *chatv1.ListGroupMentionsRequest) (*chatv1.ListGroupMentionsResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "chat.SearchService/ListGroupMentions", func(callCtx context.Context) (*chatv1.ListGroupMentionsResponse, error) {
		return c.stub.ListGroupMentions(callCtx, in)
	})
}

func (c *client) MarkGroupMentionsRead(ctx context.Context, in *chatv1.MarkGroupMentionsReadRequest) (*chatv1.MarkGroupMentionsReadResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "chat.SearchService/MarkGroupMentionsRead", func(callCtx context.Context) (*chatv1.MarkGroupMentionsReadResponse, error) {
		return c.stub.MarkGroupMentionsRead(callCtx, in)
	})
}

func (c *client) CountGroupMentionsUnread(ctx context.Context, in *chatv1.CountGroupMentionsUnreadRequest) (*chatv1.CountGroupMentionsUnreadResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "chat.SearchService/CountGroupMentionsUnread", func(callCtx context.Context) (*chatv1.CountGroupMentionsUnreadResponse, error) {
		return c.stub.CountGroupMentionsUnread(callCtx, in)
	})
}
