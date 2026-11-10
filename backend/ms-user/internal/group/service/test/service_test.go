package service_test

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	testSupport "github.com/wichat/wichat/backend/ms-user/pkg/test-support"
	"github.com/wichat/wichat/backend/wi-shared/exception"
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
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
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
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, registerInput("owner@example.com"))

	created, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "Acme")
	if err != nil || created.Name != "Acme" {
		t.Fatalf("create: %v %+v", err, created)
	}
	if created.Role != constants.GroupRoleOwner {
		t.Fatalf("create role: %q", created.Role)
	}
	if !regexp.MustCompile(`^[0-9a-f]{24}$`).MatchString(created.ID) {
		t.Fatalf("public id: %q", created.ID)
	}

	mine, err := groupSvc.ListMyGroups(ctx, owner.AccessToken)
	if err != nil || len(mine) != 1 || mine[0].ID != created.ID {
		t.Fatalf("mine: %v err %v", mine, err)
	}
	if mine[0].Role != constants.GroupRoleOwner {
		t.Fatalf("owner list role: %q", mine[0].Role)
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
	if mine[0].Role != constants.GroupRoleMember {
		t.Fatalf("member list role: %q", mine[0].Role)
	}
}

func TestCreateGroupRejectsInvalidName(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, registerInput("owner-invalid@example.com"))

	for _, name := range []string{"", "   ", strings.Repeat("a", 81)} {
		_, err := groupSvc.CreateGroup(ctx, owner.AccessToken, name)
		ex, ok := exception.As(err)
		if !ok || ex.Kind != exception.KindInvalidInput {
			t.Fatalf("name %q: expected invalid input, got %v", name, err)
		}
	}
}

func TestOwnerCannotLeaveGroup(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, registerInput("owner-leave@example.com"))
	created, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "Solo")
	if err != nil {
		t.Fatal(err)
	}

	leaveErr := groupSvc.LeaveGroup(ctx, owner.AccessToken, created.ID)
	if leaveErr == nil {
		t.Fatal("expected owner leave to be forbidden")
	}
	ex, ok := exception.As(leaveErr)
	if !ok || ex.Kind != exception.KindPermissionDenied {
		t.Fatalf("expected forbidden, got %v", leaveErr)
	}
}

func TestOwnerCanDeleteGroup(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, registerInput("owner-delete@example.com"))
	created, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "Solo")
	if err != nil {
		t.Fatal(err)
	}

	if err := groupSvc.DeleteGroup(ctx, owner.AccessToken, created.ID); err != nil {
		t.Fatalf("owner delete: %v", err)
	}

	mine, err := groupSvc.ListMyGroups(ctx, owner.AccessToken)
	if err != nil || len(mine) != 0 {
		t.Fatalf("owner should have no groups after delete: %+v err %v", mine, err)
	}
}

func TestMemberCanLeaveGroup(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, registerInput("owner-stay@example.com"))
	created, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "Team")
	if err != nil {
		t.Fatal(err)
	}

	member := testSupport.RegisterSession(t, authSvc, notify, registerInput("member-leave@example.com"))
	if err := groupSvc.JoinGroup(ctx, member.AccessToken, created.ID); err != nil {
		t.Fatal(err)
	}
	if err := groupSvc.LeaveGroup(ctx, member.AccessToken, created.ID); err != nil {
		t.Fatalf("member leave: %v", err)
	}

	mine, err := groupSvc.ListMyGroups(ctx, member.AccessToken)
	if err != nil || len(mine) != 0 {
		t.Fatalf("member should have no groups: %+v err %v", mine, err)
	}
}

func TestJoinGroupRejectsInvalidPublicID(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	member := testSupport.RegisterSession(t, authSvc, notify, registerInput("member@example.com"))
	if err := groupSvc.JoinGroup(ctx, member.AccessToken, "not-a-public-id"); err == nil {
		t.Fatal("expected invalid group_id")
	}
}
