package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/livekit/protocol/livekit"
	"github.com/redis/go-redis/v9"
	"github.com/wichat/wichat/backend/ms-core/internal/models"
	voicerepo "github.com/wichat/wichat/backend/ms-core/internal/voice/repository"
	"github.com/wichat/wichat/backend/ms-core/internal/voice/repository/mock"
	"github.com/wichat/wichat/backend/ms-core/internal/voice/service"
	livekitclient "github.com/wichat/wichat/backend/ms-core/pkg/livekit-client"
	chatEvents "github.com/wichat/wichat/backend/wi-shared/chat-events"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	userauth "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/auth"
	userchannel "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/channel"
	userpresence "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/presence"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

type fakeChannelClient struct {
	userchannel.Client
	resp *userv1.AssertGroupChannelAccessResponse
	err  error
}

func (f fakeChannelClient) AssertGroupChannelAccess(context.Context, *userv1.AssertGroupChannelAccessRequest) (*userv1.AssertGroupChannelAccessResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.resp, nil
}

func (f fakeChannelClient) ListGroupChannels(context.Context, *userv1.ListGroupChannelsRequest) (*userv1.ListGroupChannelsResponse, error) {
	return &userv1.ListGroupChannelsResponse{}, nil
}

type fakeUserClient struct {
	userauth.Client
}

func (f fakeUserClient) Me(context.Context, *userv1.MeRequest) (*userv1.MeResponse, error) {
	return &userv1.MeResponse{
		User: &userv1.User{
			Id:          "u1",
			DisplayName: "Alice",
			Username:    "alice",
		},
	}, nil
}

type fakePresenceClient struct {
	userpresence.Client
	lastActivity string
}

func (f *fakePresenceClient) SetVoiceActivityStatus(_ context.Context, req *userv1.SetVoiceActivityStatusRequest) (*userv1.SetVoiceActivityStatusResponse, error) {
	f.lastActivity = req.GetActivityStatus()
	return &userv1.SetVoiceActivityStatusResponse{Ok: true}, nil
}

type fakeLKClient struct {
	lastGrants     livekitclient.TokenGrants
	removedIDs     []string
	mutedAudioIDs  []string
	mutedVideoIDs  []string
	modUpdates     []livekitclient.ParticipantModeration
	startedEgress  int
	stoppedEgress  []string
	webhookEvent   *livekit.WebhookEvent
	webhookErr     error
	egressStartErr error
}

func (f *fakeLKClient) IssueToken(grants livekitclient.TokenGrants) (string, time.Time, error) {
	f.lastGrants = grants
	return "jwt-token", time.Now().UTC().Add(15 * time.Minute), nil
}

func (f *fakeLKClient) VerifyWebhook(string, []byte) (*livekit.WebhookEvent, error) {
	return f.webhookEvent, f.webhookErr
}

func (f *fakeLKClient) UpdateParticipantPermissions(_ context.Context, mod livekitclient.ParticipantModeration) error {
	f.modUpdates = append(f.modUpdates, mod)
	return nil
}

func (f *fakeLKClient) MuteAudioTracks(_ context.Context, _ string, identity string, _ bool) error {
	f.mutedAudioIDs = append(f.mutedAudioIDs, identity)
	return nil
}

func (f *fakeLKClient) MuteVideoTracks(_ context.Context, _ string, identity string, _ bool) error {
	f.mutedVideoIDs = append(f.mutedVideoIDs, identity)
	return nil
}

func (f *fakeLKClient) RemoveParticipant(_ context.Context, _ string, identity string) error {
	f.removedIDs = append(f.removedIDs, identity)
	return nil
}

func (f *fakeLKClient) ListParticipants(context.Context, string) ([]*livekit.ParticipantInfo, error) {
	return nil, nil
}

func (f *fakeLKClient) StartRoomCompositeEgress(context.Context, string, livekitclient.EgressS3Config) (string, error) {
	if f.egressStartErr != nil {
		return "", f.egressStartErr
	}
	f.startedEgress++
	return "egress-1", nil
}

func (f *fakeLKClient) StopEgress(_ context.Context, egressID string) error {
	f.stoppedEgress = append(f.stoppedEgress, egressID)
	return nil
}

type fakePublisher struct {
	occupancies []chatEvents.VoiceOccupancyUpdated
	recordings  []chatEvents.VoiceRecordingUpdated
	moderations []chatEvents.VoiceModerationApplied
	idleWarns   []chatEvents.VoiceIdleWarning
}

func (p *fakePublisher) PublishOccupancyUpdated(_ context.Context, ev chatEvents.VoiceOccupancyUpdated) error {
	p.occupancies = append(p.occupancies, ev)
	return nil
}
func (p *fakePublisher) PublishRecordingUpdated(_ context.Context, ev chatEvents.VoiceRecordingUpdated) error {
	p.recordings = append(p.recordings, ev)
	return nil
}
func (p *fakePublisher) PublishModerationApplied(_ context.Context, ev chatEvents.VoiceModerationApplied) error {
	p.moderations = append(p.moderations, ev)
	return nil
}
func (p *fakePublisher) PublishIdleWarning(_ context.Context, ev chatEvents.VoiceIdleWarning) error {
	p.idleWarns = append(p.idleWarns, ev)
	return nil
}

func newTestRedis(t *testing.T) voicerepo.RedisOccupancy {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return voicerepo.NewRedisOccupancy(rdb)
}

func TestIssueVoiceTokenPermissionsAndCaps(t *testing.T) {
	ctx := context.Background()
	st := mock.NewState()
	pg := mock.NewPgRepositoryMock(st)
	rds := newTestRedis(t)
	lk := &fakeLKClient{}
	pub := &fakePublisher{}
	pres := &fakePresenceClient{}

	t.Run("issues token with speak=false and video=false", func(t *testing.T) {
		svc := service.NewService(pg, rds, nil, lk, fakeChannelClient{
			resp: &userv1.AssertGroupChannelAccessResponse{
				ChannelUuid: "ch-uuid",
				UserUuid:    "u1",
				ChannelType: "text_and_voice",
				Permissions: []string{"view_channels", "connect"},
				BitrateKbps: 96,
			},
		}, fakeUserClient{}, pres, pub, service.Options{LiveKitURL: "wss://lk.test"}, nil)

		res, err := svc.IssueVoiceToken(ctx, &models.IssueVoiceTokenInput{
			AccessToken:     "tok",
			GroupPublicID:   "g1",
			ChannelPublicID: "c1",
		})
		if err != nil {
			t.Fatalf("IssueVoiceToken() error = %v", err)
		}
		if !res.InitialServerMuted || res.CanPublishAudio || res.CanPublishVideo {
			t.Fatalf("unexpected grants: %+v", res)
		}
		if res.BitrateKbps != 96 {
			t.Fatalf("BitrateKbps = %d, want 96", res.BitrateKbps)
		}
	})

	t.Run("rejects text_only channel", func(t *testing.T) {
		svcText := service.NewService(pg, rds, nil, lk, fakeChannelClient{
			resp: &userv1.AssertGroupChannelAccessResponse{
				ChannelUuid: "ch-uuid",
				UserUuid:    "u2",
				ChannelType: "text_only",
				Permissions: []string{"connect"},
			},
		}, fakeUserClient{}, pres, pub, service.Options{}, nil)

		if _, err := svcText.IssueVoiceToken(ctx, &models.IssueVoiceTokenInput{AccessToken: "tok", GroupPublicID: "g1", ChannelPublicID: "c1"}); err == nil {
			t.Fatal("expected error for text_only channel")
		} else if ex, ok := exception.As(err); !ok || ex.Kind != exception.KindInvalidInput {
			t.Fatalf("expected InvalidInput, got %v", err)
		}
	})

	t.Run("enforces room cap", func(t *testing.T) {
		rdsCap := newTestRedis(t)
		_ = rdsCap.SetRoomMeta(ctx, "ch-cap", models.RoomMeta{GroupPublicID: "g1", ChannelPublicID: "c1"})
		_ = rdsCap.UpsertOccupant(ctx, "ch-cap", "g1", models.VoiceOccupant{UserID: "other-1"})
		_ = rdsCap.UpsertOccupant(ctx, "ch-cap", "g1", models.VoiceOccupant{UserID: "other-2"})

		svc := service.NewService(pg, rdsCap, nil, lk, fakeChannelClient{
			resp: &userv1.AssertGroupChannelAccessResponse{
				ChannelUuid: "ch-cap",
				UserUuid:    "u-new",
				ChannelType: "voice_only",
				Permissions: []string{"connect", "speak"},
				UserLimit:   2,
			},
		}, fakeUserClient{}, pres, pub, service.Options{}, nil)

		_, err := svc.IssueVoiceToken(ctx, &models.IssueVoiceTokenInput{AccessToken: "tok", GroupPublicID: "g1", ChannelPublicID: "c1"})
		if ex, ok := exception.As(err); !ok || ex.Kind != exception.KindConflict {
			t.Fatalf("expected Conflict (room_full), got %v", err)
		}
	})

	t.Run("lists group occupancy and acknowledges idle warning", func(t *testing.T) {
		rdsOcc := newTestRedis(t)
		_ = rdsOcc.SetRoomMeta(ctx, "ch-1", models.RoomMeta{GroupPublicID: "g1", ChannelPublicID: "c1", IsRecording: true})
		_ = rdsOcc.UpsertOccupant(ctx, "ch-1", "g1", models.VoiceOccupant{UserID: "u1", IdleWarned: true})

		svc := service.NewService(pg, rdsOcc, nil, lk, fakeChannelClient{
			resp: &userv1.AssertGroupChannelAccessResponse{
				ChannelUuid: "ch-1",
				UserUuid:    "u1",
				ChannelType: "voice_only",
				Permissions: []string{"view_channels", "connect"},
			},
		}, fakeUserClient{}, pres, pub, service.Options{}, nil)

		occ, err := svc.ListGroupOccupancy(ctx, "tok", "g1")
		if err != nil || len(occ) != 1 || !occ[0].IsRecording || len(occ[0].Occupants) != 1 {
			t.Fatalf("ListGroupOccupancy = (%+v, %v)", occ, err)
		}
		if err := svc.AcknowledgeIdleWarning(ctx, "tok", "g1", "c1"); err != nil {
			t.Fatalf("AcknowledgeIdleWarning() error = %v", err)
		}
		updated, _ := rdsOcc.ListOccupants(ctx, "ch-1")
		if len(updated) != 1 || updated[0].IdleWarned {
			t.Fatalf("expected IdleWarned=false after ack, got %+v", updated)
		}

		// Rate limit check (11th token request in 1 min)
		for i := 0; i < service.TokenRateLimitPerMin; i++ {
			_, _ = svc.IssueVoiceToken(ctx, &models.IssueVoiceTokenInput{
				AccessToken: "tok", GroupPublicID: "g1", ChannelPublicID: "c1",
			})
		}
		if _, err := svc.IssueVoiceToken(ctx, &models.IssueVoiceTokenInput{
			AccessToken: "tok", GroupPublicID: "g1", ChannelPublicID: "c1",
		}); err == nil {
			t.Fatal("expected rate limit error on 11th token request")
		}

		// DM channel + error paths
		svcDM := service.NewService(pg, newTestRedis(t), nil, lk, fakeChannelClient{
			resp: &userv1.AssertGroupChannelAccessResponse{
				ChannelUuid: "dm-1", UserUuid: "u1", ChannelType: "dm",
			},
		}, fakeUserClient{}, pres, pub, service.Options{}, nil)
		if res, err := svcDM.IssueVoiceToken(ctx, &models.IssueVoiceTokenInput{
			AccessToken: "tok", ChannelPublicID: "dm-1",
		}); err != nil || res.RoomName != "dm:dm-1" {
			t.Fatalf("DM IssueVoiceToken = (%+v, %v)", res, err)
		}

		svcErr := service.NewService(pg, rdsOcc, nil, lk, fakeChannelClient{
			err: exception.PermissionDenied("denied", nil),
		}, fakeUserClient{}, pres, pub, service.Options{}, nil)
		if occ, err := svcErr.ListGroupOccupancy(ctx, "tok", "g1"); err != nil || len(occ) != 1 {
			t.Fatalf("expected 1 occupancy row, got (%+v, %v)", occ, err)
		}
		if err := svcErr.AcknowledgeIdleWarning(ctx, "tok", "g1", "c1"); err == nil {
			t.Fatal("expected error from AcknowledgeIdleWarning")
		}
		if _, err := svcErr.GetQuotaUsage(ctx, "", "g1"); err == nil {
			t.Fatal("expected error from GetQuotaUsage with empty token")
		}
		if _, err := svcErr.ModerateParticipant(ctx, &models.ModerateParticipantInput{
			AccessToken: "tok", GroupPublicID: "g1", ChannelPublicID: "c1", TargetUserID: "u2", Action: "mute",
		}); err == nil {
			t.Fatal("expected error from ModerateParticipant")
		}
	})
}
