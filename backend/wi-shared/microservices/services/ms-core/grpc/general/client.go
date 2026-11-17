// Package general is the gRPC client for ms-core GeneralService.
package general

import (
	"context"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	corev1 "github.com/wichat/wichat/backend/wi-shared/models/gen/core/v1"
	"google.golang.org/grpc"
)

// Client calls GeneralService with shared timeouts and retries.
type Client interface {
	Health(ctx context.Context, in *corev1.HealthRequest) (*corev1.HealthResponse, error)
}

type client struct {
	stub corev1.GeneralServiceClient
	log  logger.Logger
	read utils.CallConfig
}

// New wraps an existing gRPC connection.
func New(conn *grpc.ClientConn, settings utils.Settings, log logger.Logger) Client {
	read := utils.CallConfigFromSettings(settings, utils.RetryNone)

	return &client{
		stub: corev1.NewGeneralServiceClient(conn),
		log:  log,
		read: read,
	}
}

func (c *client) Health(ctx context.Context, in *corev1.HealthRequest) (*corev1.HealthResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "core.GeneralService/Health", func(callCtx context.Context) (*corev1.HealthResponse, error) {
		return c.stub.Health(callCtx, in)
	})
}
