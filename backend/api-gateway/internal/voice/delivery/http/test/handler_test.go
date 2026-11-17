package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	voicedelivery "github.com/wichat/wichat/backend/api-gateway/internal/voice/delivery/http"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
	msCore "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-core"
	corevoice "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-core/grpc/voice"
	corev1 "github.com/wichat/wichat/backend/wi-shared/models/gen/core/v1"
)

type fakeVoiceClient struct {
	corevoice.Client
}

func (fakeVoiceClient) IssueVoiceToken(context.Context, *corev1.IssueVoiceTokenRequest) (*corev1.IssueVoiceTokenResponse, error) {
	return &corev1.IssueVoiceTokenResponse{
		Token:           "jwt-token",
		LivekitUrl:      "ws://localhost:7880",
		RoomName:        "grp_g1_ch_c1",
		ChannelId:       "c1",
		CanPublishAudio: true,
		BitrateKbps:     64,
		VideoQuality:    "auto",
	}, nil
}

func (fakeVoiceClient) ListGroupOccupancy(context.Context, *corev1.ListGroupOccupancyRequest) (*corev1.ListGroupOccupancyResponse, error) {
	return &corev1.ListGroupOccupancyResponse{
		Channels: []*corev1.ChannelOccupancy{{ChannelId: "c1", IsRecording: false}},
	}, nil
}

func (fakeVoiceClient) ModerateParticipant(context.Context, *corev1.ModerateParticipantRequest) (*corev1.ModerateParticipantResponse, error) {
	return &corev1.ModerateParticipantResponse{Ok: true, Occupant: &corev1.VoiceOccupant{UserId: "u2", IsServerMuted: true}}, nil
}

func (fakeVoiceClient) AcknowledgeIdleWarning(context.Context, *corev1.AcknowledgeIdleWarningRequest) (*corev1.AcknowledgeIdleWarningResponse, error) {
	return &corev1.AcknowledgeIdleWarningResponse{Ok: true}, nil
}

func (fakeVoiceClient) StartRecording(context.Context, *corev1.StartRecordingRequest) (*corev1.StartRecordingResponse, error) {
	return &corev1.StartRecordingResponse{Recording: &corev1.CallRecordingSummary{Id: "rec-1", Status: "recording"}}, nil
}

func (fakeVoiceClient) StopRecording(context.Context, *corev1.StopRecordingRequest) (*corev1.StopRecordingResponse, error) {
	return &corev1.StopRecordingResponse{Recording: &corev1.CallRecordingSummary{Id: "rec-1", Status: "completed"}}, nil
}

func (fakeVoiceClient) ListChannelRecordings(context.Context, *corev1.ListChannelRecordingsRequest) (*corev1.ListChannelRecordingsResponse, error) {
	return &corev1.ListChannelRecordingsResponse{Recordings: []*corev1.CallRecordingSummary{{Id: "rec-1"}}}, nil
}

func (fakeVoiceClient) GetQuotaUsage(context.Context, *corev1.GetQuotaUsageRequest) (*corev1.GetQuotaUsageResponse, error) {
	return &corev1.GetQuotaUsageResponse{Month: "2026-11", UsedParticipantMinutes: 120, LimitParticipantMinutes: 5000, UsagePercent: 2.4}, nil
}

func (fakeVoiceClient) HandleWebhook(context.Context, *corev1.HandleWebhookRequest) (*corev1.HandleWebhookResponse, error) {
	return &corev1.HandleWebhookResponse{Ok: true}, nil
}

func newTestClients() *microservices.Clients {
	return &microservices.Clients{
		Core: msCore.Service{
			GRPC: msCore.GRPC{
				Voice: fakeVoiceClient{},
			},
		},
	}
}

func TestVoiceHandler_Endpoints(t *testing.T) {
	mux := http.NewServeMux()
	voicedelivery.Register(mux, voicedelivery.NewHandler(newTestClients(), logger.New(logger.Options{Service: "gw-test", Env: "test", Level: "error"})))

	cases := []struct {
		name   string
		method string
		path   string
		body   string
		auth   string
		want   int
	}{
		{"unauth token", http.MethodPost, "/api/v1/channels/c1/voice/token", "", "", http.StatusUnauthorized},
		{"issue token", http.MethodPost, "/api/v1/groups/g1/channels/c1/voice/token", "", "Bearer at", http.StatusOK},
		{"list occupancy", http.MethodGet, "/api/v1/groups/g1/voice/occupancy", "", "Bearer at", http.StatusOK},
		{"moderate", http.MethodPost, "/api/v1/channels/c1/voice/moderate", `{"group_id":"g1","target_user_id":"u2","action":"mute"}`, "Bearer at", http.StatusOK},
		{"keepalive", http.MethodPost, "/api/v1/channels/c1/voice/keepalive?group_id=g1", "", "Bearer at", http.StatusOK},
		{"start recording", http.MethodPost, "/api/v1/channels/c1/voice/recording/start?group_id=g1", "", "Bearer at", http.StatusOK},
		{"stop recording", http.MethodPost, "/api/v1/channels/c1/voice/recording/stop", `{"group_id":"g1","recording_id":"rec-1"}`, "Bearer at", http.StatusOK},
		{"list recordings", http.MethodGet, "/api/v1/channels/c1/voice/recordings?group_id=g1&limit=10", "", "Bearer at", http.StatusOK},
		{"quota", http.MethodGet, "/api/v1/voice/quota", "", "Bearer at", http.StatusOK},
		{"webhook", http.MethodPost, "/api/v1/webhooks/livekit", `{"event":"room_finished"}`, "Bearer lk-sig", http.StatusOK},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.auth != "" {
				req.Header.Set("Authorization", tc.auth)
			}
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Fatalf("%s %s = %d (%s), want %d", tc.method, tc.path, rec.Code, rec.Body.String(), tc.want)
			}
		})
	}
}
