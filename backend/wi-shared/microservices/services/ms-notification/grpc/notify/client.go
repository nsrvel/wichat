// Package notify is the gRPC client for ms-notification NotifyService.
package notify

import (
	"context"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	notificationv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/notification/v1"
	"google.golang.org/grpc"
)

// Client calls NotifyService with shared timeouts and retries.
type Client interface {
	Send(ctx context.Context, in *notificationv1.SendRequest) (*notificationv1.SendResponse, error)
}

type client struct {
	stub userv1Notify
	log  logger.Logger
	none utils.CallConfig
}

type userv1Notify = notificationv1.NotifyServiceClient

// New wraps an existing gRPC connection.
func New(conn *grpc.ClientConn, settings utils.Settings, log logger.Logger) Client {
	breaker := utils.NewBreaker(settings)
	none := utils.CallConfigFromSettings(settings, utils.RetryNone)
	none.Breaker = breaker

	return &client{
		stub: notificationv1.NewNotifyServiceClient(conn),
		log:  log,
		none: none,
	}
}

func (c *client) Send(ctx context.Context, in *notificationv1.SendRequest) (*notificationv1.SendResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "notification.NotifyService/Send", func(callCtx context.Context) (*notificationv1.SendResponse, error) {
		return c.stub.Send(callCtx, in)
	})
}
