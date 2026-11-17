package mock_test

import (
	"context"
	"testing"
	"time"

	"github.com/wichat/wichat/backend/ms-core/internal/models"
	"github.com/wichat/wichat/backend/ms-core/internal/voice/repository/mock"
)

func TestPgRepositoryMockLifecycle(t *testing.T) {
	ctx := context.Background()
	st := mock.NewState()
	repo := mock.NewPgRepositoryMock(st)

	now := time.Now().UTC()
	sess, err := repo.UpsertActiveSession(ctx, models.CallSession{
		ID:               "s1",
		PublicID:         "000000000000000000000001",
		ChannelID:        "ch1",
		RoomName:         "channel:ch1",
		StartedAt:        now,
		PeakParticipants: 1,
	})
	if err != nil {
		t.Fatalf("UpsertActiveSession() error = %v", err)
	}
	if sess.ID != "s1" {
		t.Fatalf("sess.ID = %q, want s1", sess.ID)
	}

	if err := repo.UpdateSessionPeak(ctx, "s1", 3); err != nil {
		t.Fatalf("UpdateSessionPeak() error = %v", err)
	}
	if err := repo.InsertParticipantJoin(ctx, models.CallParticipant{
		ID:        "p1",
		SessionID: "s1",
		UserID:    "u1",
		JoinedAt:  now,
	}); err != nil {
		t.Fatalf("InsertParticipantJoin() error = %v", err)
	}
	muted := true
	if err := repo.UpdateParticipantModeration(ctx, "s1", "u1", &muted, nil); err != nil {
		t.Fatalf("UpdateParticipantModeration() error = %v", err)
	}
	dur, err := repo.CloseParticipantLeave(ctx, "s1", "u1", now.Add(45*time.Second), "left")
	if err != nil || dur != 45 {
		t.Fatalf("CloseParticipantLeave() = (%d, %v), want (45, nil)", dur, err)
	}
	if err := repo.CloseSession(ctx, "s1", now.Add(60*time.Second), 0); err != nil {
		t.Fatalf("CloseSession() error = %v", err)
	}

	total, err := repo.SumParticipantSecondsBetween(ctx, now.Add(-time.Minute), now.Add(2*time.Minute))
	if err != nil || total != 45 {
		t.Fatalf("SumParticipantSecondsBetween() = (%d, %v), want (45, nil)", total, err)
	}

	_, err = repo.CreateRecording(ctx, models.CallRecording{
		ID:        "r1",
		PublicID:  "000000000000000000000002",
		SessionID: "s1",
		ChannelID: "ch1",
		EgressID:  "eg1",
		Status:    "starting",
		StartedBy: "u1",
		StartedAt: now,
	})
	if err != nil {
		t.Fatalf("CreateRecording() error = %v", err)
	}
	if _, err := repo.FindActiveRecordingByChannel(ctx, "ch1"); err != nil {
		t.Fatalf("FindActiveRecordingByChannel() error = %v", err)
	}
	if _, err := repo.UpdateRecordingByEgressID(ctx, "eg1", "completed", "recordings/r1.mp4", 1024, 45, nil, &now); err != nil {
		t.Fatalf("UpdateRecordingByEgressID() error = %v", err)
	}
	list, err := repo.ListRecordingsByChannel(ctx, "ch1", 10)
	if err != nil || len(list) != 1 || list[0].Status != "completed" {
		t.Fatalf("ListRecordingsByChannel() = (%+v, %v)", list, err)
	}
	if rec, err := repo.FindRecordingByEgressID(ctx, "eg1"); err != nil || rec.ID != "r1" {
		t.Fatalf("FindRecordingByEgressID() = (%+v, %v)", rec, err)
	}
	if _, err := repo.FindRecordingByEgressID(ctx, "missing"); err == nil {
		t.Fatal("expected error for missing egress")
	}
	if _, err := repo.FindActiveSessionByChannel(ctx, "missing"); err == nil {
		t.Fatal("expected error for missing active session")
	}
	if err := repo.CloseSession(ctx, "missing", now, 10); err == nil {
		t.Fatal("expected error for missing session close")
	}
	if err := repo.UpdateSessionPeak(ctx, "missing", 2); err == nil {
		t.Fatal("expected error for missing session peak")
	}
	if err := repo.UpdateParticipantModeration(ctx, "s1", "missing", &muted, &muted); err != nil {
		t.Fatalf("UpdateParticipantModeration(missing) err = %v", err)
	}
}
