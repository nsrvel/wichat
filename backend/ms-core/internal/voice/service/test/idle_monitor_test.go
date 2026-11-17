package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/wichat/wichat/backend/ms-core/internal/models"
	"github.com/wichat/wichat/backend/ms-core/internal/voice/repository/mock"
	"github.com/wichat/wichat/backend/ms-core/internal/voice/service"
)

func TestRunIdleCheckWarnAndKick(t *testing.T) {
	ctx := context.Background()
	st := mock.NewState()
	pg := mock.NewPgRepositoryMock(st)
	rds := newTestRedis(t)
	lk := &fakeLKClient{}
	pub := &fakePublisher{}
	pres := &fakePresenceClient{}

	now := time.Now().UTC()
	_ = rds.SetRoomMeta(ctx, "ch1", models.RoomMeta{
		SessionID:       "s1",
		GroupPublicID:   "g1",
		ChannelPublicID: "c1",
		RoomName:        "channel:ch1",
	})
	_ = rds.UpsertOccupant(ctx, "ch1", "g1", models.VoiceOccupant{
		UserID:       "u-warn",
		LastActiveAt: now.Add(-29 * time.Minute).Unix(),
	})
	_ = rds.UpsertOccupant(ctx, "ch1", "g1", models.VoiceOccupant{
		UserID:       "u-kick",
		LastActiveAt: now.Add(-31 * time.Minute).Unix(),
	})

	svc := service.NewService(pg, rds, nil, lk, fakeChannelClient{}, fakeUserClient{}, pres, pub, service.Options{}, nil)
	if err := svc.RunIdleCheck(ctx); err != nil {
		t.Fatalf("RunIdleCheck() error = %v", err)
	}

	if len(pub.idleWarns) != 1 || pub.idleWarns[0].UserID != "u-warn" {
		t.Fatalf("expected 1 idle warning for u-warn, got %+v", pub.idleWarns)
	}
	if len(lk.removedIDs) != 1 || lk.removedIDs[0] != "u-kick" {
		t.Fatalf("expected u-kick to be removed, got %+v", lk.removedIDs)
	}
}
