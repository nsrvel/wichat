package test

import (
	"context"
	"testing"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usergrpc "github.com/wichat/wichat/backend/ms-user/internal/user/delivery/grpc"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	testSupport "github.com/wichat/wichat/backend/ms-user/pkg/test-support"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

func TestUserGRPCServer(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, userSvc, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	srv := usergrpc.NewServer(userSvc)
	ctx := context.Background()

	statusRes, err := srv.SetupStatus(ctx, &userv1.SetupStatusRequest{})
	if err != nil || statusRes.GetRegistrationMode() == "" {
		t.Fatalf("SetupStatus: %v, res: %+v", err, statusRes)
	}

	checkRes, err := srv.CheckUsername(ctx, &userv1.CheckUsernameRequest{Username: "fresh.user"})
	if err != nil || !checkRes.GetAvailable() || checkRes.GetNormalized() != "fresh.user" {
		t.Fatalf("CheckUsername: %v, res: %+v", err, checkRes)
	}

	owner := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "user-grpc@example.com", Username: "user.grpc", Password: "correct-horse-battery", DisplayName: "Owner", BirthDate: "2000-01-01",
	})
	_ = testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "findme@example.com", Username: "findme", Password: "correct-horse-battery", DisplayName: "FindMe", BirthDate: "2000-01-01",
	})

	group, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "User Search Group")
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}

	searchRes, err := srv.SearchUsers(ctx, &userv1.SearchUsersRequest{
		AccessToken: owner.AccessToken,
		Query:       "findme",
		Limit:       10,
		GroupId:     group.ID,
	})
	if err != nil || len(searchRes.GetUsers()) != 1 {
		t.Fatalf("SearchUsers: %v, len: %d", err, len(searchRes.GetUsers()))
	}
}
