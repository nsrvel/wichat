package test

import (
	"context"
	"testing"

	groupgrpc "github.com/wichat/wichat/backend/ms-user/internal/group/delivery/grpc"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	testSupport "github.com/wichat/wichat/backend/ms-user/pkg/test-support"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

func TestGroupGRPCServer(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "grp-grpc@example.com", Password: "correct-horse-battery", DisplayName: "Owner", BirthDate: "2000-01-01",
	})

	srv := groupgrpc.NewServer(groupSvc)
	res, err := srv.CreateGroup(ctx, &userv1.CreateGroupRequest{
		AccessToken: owner.AccessToken,
		Name:        "Test Server Group",
	})
	if err != nil || res.GetGroup().GetName() != "Test Server Group" {
		t.Fatalf("CreateGroup: %v, res: %+v", err, res)
	}

	mine, err := srv.ListMyGroups(ctx, &userv1.ListMyGroupsRequest{
		AccessToken: owner.AccessToken,
	})
	if err != nil || len(mine.GetGroups()) != 1 {
		t.Fatalf("ListMyGroups: %v, len: %d", err, len(mine.GetGroups()))
	}
}
