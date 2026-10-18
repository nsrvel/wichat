// Package general is the gRPC client for ms-user GeneralService.
package general

import (
	"context"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"google.golang.org/grpc"
)

// Client calls ms-user GeneralService with shared timeouts, retries, and circuit breaking.
type Client interface {
	Health(ctx context.Context, in *userv1.HealthRequest) (*userv1.HealthResponse, error)
}

type client struct {
	stub userv1.GeneralServiceClient
	log  logger.Logger
	read utils.CallConfig
}

// New wraps an existing gRPC connection (tests or custom dial). Caller owns conn lifetime.
func New(conn *grpc.ClientConn, settings utils.Settings, log logger.Logger) Client {
	breaker := utils.NewBreaker(settings)
	read := utils.CallConfigFromSettings(settings, utils.RetryIdempotent)
	read.Breaker = breaker

	return &client{
		stub: userv1.NewGeneralServiceClient(conn),
		log:  log,
		read: read,
	}
}

func (c *client) Health(ctx context.Context, in *userv1.HealthRequest) (*userv1.HealthResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "general.GeneralService/Health", func(callCtx context.Context) (*userv1.HealthResponse, error) {
		return c.stub.Health(callCtx, in)
	})
}
