// Package general is the gRPC client for ms-notification GeneralService.
package general

import (
	"context"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	notificationv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/notification/v1"
	"google.golang.org/grpc"
)

// Client calls GeneralService health RPCs.
type Client interface {
	Health(ctx context.Context, in *notificationv1.HealthRequest) (*notificationv1.HealthResponse, error)
}

type client struct {
	stub notificationv1.GeneralServiceClient
	log  logger.Logger
	read utils.CallConfig
}

// New wraps an existing gRPC connection.
func New(conn *grpc.ClientConn, settings utils.Settings, log logger.Logger) Client {
	breaker := utils.NewBreaker(settings)
	read := utils.CallConfigFromSettings(settings, utils.RetryIdempotent)
	read.Breaker = breaker

	return &client{
		stub: notificationv1.NewGeneralServiceClient(conn),
		log:  log,
		read: read,
	}
}

func (c *client) Health(ctx context.Context, in *notificationv1.HealthRequest) (*notificationv1.HealthResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "notification.GeneralService/Health", func(callCtx context.Context) (*notificationv1.HealthResponse, error) {
		return c.stub.Health(callCtx, in)
	})
}
