package mock_test

import (
	"context"
	"testing"
	"time"

	groupactivitymock "github.com/wichat/wichat/backend/ms-user/internal/group-activity-log/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

func TestPgRepositoryMockInsertAndList(t *testing.T) {
	ctx := context.Background()
	repo := groupactivitymock.NewPgRepositoryMock()

	t1 := time.Date(2026, 11, 15, 8, 0, 0, 0, time.UTC)
	t2 := t1.Add(time.Minute)

	_ = repo.Insert(ctx, models.GroupActivityLog{
		ID: "glog-1", GroupID: "g-1", ActorID: "u-1",
		Category: constants.GroupAuditCategoryProfile, Action: constants.GroupAuditGroupProfileUpdated,
		CreatedAt: t1,
	})
	_ = repo.Insert(ctx, models.GroupActivityLog{
		ID: "glog-2", GroupID: "g-1", ActorID: "u-2",
		Category: constants.GroupAuditCategoryModeration, Action: constants.GroupAuditMemberKicked,
		CreatedAt: t2,
	})
	_ = repo.Insert(ctx, models.GroupActivityLog{
		ID: "glog-3", GroupID: "g-2", ActorID: "u-1",
		Category: constants.GroupAuditCategoryRole, Action: constants.GroupAuditRoleCreated,
		CreatedAt: t2,
	})

	all, err := repo.ListByGroup(ctx, models.GroupActivityLogFilter{GroupID: "g-1"}, 10, nil, "")
	if err != nil || len(all) != 2 || all[0].ID != "glog-2" || all[1].ID != "glog-1" {
		t.Fatalf("unexpected list all: %+v err=%v", all, err)
	}

	byActor, err := repo.ListByGroup(ctx, models.GroupActivityLogFilter{GroupID: "g-1", ActorID: "u-1"}, 10, nil, "")
	if err != nil || len(byActor) != 1 || byActor[0].ID != "glog-1" {
		t.Fatalf("unexpected actor filter: %+v err=%v", byActor, err)
	}

	byAction, err := repo.ListByGroup(ctx, models.GroupActivityLogFilter{GroupID: "g-1", Action: constants.GroupAuditMemberKicked}, 10, nil, "")
	if err != nil || len(byAction) != 1 || byAction[0].ID != "glog-2" {
		t.Fatalf("unexpected action filter: %+v err=%v", byAction, err)
	}

	paged, err := repo.ListByGroup(ctx, models.GroupActivityLogFilter{GroupID: "g-1"}, 10, &t2, "glog-2")
	if err != nil || len(paged) != 1 || paged[0].ID != "glog-1" {
		t.Fatalf("unexpected cursor filter: %+v err=%v", paged, err)
	}
}
