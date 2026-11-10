// Package general is the gRPC client for ms-chat GeneralService.
package general

import (
	"context"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	chatv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/chat/v1"
	"google.golang.org/grpc"
)

// Client calls GeneralService with shared timeouts and retries.
type Client interface {
	Health(ctx context.Context, in *chatv1.HealthRequest) (*chatv1.HealthResponse, error)
}

type client struct {
	stub chatv1.GeneralServiceClient
	log  logger.Logger
	read utils.CallConfig
}

// New wraps an existing gRPC connection.
func New(conn *grpc.ClientConn, settings utils.Settings, log logger.Logger) Client {
	breaker := utils.NewBreaker(settings)
	read := utils.CallConfigFromSettings(settings, utils.RetryIdempotent)
	read.Breaker = breaker

	return &client{
		stub: chatv1.NewGeneralServiceClient(conn),
		log:  log,
		read: read,
	}
}

func (c *client) Health(ctx context.Context, in *chatv1.HealthRequest) (*chatv1.HealthResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "chat.GeneralService/Health", func(callCtx context.Context) (*chatv1.HealthResponse, error) {
		return c.stub.Health(callCtx, in)
	})
}
