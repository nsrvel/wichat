package repository_test

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/wichat/wichat/backend/ms-core/internal/models"
	"github.com/wichat/wichat/backend/ms-core/internal/voice/repository"
)

func TestRedisOccupancyLifecycle(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = rdb.Close() }()

	ctx := context.Background()
	repo := repository.NewRedisOccupancy(rdb)

	meta := models.RoomMeta{
		SessionID:       "s1",
		GroupPublicID:   "g1",
		ChannelPublicID: "c1",
		RoomName:        "channel:ch1",
		IsRecording:     true,
	}
	if err := repo.SetRoomMeta(ctx, "ch1", meta); err != nil {
		t.Fatalf("SetRoomMeta() error = %v", err)
	}
	gotMeta, err := repo.GetRoomMeta(ctx, "ch1")
	if err != nil || !gotMeta.IsRecording || gotMeta.SessionID != "s1" {
		t.Fatalf("GetRoomMeta() = (%+v, %v)", gotMeta, err)
	}

	if err := repo.UpsertOccupant(ctx, "ch1", "g1", models.VoiceOccupant{
		UserID:     "u1",
		IsCameraOn: true,
	}); err != nil {
		t.Fatalf("UpsertOccupant() error = %v", err)
	}

	if count, err := repo.CountOccupants(ctx, "ch1"); err != nil || count != 1 {
		t.Fatalf("CountOccupants() = (%d, %v), want 1", count, err)
	}
	if vidCount, err := repo.CountVideoPublishers(ctx, "ch1"); err != nil || vidCount != 1 {
		t.Fatalf("CountVideoPublishers() = (%d, %v), want 1", vidCount, err)
	}
	if err := repo.MarkOccupantIdleWarned(ctx, "ch1", "u1", true); err != nil {
		t.Fatalf("MarkOccupantIdleWarned() error = %v", err)
	}
	if err := repo.TouchOccupantActivity(ctx, "ch1", "u1", 12345); err != nil {
		t.Fatalf("TouchOccupantActivity() error = %v", err)
	}
	occs, err := repo.ListOccupants(ctx, "ch1")
	if err != nil || len(occs) != 1 || occs[0].IdleWarned || occs[0].LastActiveAt != 12345 {
		t.Fatalf("ListOccupants() = (%+v, %v)", occs, err)
	}

	for i := 0; i < 10; i++ {
		ok, err := repo.AllowTokenRequest(ctx, "u1", 10)
		if err != nil || !ok {
			t.Fatalf("AllowTokenRequest(%d) = (%v, %v), want true", i, ok, err)
		}
	}
	if ok, _ := repo.AllowTokenRequest(ctx, "u1", 10); ok {
		t.Fatal("11th token request should be rate limited")
	}

	if first, err := repo.MarkWebhookProcessed(ctx, "ev1"); err != nil || !first {
		t.Fatalf("first MarkWebhookProcessed = (%v, %v)", first, err)
	}
	if second, err := repo.MarkWebhookProcessed(ctx, "ev1"); err != nil || second {
		t.Fatalf("second MarkWebhookProcessed = (%v, %v)", second, err)
	}

	if err := repo.RemoveOccupant(ctx, "ch1", "g1", "u1"); err != nil {
		t.Fatalf("RemoveOccupant() error = %v", err)
	}
	if err := repo.ClearRoom(ctx, "ch1", "g1"); err != nil {
		t.Fatalf("ClearRoom() error = %v", err)
	}
}
