package mock_test

import (
	"context"
	"testing"
	"time"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
	activitymock "github.com/wichat/wichat/backend/ms-user/internal/user-activity-log/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

func TestPgRepositoryMockInsertAndList(t *testing.T) {
	ctx := context.Background()
	st := activitymock.NewState()
	repo := activitymock.NewPgRepositoryMock(st)

	t1 := time.Date(2026, 11, 15, 8, 0, 0, 0, time.UTC)
	t2 := t1.Add(time.Minute)

	_ = repo.Insert(ctx, models.UserActivityLog{
		ID: "log-1", UserID: "u-1", ActorID: "u-1",
		Category: constants.ActivityCategoryProfile, Action: constants.ActivityUserProfileUpdated,
		CreatedAt: t1,
	})
	_ = repo.Insert(ctx, models.UserActivityLog{
		ID: "log-2", UserID: "u-1", ActorID: "u-1",
		Category: constants.ActivityCategoryGroup, Action: constants.ActivityGroupJoined,
		CreatedAt: t2,
	})
	_ = repo.Insert(ctx, models.UserActivityLog{
		ID: "log-3", UserID: "u-2", ActorID: "u-2",
		Category: constants.ActivityCategoryProfile, Action: constants.ActivityUserProfileUpdated,
		CreatedAt: t2,
	})

	all, err := repo.ListByUser(ctx, "u-1", "", "", 10, nil, "")
	if err != nil || len(all) != 2 || all[0].ID != "log-2" || all[1].ID != "log-1" {
		t.Fatalf("unexpected list all: %+v err=%v", all, err)
	}

	groupOnly, err := repo.ListByUser(ctx, "u-1", constants.ActivityCategoryGroup, "", 10, nil, "")
	if err != nil || len(groupOnly) != 1 || groupOnly[0].ID != "log-2" {
		t.Fatalf("unexpected category filter: %+v err=%v", groupOnly, err)
	}

	paged, err := repo.ListByUser(ctx, "u-1", "", "", 10, &t2, "log-2")
	if err != nil || len(paged) != 1 || paged[0].ID != "log-1" {
		t.Fatalf("unexpected cursor filter: %+v err=%v", paged, err)
	}
}
