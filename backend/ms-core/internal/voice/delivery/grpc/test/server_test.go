package grpc_test

import (
	"context"
	"testing"
	"time"

	"github.com/wichat/wichat/backend/ms-core/internal/models"
	voicegrpc "github.com/wichat/wichat/backend/ms-core/internal/voice/delivery/grpc"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	corev1 "github.com/wichat/wichat/backend/wi-shared/models/gen/core/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type stubVoiceService struct {
	tokenRes   models.IssueVoiceTokenResult
	tokenErr   error
	modRes     models.VoiceOccupant
	modErr     error
	webhookErr error
	occRes     []models.ChannelOccupancy
	occErr     error
	recRes     models.CallRecording
	recErr     error
	recsRes    []models.CallRecording
	recsErr    error
	ackErr     error
	quotaRes   models.QuotaUsageResult
	quotaErr   error
}

func (s *stubVoiceService) IssueVoiceToken(_ context.Context, _ *models.IssueVoiceTokenInput) (models.IssueVoiceTokenResult, error) {
	return s.tokenRes, s.tokenErr
}
func (s *stubVoiceService) ModerateParticipant(_ context.Context, _ *models.ModerateParticipantInput) (models.VoiceOccupant, error) {
	return s.modRes, s.modErr
}
func (s *stubVoiceService) HandleWebhook(_ context.Context, _ string, _ []byte) error {
	return s.webhookErr
}
func (s *stubVoiceService) ListGroupOccupancy(_ context.Context, _, _ string) ([]models.ChannelOccupancy, error) {
	return s.occRes, s.occErr
}
func (s *stubVoiceService) StartRecording(_ context.Context, _, _, _ string) (models.CallRecording, error) {
	return s.recRes, s.recErr
}
func (s *stubVoiceService) StopRecording(_ context.Context, _, _, _, _ string) (models.CallRecording, error) {
	return s.recRes, s.recErr
}
func (s *stubVoiceService) ListChannelRecordings(_ context.Context, _, _, _ string, _ int) ([]models.CallRecording, error) {
	return s.recsRes, s.recsErr
}
func (s *stubVoiceService) AcknowledgeIdleWarning(_ context.Context, _, _, _ string) error {
	return s.ackErr
}
func (s *stubVoiceService) GetQuotaUsage(_ context.Context, _, _ string) (models.QuotaUsageResult, error) {
	return s.quotaRes, s.quotaErr
}
func (s *stubVoiceService) RunIdleCheck(_ context.Context) error { return nil }

func TestServer_RPCsSuccess(t *testing.T) {
	now := time.Date(2026, 11, 17, 10, 0, 0, 0, time.UTC)
	stoppedBy := "usr-1"
	srv := voicegrpc.NewServer(&stubVoiceService{
		tokenRes: models.IssueVoiceTokenResult{
			Token:            "jwt-1",
			LiveKitURL:       "ws://localhost:7880",
			RoomName:         "grp_g1_ch_c1",
			ChannelID:        "c1",
			ExpiresAt:        now,
			CanPublishAudio:  true,
			CanPublishVideo:  true,
			UseVoiceActivity: true,
			BitrateKbps:      64,
			VideoQuality:     "auto",
			IsRecording:      true,
			Occupants: []models.VoiceOccupant{{
				UserID:      "usr-1",
				DisplayName: "Alice",
				Username:    "alice",
				JoinedAt:    now.Format(time.RFC3339),
			}},
		},
		modRes: models.VoiceOccupant{UserID: "usr-2", IsServerMuted: true},
		occRes: []models.ChannelOccupancy{{ChannelID: "c1", IsRecording: true, Occupants: []models.VoiceOccupant{{UserID: "usr-1"}}}},
		recRes: models.CallRecording{
			PublicID:  "rec-1",
			SessionID: "sess-1",
			ChannelID: "c1",
			EgressID:  "eg-1",
			Status:    "recording",
			StartedBy: "usr-1",
			StoppedBy: &stoppedBy,
			StartedAt: now,
			EndedAt:   &now,
		},
		recsRes:  []models.CallRecording{{ID: "rec-id-1", ChannelID: "c1", StartedAt: now}},
		quotaRes: models.QuotaUsageResult{Month: "2026-11", UsedParticipantMinutes: 100, LimitParticipantMinutes: 5000, UsagePercent: 2.0, WarningActive: false},
	})

	ctx := context.Background()
	tokResp, err := srv.IssueVoiceToken(ctx, &corev1.IssueVoiceTokenRequest{AccessToken: "at", GroupId: "g1", ChannelId: "c1"})
	if err != nil || tokResp.GetToken() != "jwt-1" || len(tokResp.GetOccupants()) != 1 {
		t.Fatalf("IssueVoiceToken = (%+v, %v)", tokResp, err)
	}

	modResp, err := srv.ModerateParticipant(ctx, &corev1.ModerateParticipantRequest{AccessToken: "at", TargetUserId: "usr-2", Action: "mute"})
	if err != nil || !modResp.GetOk() || !modResp.GetOccupant().GetIsServerMuted() {
		t.Fatalf("ModerateParticipant = (%+v, %v)", modResp, err)
	}

	whResp, err := srv.HandleWebhook(ctx, &corev1.HandleWebhookRequest{AuthHeader: "Bearer h", RawBody: []byte("{}")})
	if err != nil || !whResp.GetOk() {
		t.Fatalf("HandleWebhook = (%+v, %v)", whResp, err)
	}

	occResp, err := srv.ListGroupOccupancy(ctx, &corev1.ListGroupOccupancyRequest{AccessToken: "at", GroupId: "g1"})
	if err != nil || len(occResp.GetChannels()) != 1 {
		t.Fatalf("ListGroupOccupancy = (%+v, %v)", occResp, err)
	}

	startResp, err := srv.StartRecording(ctx, &corev1.StartRecordingRequest{AccessToken: "at", GroupId: "g1", ChannelId: "c1"})
	if err != nil || startResp.GetRecording().GetId() != "rec-1" {
		t.Fatalf("StartRecording = (%+v, %v)", startResp, err)
	}

	stopResp, err := srv.StopRecording(ctx, &corev1.StopRecordingRequest{AccessToken: "at", GroupId: "g1", ChannelId: "c1", RecordingId: "rec-1"})
	if err != nil || stopResp.GetRecording().GetStoppedBy() != "usr-1" {
		t.Fatalf("StopRecording = (%+v, %v)", stopResp, err)
	}

	listResp, err := srv.ListChannelRecordings(ctx, &corev1.ListChannelRecordingsRequest{AccessToken: "at", GroupId: "g1", ChannelId: "c1", Limit: 10})
	if err != nil || len(listResp.GetRecordings()) != 1 || listResp.GetRecordings()[0].GetId() != "rec-id-1" {
		t.Fatalf("ListChannelRecordings = (%+v, %v)", listResp, err)
	}

	ackResp, err := srv.AcknowledgeIdleWarning(ctx, &corev1.AcknowledgeIdleWarningRequest{AccessToken: "at", GroupId: "g1", ChannelId: "c1"})
	if err != nil || !ackResp.GetOk() {
		t.Fatalf("AcknowledgeIdleWarning = (%+v, %v)", ackResp, err)
	}

	qResp, err := srv.GetQuotaUsage(ctx, &corev1.GetQuotaUsageRequest{AccessToken: "at", GroupId: "g1"})
	if err != nil || qResp.GetUsedParticipantMinutes() != 100 {
		t.Fatalf("GetQuotaUsage = (%+v, %v)", qResp, err)
	}
}

func TestServer_RPCsErrorMapping(t *testing.T) {
	forbidden := exception.PermissionDenied("forbidden", nil)
	srv := voicegrpc.NewServer(&stubVoiceService{
		tokenErr:   forbidden,
		modErr:     forbidden,
		webhookErr: forbidden,
		occErr:     forbidden,
		recErr:     forbidden,
		recsErr:    forbidden,
		ackErr:     forbidden,
		quotaErr:   forbidden,
	})
	ctx := context.Background()

	check := func(name string, err error) {
		t.Helper()
		if status.Code(err) != codes.PermissionDenied {
			t.Fatalf("%s: expected PermissionDenied, got %v", name, err)
		}
	}

	_, err := srv.IssueVoiceToken(ctx, &corev1.IssueVoiceTokenRequest{})
	check("IssueVoiceToken", err)
	_, err = srv.ModerateParticipant(ctx, &corev1.ModerateParticipantRequest{})
	check("ModerateParticipant", err)
	_, err = srv.HandleWebhook(ctx, &corev1.HandleWebhookRequest{})
	check("HandleWebhook", err)
	_, err = srv.ListGroupOccupancy(ctx, &corev1.ListGroupOccupancyRequest{})
	check("ListGroupOccupancy", err)
	_, err = srv.StartRecording(ctx, &corev1.StartRecordingRequest{})
	check("StartRecording", err)
	_, err = srv.StopRecording(ctx, &corev1.StopRecordingRequest{})
	check("StopRecording", err)
	_, err = srv.ListChannelRecordings(ctx, &corev1.ListChannelRecordingsRequest{})
	check("ListChannelRecordings", err)
	_, err = srv.AcknowledgeIdleWarning(ctx, &corev1.AcknowledgeIdleWarningRequest{})
	check("AcknowledgeIdleWarning", err)
	_, err = srv.GetQuotaUsage(ctx, &corev1.GetQuotaUsageRequest{})
	check("GetQuotaUsage", err)
}
