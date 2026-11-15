package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
	activitymock "github.com/wichat/wichat/backend/ms-user/internal/user-activity-log/repository/mock"
	activityservice "github.com/wichat/wichat/backend/ms-user/internal/user-activity-log/service"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	testSupport "github.com/wichat/wichat/backend/ms-user/pkg/test-support"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

func TestRecordAndListMyActivityLogs(t *testing.T) {
	ctx := context.Background()
	st := usermock.NewState()
	authSvc, notify, _, _, _, _ := testSupport.NewAuthService(t, st, nil)
	_, _, sessions := testSupport.TestRepos(st)

	sess := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email:       "audit@example.com",
		DisplayName: "Audit User",
		Username:    "audit.user",
		Password:    "strong-password-123",
		BirthDate:   "1998-01-01",
		Locale:      "en",
	})

	actSt := activitymock.NewState()
	actRepo := activitymock.NewPgRepositoryMock(actSt)
	log := logger.New(logger.Options{Service: "activity-test", Env: "test", Level: "error"})
	svc := activityservice.NewService(actRepo, sessions, []byte(testSupport.TestJWTSecret), log)

	base := time.Now().UTC().Add(-5 * time.Minute)
	svc.Record(ctx, models.UserActivityLogEntry{
		UserID:     sess.User.ID,
		Category:   constants.ActivityCategoryProfile,
		Action:     constants.ActivityUserProfileUpdated,
		TargetType: constants.ActivityTargetUser,
		TargetID:   sess.User.ID,
		Changes:    map[string]models.ActivityFieldChange{"display_name": {Old: "Audit User", New: "Audit Two"}},
		CreatedAt:  base,
	})
	svc.Record(ctx, models.UserActivityLogEntry{
		UserID:     sess.User.ID,
		Category:   constants.ActivityCategoryGroup,
		Action:     constants.ActivityGroupJoined,
		TargetType: constants.ActivityTargetGroup,
		TargetID:   "g-1",
		Metadata:   map[string]any{"group_name": "Core"},
		CreatedAt:  base.Add(time.Minute),
	})

	page1, err := svc.ListMyActivityLogs(ctx, sess.AccessToken, "", "", 1, "")
	if err != nil || len(page1.Logs) != 1 || page1.Logs[0].Action != constants.ActivityGroupJoined || page1.NextCursor == "" {
		t.Fatalf("unexpected page1: %+v err=%v", page1, err)
	}

	page2, err := svc.ListMyActivityLogs(ctx, sess.AccessToken, "", "", 1, page1.NextCursor)
	if err != nil || len(page2.Logs) != 1 || page2.Logs[0].Action != constants.ActivityUserProfileUpdated || page2.NextCursor != "" {
		t.Fatalf("unexpected page2: %+v err=%v", page2, err)
	}

	if _, err := svc.ListMyActivityLogs(ctx, sess.AccessToken, "", "", 10, "invalid-cursor"); err == nil {
		t.Fatal("expected invalid cursor error")
	}
}
