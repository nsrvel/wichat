package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/wichat/wichat/backend/ms-core/internal/models"
	"github.com/wichat/wichat/backend/ms-core/internal/voice/repository/mock"
	"github.com/wichat/wichat/backend/ms-core/internal/voice/service"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

func TestGetQuotaUsageWarningThreshold(t *testing.T) {
	ctx := context.Background()
	st := mock.NewState()
	pg := mock.NewPgRepositoryMock(st)
	now := time.Now().UTC()

	_, _ = pg.UpsertActiveSession(ctx, models.CallSession{
		ID:                      "s1",
		PublicID:                "000000000000000000000001",
		ChannelID:               "ch1",
		RoomName:                "channel:ch1",
		StartedAt:               now,
		TotalParticipantSeconds: 4100 * 60,
	})

	svc := service.NewService(pg, nil, nil, &fakeLKClient{}, fakeChannelClient{
		resp: &userv1.AssertGroupChannelAccessResponse{},
	}, fakeUserClient{}, nil, nil, service.Options{
		MonthlyQuotaMins: 5000,
	}, nil)

	q, err := svc.GetQuotaUsage(ctx, "tok", "g1")
	if err != nil {
		t.Fatalf("GetQuotaUsage() error = %v", err)
	}
	if !q.WarningActive || q.UsedParticipantMinutes != 4100 {
		t.Fatalf("unexpected quota: %+v", q)
	}

	// Hard cap enforcement on IssueVoiceToken
	svcHard := service.NewService(pg, newTestRedis(t), nil, &fakeLKClient{}, fakeChannelClient{
		resp: &userv1.AssertGroupChannelAccessResponse{
			ChannelUuid: "ch-q",
			UserUuid:    "u1",
			ChannelType: "voice_only",
			Permissions: []string{"connect", "speak"},
		},
	}, fakeUserClient{}, nil, nil, service.Options{
		MonthlyQuotaMins: 1000,
		EnforceHardCap:   true,
	}, nil)
	if _, err := svcHard.IssueVoiceToken(ctx, &models.IssueVoiceTokenInput{
		AccessToken: "tok", GroupPublicID: "g1", ChannelPublicID: "c1",
	}); err == nil {
		t.Fatal("expected quota exceeded error when EnforceHardCap=true")
	}
}
