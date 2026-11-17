package service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/livekit/protocol/livekit"
	"github.com/wichat/wichat/backend/ms-core/internal/models"
	"github.com/wichat/wichat/backend/ms-core/internal/voice/repository/mock"
	"github.com/wichat/wichat/backend/ms-core/internal/voice/service"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

func TestHandleWebhookLifecycleAndVideoCap(t *testing.T) {
	ctx := context.Background()
	st := mock.NewState()
	pg := mock.NewPgRepositoryMock(st)
	rds := newTestRedis(t)
	lk := &fakeLKClient{}
	pub := &fakePublisher{}
	pres := &fakePresenceClient{}

	_ = rds.SetRoomMeta(ctx, "ch1", models.RoomMeta{GroupPublicID: "g1", ChannelPublicID: "c1", RoomName: "channel:ch1"})

	svc := service.NewService(pg, rds, nil, lk, fakeChannelClient{
		resp: &userv1.AssertGroupChannelAccessResponse{ChannelUuid: "ch1", UserUuid: "u1", ChannelType: "voice_only"},
	}, fakeUserClient{}, pres, pub, service.Options{}, nil)

	lk.webhookEvent = &livekit.WebhookEvent{
		Id:    "ev-start",
		Event: "room_started",
		Room:  &livekit.Room{Name: "channel:ch1"},
	}
	if err := svc.HandleWebhook(ctx, "sig", []byte("{}")); err != nil {
		t.Fatalf("room_started error = %v", err)
	}

	lk.webhookEvent = &livekit.WebhookEvent{
		Id:          "ev-join",
		Event:       "participant_joined",
		Room:        &livekit.Room{Name: "channel:ch1"},
		Participant: &livekit.ParticipantInfo{Identity: "u1", Name: "Alice", Metadata: `{"display_name":"Alice B","username":"aliceb","avatar_url":"https://a.test/1.png"}`},
	}
	if err := svc.HandleWebhook(ctx, "sig", []byte("{}")); err != nil {
		t.Fatalf("participant_joined error = %v", err)
	}
	if pres.lastActivity != "In a call" {
		t.Fatalf("lastActivity = %q, want 'In a call'", pres.lastActivity)
	}

	// Fill 25 video publishers, then 26th must be muted by SFU
	for i := 0; i < service.MaxVideoPublishers; i++ {
		_ = rds.UpsertOccupant(ctx, "ch1", "g1", models.VoiceOccupant{
			UserID:     fmt.Sprintf("vid-%d", i),
			IsCameraOn: true,
		})
	}
	lk.webhookEvent = &livekit.WebhookEvent{
		Id:          "ev-vid-26",
		Event:       "track_published",
		Room:        &livekit.Room{Name: "channel:ch1"},
		Participant: &livekit.ParticipantInfo{Identity: "u1"},
		Track:       &livekit.TrackInfo{Source: livekit.TrackSource_CAMERA},
	}
	if err := svc.HandleWebhook(ctx, "sig", []byte("{}")); err != nil {
		t.Fatalf("track_published error = %v", err)
	}
	if len(lk.mutedVideoIDs) != 1 || lk.mutedVideoIDs[0] != "u1" {
		t.Fatalf("expected 26th video publisher u1 to be muted, got %+v", lk.mutedVideoIDs)
	}

	// Duplicate webhook must be ignored
	if err := svc.HandleWebhook(ctx, "sig", []byte("{}")); err != nil {
		t.Fatalf("duplicate webhook error = %v", err)
	}

	// Participant left restores activity status
	lk.webhookEvent = &livekit.WebhookEvent{
		Id:          "ev-leave",
		Event:       "participant_left",
		Room:        &livekit.Room{Name: "channel:ch1"},
		Participant: &livekit.ParticipantInfo{Identity: "u1"},
	}
	if err := svc.HandleWebhook(ctx, "sig", []byte("{}")); err != nil {
		t.Fatalf("participant_left error = %v", err)
	}
	if pres.lastActivity != "" {
		t.Fatalf("lastActivity = %q, want empty on leave", pres.lastActivity)
	}

	lk.webhookEvent = &livekit.WebhookEvent{
		Id:    "ev-finish",
		Event: "room_finished",
		Room:  &livekit.Room{Name: "channel:ch1"},
	}
	if err := svc.HandleWebhook(ctx, "sig", []byte("{}")); err != nil {
		t.Fatalf("room_finished error = %v", err)
	}

	// Track mute/unmute + screen share + egress lifecycle
	_ = rds.SetRoomMeta(ctx, "ch1", models.RoomMeta{GroupPublicID: "g1", ChannelPublicID: "c1", RoomName: "channel:ch1"})
	_ = rds.UpsertOccupant(ctx, "ch1", "g1", models.VoiceOccupant{UserID: "u1"})
	for idx, tc := range []struct {
		evName string
		src    livekit.TrackSource
	}{
		{"track_published", livekit.TrackSource_SCREEN_SHARE},
		{"track_unpublished", livekit.TrackSource_SCREEN_SHARE},
		{"track_published", livekit.TrackSource_CAMERA},
		{"track_unpublished", livekit.TrackSource_CAMERA},
		{"track_unpublished", livekit.TrackSource_MICROPHONE},
		{"track_published", livekit.TrackSource_MICROPHONE},
	} {
		lk.webhookEvent = &livekit.WebhookEvent{
			Id:          fmt.Sprintf("ev-tr-%d", idx),
			Event:       tc.evName,
			Room:        &livekit.Room{Name: "channel:ch1"},
			Participant: &livekit.ParticipantInfo{Identity: "u1"},
			Track:       &livekit.TrackInfo{Source: tc.src},
		}
		if err := svc.HandleWebhook(ctx, "sig", []byte("{}")); err != nil {
			t.Fatalf("%s error = %v", tc.evName, err)
		}
	}

	_, _ = pg.CreateRecording(ctx, models.CallRecording{
		PublicID: "rec-w1", SessionID: "s1", ChannelID: "ch1", EgressID: "eg-w1", Status: "starting", StartedBy: "u1",
	})
	lk.webhookEvent = &livekit.WebhookEvent{
		Id:         "ev-eg-start",
		Event:      "egress_started",
		EgressInfo: &livekit.EgressInfo{EgressId: "eg-w1", Status: livekit.EgressStatus_EGRESS_ACTIVE},
	}
	if err := svc.HandleWebhook(ctx, "sig", []byte("{}")); err != nil {
		t.Fatalf("egress_started error = %v", err)
	}

	lk.webhookEvent = &livekit.WebhookEvent{
		Id:    "ev-eg-end",
		Event: "egress_ended",
		EgressInfo: &livekit.EgressInfo{
			EgressId: "eg-w1",
			Status:   livekit.EgressStatus_EGRESS_COMPLETE,
			FileResults: []*livekit.FileInfo{{
				Filename: "recordings/ch1/rec.mp4",
				Size:     4096,
			}},
		},
	}
	if err := svc.HandleWebhook(ctx, "sig", []byte("{}")); err != nil {
		t.Fatalf("egress_ended error = %v", err)
	}

	// Egress failed + invalid signature + dm/grp room names
	lk.webhookEvent = &livekit.WebhookEvent{
		Id:         "ev-eg-fail",
		Event:      "egress_ended",
		EgressInfo: &livekit.EgressInfo{EgressId: "eg-w1", Status: livekit.EgressStatus_EGRESS_FAILED},
	}
	if err := svc.HandleWebhook(ctx, "sig", []byte("{}")); err != nil {
		t.Fatalf("egress_ended failed error = %v", err)
	}
	lk.webhookEvent = &livekit.WebhookEvent{
		Id:    "ev-grp-start",
		Event: "room_started",
		Room:  &livekit.Room{Name: "grp_g1_ch_ch1"},
	}
	if err := svc.HandleWebhook(ctx, "sig", []byte("{}")); err != nil {
		t.Fatalf("grp room_started error = %v", err)
	}
	lk.webhookErr = fmt.Errorf("bad sig")
	if err := svc.HandleWebhook(ctx, "bad", []byte("{}")); err == nil {
		t.Fatal("expected error on bad webhook signature")
	}
}
