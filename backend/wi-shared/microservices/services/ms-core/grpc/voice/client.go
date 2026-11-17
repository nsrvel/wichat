// Package voice is the gRPC client for ms-core VoiceService.
package voice

import (
	"context"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	corev1 "github.com/wichat/wichat/backend/wi-shared/models/gen/core/v1"
	"google.golang.org/grpc"
)

// Client calls VoiceService with shared timeouts and retries.
type Client interface {
	IssueVoiceToken(ctx context.Context, in *corev1.IssueVoiceTokenRequest) (*corev1.IssueVoiceTokenResponse, error)
	ModerateParticipant(ctx context.Context, in *corev1.ModerateParticipantRequest) (*corev1.ModerateParticipantResponse, error)
	HandleWebhook(ctx context.Context, in *corev1.HandleWebhookRequest) (*corev1.HandleWebhookResponse, error)
	ListGroupOccupancy(ctx context.Context, in *corev1.ListGroupOccupancyRequest) (*corev1.ListGroupOccupancyResponse, error)
	StartRecording(ctx context.Context, in *corev1.StartRecordingRequest) (*corev1.StartRecordingResponse, error)
	StopRecording(ctx context.Context, in *corev1.StopRecordingRequest) (*corev1.StopRecordingResponse, error)
	ListChannelRecordings(ctx context.Context, in *corev1.ListChannelRecordingsRequest) (*corev1.ListChannelRecordingsResponse, error)
	AcknowledgeIdleWarning(ctx context.Context, in *corev1.AcknowledgeIdleWarningRequest) (*corev1.AcknowledgeIdleWarningResponse, error)
	GetQuotaUsage(ctx context.Context, in *corev1.GetQuotaUsageRequest) (*corev1.GetQuotaUsageResponse, error)
}

type client struct {
	stub corev1.VoiceServiceClient
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
		stub: corev1.NewVoiceServiceClient(conn),
		log:  log,
		read: read,
		none: none,
	}
}

func (c *client) IssueVoiceToken(ctx context.Context, in *corev1.IssueVoiceTokenRequest) (*corev1.IssueVoiceTokenResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "core.VoiceService/IssueVoiceToken", func(callCtx context.Context) (*corev1.IssueVoiceTokenResponse, error) {
		return c.stub.IssueVoiceToken(callCtx, in)
	})
}

func (c *client) ModerateParticipant(ctx context.Context, in *corev1.ModerateParticipantRequest) (*corev1.ModerateParticipantResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "core.VoiceService/ModerateParticipant", func(callCtx context.Context) (*corev1.ModerateParticipantResponse, error) {
		return c.stub.ModerateParticipant(callCtx, in)
	})
}

func (c *client) HandleWebhook(ctx context.Context, in *corev1.HandleWebhookRequest) (*corev1.HandleWebhookResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "core.VoiceService/HandleWebhook", func(callCtx context.Context) (*corev1.HandleWebhookResponse, error) {
		return c.stub.HandleWebhook(callCtx, in)
	})
}

func (c *client) ListGroupOccupancy(ctx context.Context, in *corev1.ListGroupOccupancyRequest) (*corev1.ListGroupOccupancyResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "core.VoiceService/ListGroupOccupancy", func(callCtx context.Context) (*corev1.ListGroupOccupancyResponse, error) {
		return c.stub.ListGroupOccupancy(callCtx, in)
	})
}

func (c *client) StartRecording(ctx context.Context, in *corev1.StartRecordingRequest) (*corev1.StartRecordingResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "core.VoiceService/StartRecording", func(callCtx context.Context) (*corev1.StartRecordingResponse, error) {
		return c.stub.StartRecording(callCtx, in)
	})
}

func (c *client) StopRecording(ctx context.Context, in *corev1.StopRecordingRequest) (*corev1.StopRecordingResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "core.VoiceService/StopRecording", func(callCtx context.Context) (*corev1.StopRecordingResponse, error) {
		return c.stub.StopRecording(callCtx, in)
	})
}

func (c *client) ListChannelRecordings(ctx context.Context, in *corev1.ListChannelRecordingsRequest) (*corev1.ListChannelRecordingsResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "core.VoiceService/ListChannelRecordings", func(callCtx context.Context) (*corev1.ListChannelRecordingsResponse, error) {
		return c.stub.ListChannelRecordings(callCtx, in)
	})
}

func (c *client) AcknowledgeIdleWarning(ctx context.Context, in *corev1.AcknowledgeIdleWarningRequest) (*corev1.AcknowledgeIdleWarningResponse, error) {
	return utils.CallUnaryResult(ctx, c.none, c.log, "core.VoiceService/AcknowledgeIdleWarning", func(callCtx context.Context) (*corev1.AcknowledgeIdleWarningResponse, error) {
		return c.stub.AcknowledgeIdleWarning(callCtx, in)
	})
}

func (c *client) GetQuotaUsage(ctx context.Context, in *corev1.GetQuotaUsageRequest) (*corev1.GetQuotaUsageResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "core.VoiceService/GetQuotaUsage", func(callCtx context.Context) (*corev1.GetQuotaUsageResponse, error) {
		return c.stub.GetQuotaUsage(callCtx, in)
	})
}
