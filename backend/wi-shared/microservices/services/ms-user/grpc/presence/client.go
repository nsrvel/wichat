// Package presence is the gRPC client for ms-user presence domain RPCs on PresenceService.
package presence

import (
	"context"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

// Client calls presence domain PresenceService RPCs with shared timeouts, retries, and circuit breaking.
type Client interface {
	HeartbeatPresence(ctx context.Context, in *userv1.HeartbeatPresenceRequest) (*userv1.PresenceTransitionResponse, error)
	DisconnectPresence(ctx context.Context, in *userv1.DisconnectPresenceRequest) (*userv1.PresenceTransitionResponse, error)
	SetManualPresence(ctx context.Context, in *userv1.SetManualPresenceRequest) (*userv1.SetManualPresenceResponse, error)
	GetMyPresence(ctx context.Context, in *userv1.GetMyPresenceRequest) (*userv1.GetMyPresenceResponse, error)
	SetVoiceActivityStatus(ctx context.Context, in *userv1.SetVoiceActivityStatusRequest) (*userv1.SetVoiceActivityStatusResponse, error)
}

type client struct {
	stub userv1.PresenceServiceClient
	log  logger.Logger
	none utils.CallConfig
	read utils.CallConfig
}

// New wraps an existing PresenceService client stub.
func New(stub userv1.PresenceServiceClient, settings utils.Settings, log logger.Logger) Client {
	breaker := utils.NewBreaker(settings)
	none := utils.CallConfigFromSettings(settings, utils.RetryNone)
	none.Breaker = breaker
	read := utils.CallConfigFromSettings(settings, utils.RetryIdempotent)
	read.Breaker = breaker

	return &client{
		stub: stub,
		log:  log,
		none: none,
		read: read,
	}
}

func (c *client) HeartbeatPresence(ctx context.Context, in *userv1.HeartbeatPresenceRequest) (*userv1.PresenceTransitionResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "presence.PresenceService/HeartbeatPresence", func(callCtx context.Context) (*userv1.PresenceTransitionResponse, error) {
		return c.stub.HeartbeatPresence(callCtx, in)
	})
}

func (c *client) DisconnectPresence(ctx context.Context, in *userv1.DisconnectPresenceRequest) (*userv1.PresenceTransitionResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "presence.PresenceService/DisconnectPresence", func(callCtx context.Context) (*userv1.PresenceTransitionResponse, error) {
		return c.stub.DisconnectPresence(callCtx, in)
	})
}

func (c *client) SetManualPresence(ctx context.Context, in *userv1.SetManualPresenceRequest) (*userv1.SetManualPresenceResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "presence.PresenceService/SetManualPresence", func(callCtx context.Context) (*userv1.SetManualPresenceResponse, error) {
		return c.stub.SetManualPresence(callCtx, in)
	})
}

func (c *client) GetMyPresence(ctx context.Context, in *userv1.GetMyPresenceRequest) (*userv1.GetMyPresenceResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "presence.PresenceService/GetMyPresence", func(callCtx context.Context) (*userv1.GetMyPresenceResponse, error) {
		return c.stub.GetMyPresence(callCtx, in)
	})
}

func (c *client) SetVoiceActivityStatus(ctx context.Context, in *userv1.SetVoiceActivityStatusRequest) (*userv1.SetVoiceActivityStatusResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "presence.PresenceService/SetVoiceActivityStatus", func(callCtx context.Context) (*userv1.SetVoiceActivityStatusResponse, error) {
		return c.stub.SetVoiceActivityStatus(callCtx, in)
	})
}
