package service_test

import (
	"context"
	"testing"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	testSupport "github.com/wichat/wichat/backend/ms-user/pkg/test-support"
)

const testPassword = "correct-horse-battery"

func registerInput(email string) models.RegisterInput {
	return models.RegisterInput{
		Email:       email,
		Password:    testPassword,
		DisplayName: "Ada",
		BirthDate:   "2000-01-15",
	}
}

func TestNewUserHasNoGroups(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st)
	ctx := context.Background()

	first := testSupport.RegisterSession(t, authSvc, notify, registerInput("first@example.com"))
	_ = testSupport.RegisterSession(t, authSvc, notify, registerInput("second@example.com"))

	mine, err := groupSvc.ListMyGroups(ctx, first.AccessToken)
	if err != nil || len(mine) != 0 {
		t.Fatalf("first groups: %v %v", mine, err)
	}
}

func TestCreateJoinGroupFlow(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, registerInput("owner@example.com"))

	created, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "Acme")
	if err != nil || created.Name != "Acme" {
		t.Fatalf("create: %v %+v", err, created)
	}

	mine, err := groupSvc.ListMyGroups(ctx, owner.AccessToken)
	if err != nil || len(mine) != 1 || mine[0].ID != created.ID {
		t.Fatalf("mine: %v err %v", mine, err)
	}

	member := testSupport.RegisterSession(t, authSvc, notify, registerInput("member@example.com"))

	joinable, err := groupSvc.ListJoinableGroups(ctx, member.AccessToken)
	if err != nil || len(joinable) != 1 || joinable[0].ID != created.ID {
		t.Fatalf("joinable: %+v err %v", joinable, err)
	}

	if err := groupSvc.JoinGroup(ctx, member.AccessToken, created.ID); err != nil {
		t.Fatal(err)
	}

	joinable, err = groupSvc.ListJoinableGroups(ctx, member.AccessToken)
	if err != nil || len(joinable) != 0 {
		t.Fatalf("joinable after join: %+v", joinable)
	}

	mine, err = groupSvc.ListMyGroups(ctx, member.AccessToken)
	if err != nil || len(mine) != 1 {
		t.Fatalf("member mine: %+v err %v", mine, err)
	}
}
