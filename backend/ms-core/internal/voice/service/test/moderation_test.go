package service_test

import (
	"context"
	"testing"

	"github.com/wichat/wichat/backend/ms-core/internal/models"
	"github.com/wichat/wichat/backend/ms-core/internal/voice/repository/mock"
	"github.com/wichat/wichat/backend/ms-core/internal/voice/service"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

func TestModerateParticipant(t *testing.T) {
	ctx := context.Background()
	st := mock.NewState()
	pg := mock.NewPgRepositoryMock(st)
	rds := newTestRedis(t)
	lk := &fakeLKClient{}
	pub := &fakePublisher{}
	pres := &fakePresenceClient{}

	_ = rds.SetRoomMeta(ctx, "ch1", models.RoomMeta{GroupPublicID: "g1", ChannelPublicID: "c1", RoomName: "channel:ch1"})
	_ = rds.UpsertOccupant(ctx, "ch1", "g1", models.VoiceOccupant{UserID: "u2"})

	svc := service.NewService(pg, rds, nil, lk, fakeChannelClient{
		resp: &userv1.AssertGroupChannelAccessResponse{
			ChannelUuid: "ch1",
			UserUuid:    "admin-1",
			ChannelType: "voice_only",
			Permissions: []string{"mute_members", "move_members"},
		},
	}, fakeUserClient{}, pres, pub, service.Options{}, nil)

	occ, err := svc.ModerateParticipant(ctx, &models.ModerateParticipantInput{
		AccessToken:     "tok",
		GroupPublicID:   "g1",
		ChannelPublicID: "c1",
		TargetUserID:    "u2",
		Action:          "mute",
	})
	if err != nil || !occ.IsServerMuted {
		t.Fatalf("mute failed: occ=%+v err=%v", occ, err)
	}
	if len(pub.moderations) != 1 || pub.moderations[0].Action != "mute" {
		t.Fatalf("expected moderation event, got %+v", pub.moderations)
	}

	// Missing deafen_members permission
	if _, err := svc.ModerateParticipant(ctx, &models.ModerateParticipantInput{
		AccessToken:     "tok",
		GroupPublicID:   "g1",
		ChannelPublicID: "c1",
		TargetUserID:    "u2",
		Action:          "deafen",
	}); err == nil {
		t.Fatal("expected permission denied for deafen")
	} else if ex, ok := exception.As(err); !ok || ex.Kind != exception.KindPermissionDenied {
		t.Fatalf("expected PermissionDenied, got %v", err)
	}

	if _, err := svc.ModerateParticipant(ctx, &models.ModerateParticipantInput{
		AccessToken:     "tok",
		GroupPublicID:   "g1",
		ChannelPublicID: "c1",
		TargetUserID:    "u2",
		Action:          "disconnect",
	}); err != nil {
		t.Fatalf("disconnect failed: %v", err)
	}
	if len(lk.removedIDs) != 1 || lk.removedIDs[0] != "u2" {
		t.Fatalf("expected RemoveParticipant(u2), got %+v", lk.removedIDs)
	}

	// Admin can unmute, deafen, undeafen
	_ = rds.UpsertOccupant(ctx, "ch1", "g1", models.VoiceOccupant{UserID: "u3"})
	svcAdmin := service.NewService(pg, rds, nil, lk, fakeChannelClient{
		resp: &userv1.AssertGroupChannelAccessResponse{
			ChannelUuid: "ch1",
			UserUuid:    "admin-1",
			ChannelType: "voice_only",
			Permissions: []string{"administrator"},
		},
	}, fakeUserClient{}, pres, pub, service.Options{}, nil)
	for _, action := range []string{"unmute", "deafen", "undeafen"} {
		if _, err := svcAdmin.ModerateParticipant(ctx, &models.ModerateParticipantInput{
			AccessToken: "tok", GroupPublicID: "g1", ChannelPublicID: "c1", TargetUserID: "u3", Action: action,
		}); err != nil {
			t.Fatalf("%s failed: %v", action, err)
		}
	}
	if _, err := svcAdmin.ModerateParticipant(ctx, &models.ModerateParticipantInput{
		AccessToken: "tok", GroupPublicID: "g1", ChannelPublicID: "c1", TargetUserID: "u3", Action: "invalid_action",
	}); err == nil {
		t.Fatal("expected error for invalid moderation action")
	}
}
