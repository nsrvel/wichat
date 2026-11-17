package service_test

import (
	"context"
	"testing"

	"github.com/wichat/wichat/backend/ms-core/internal/voice/repository/mock"
	"github.com/wichat/wichat/backend/ms-core/internal/voice/service"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

type fakeMinioRepo struct{}

func (fakeMinioRepo) RecordingURL(key string) string {
	return "https://cdn.test/" + key
}

func TestRecordingLifecycle(t *testing.T) {
	ctx := context.Background()
	st := mock.NewState()
	pg := mock.NewPgRepositoryMock(st)
	rds := newTestRedis(t)
	lk := &fakeLKClient{}
	pub := &fakePublisher{}

	svc := service.NewService(pg, rds, fakeMinioRepo{}, lk, fakeChannelClient{
		resp: &userv1.AssertGroupChannelAccessResponse{
			ChannelUuid: "ch1",
			UserUuid:    "u1",
			ChannelType: "voice_only",
			Permissions: []string{"record_call", "read_message_history"},
		},
	}, fakeUserClient{}, nil, pub, service.Options{
		MinioEndpoint: "localhost:9000",
		MinioBucket:   "wichat-public",
	}, nil)

	rec, err := svc.StartRecording(ctx, "tok", "g1", "c1")
	if err != nil || rec.EgressID != "egress-1" {
		t.Fatalf("StartRecording() = (%+v, %v)", rec, err)
	}
	if len(pub.recordings) != 1 || !pub.recordings[0].IsRecording {
		t.Fatalf("expected recording.updated event, got %+v", pub.recordings)
	}

	// Second start while active must fail with Conflict
	if _, err := svc.StartRecording(ctx, "tok", "g1", "c1"); err == nil {
		t.Fatal("expected conflict on duplicate recording")
	} else if ex, ok := exception.As(err); !ok || ex.Kind != exception.KindConflict {
		t.Fatalf("expected Conflict, got %v", err)
	}

	stopped, err := svc.StopRecording(ctx, "tok", "g1", "c1", "")
	if err != nil || stopped.Status != "stopping" {
		t.Fatalf("StopRecording() = (%+v, %v)", stopped, err)
	}
	svcMissing := service.NewService(pg, rds, fakeMinioRepo{}, lk, fakeChannelClient{
		resp: &userv1.AssertGroupChannelAccessResponse{
			ChannelUuid: "ch-none",
			UserUuid:    "u1",
			ChannelType: "voice_only",
			Permissions: []string{"record_call"},
		},
	}, fakeUserClient{}, nil, pub, service.Options{}, nil)
	if _, err := svcMissing.StopRecording(ctx, "tok", "g1", "c-none", ""); err == nil {
		t.Fatal("expected error for non-active recording")
	}
	list, err := svc.ListChannelRecordings(ctx, "tok", "g1", "c1", 20)
	if err != nil || len(list) != 1 {
		t.Fatalf("ListChannelRecordings() = (%+v, %v)", list, err)
	}

	// Missing record_call permission denied
	svcNoPerm := service.NewService(pg, rds, fakeMinioRepo{}, lk, fakeChannelClient{
		resp: &userv1.AssertGroupChannelAccessResponse{
			ChannelUuid: "ch1",
			UserUuid:    "u1",
			ChannelType: "voice_only",
			Permissions: []string{"connect"},
		},
	}, fakeUserClient{}, nil, pub, service.Options{}, nil)
	if _, err := svcNoPerm.StartRecording(ctx, "tok", "g1", "c1"); err == nil {
		t.Fatal("expected permission denied on StartRecording")
	}
	if _, err := svcNoPerm.StopRecording(ctx, "tok", "g1", "c1", ""); err == nil {
		t.Fatal("expected permission denied on StopRecording")
	}
	if _, err := svcNoPerm.ListChannelRecordings(ctx, "tok", "g1", "c1", 10); err == nil {
		t.Fatal("expected permission denied on ListChannelRecordings")
	}

	// LiveKit egress start failure
	lkFail := &fakeLKClient{egressStartErr: exception.Internal(nil)}
	svcEgressFail := service.NewService(pg, newTestRedis(t), fakeMinioRepo{}, lkFail, fakeChannelClient{
		resp: &userv1.AssertGroupChannelAccessResponse{
			ChannelUuid: "ch-fail",
			UserUuid:    "u1",
			ChannelType: "voice_only",
			Permissions: []string{"record_call"},
		},
	}, fakeUserClient{}, nil, pub, service.Options{MinioEndpoint: "localhost:9000", MinioBucket: "b"}, nil)
	if _, err := svcEgressFail.StartRecording(ctx, "tok", "g1", "c-fail"); err == nil {
		t.Fatal("expected error when StartRoomCompositeEgress fails")
	}
}
