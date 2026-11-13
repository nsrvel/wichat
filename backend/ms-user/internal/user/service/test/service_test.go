package service_test

import (
	"context"
	"testing"

	groupmembermock "github.com/wichat/wichat/backend/ms-user/internal/group-member/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	userservice "github.com/wichat/wichat/backend/ms-user/internal/user/service"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	testSupport "github.com/wichat/wichat/backend/ms-user/pkg/test-support"
	"github.com/wichat/wichat/backend/ms-user/pkg/username"
	"github.com/wichat/wichat/backend/wi-shared/exception"
)

const testPassword = "correct-horse-battery"

func registerInput(email string) models.RegisterInput {
	handle, _ := username.BaseFromEmail(email)
	return models.RegisterInput{
		Email:       email,
		Username:    handle,
		Password:    testPassword,
		DisplayName: "Ada",
		BirthDate:   "2000-01-15",
	}
}

func codeOf(t *testing.T, err error) string {
	t.Helper()

	ex, ok := exception.As(err)
	if !ok {
		t.Fatalf("not an exception: %v", err)
	}

	return string(ex.Code)
}

func TestSetupStatusDefaultsBeforeWizard(t *testing.T) {
	st := usermock.NewState()
	_, _, _, userSvc, _, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	status, err := userSvc.SetupStatus(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if status.RegistrationMode != constants.RegistrationInviteOnly {
		t.Fatalf("%+v", status)
	}
}

func TestCheckUsernameAvailability(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, userSvc, _, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	res, err := userSvc.CheckUsername(ctx, "  Ada.Lovelace ")
	if err != nil || !res.Available || res.Normalized != "ada.lovelace" || res.Reason != "" {
		t.Fatalf("available: %+v %v", res, err)
	}

	resBad, err := userSvc.CheckUsername(ctx, "admin")
	if err != nil || resBad.Available || resBad.Reason != username.ReasonReserved {
		t.Fatalf("reserved: %+v %v", resBad, err)
	}

	_ = testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email:       "ada-check@example.com",
		Username:    "ada.lovelace",
		Password:    testPassword,
		DisplayName: "Ada",
		BirthDate:   "2000-01-15",
	})

	resTaken, err := userSvc.CheckUsername(ctx, "ada.lovelace")
	if err != nil || resTaken.Available || resTaken.Reason != username.ReasonTaken {
		t.Fatalf("taken: %+v %v", resTaken, err)
	}
}

func TestCompleteInstanceSetupRejectsNonOwner(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, userSvc, _, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, registerInput("owner-setup@example.com"))
	if err := completeInstanceSetup(t, userSvc, owner.AccessToken); err != nil {
		t.Fatal(err)
	}

	member := testSupport.RegisterSession(t, authSvc, notify, registerInput("member-setup@example.com"))
	err := userSvc.CompleteInstanceSetup(ctx, member.AccessToken, models.CompleteInstanceSetupInput{
		Name: "X", RegistrationMode: constants.RegistrationOpen,
	})
	if codeOf(t, err) != string(exception.MsgForbidden) {
		t.Fatalf("%v", err)
	}
}

func TestCompleteInstanceSetupCreatesFirstGroup(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, userSvc, _, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, registerInput("admin@example.com"))
	if err := completeInstanceSetup(t, userSvc, owner.AccessToken); err != nil {
		t.Fatal(err)
	}

	_, groups, _ := testSupport.TestRepos(st)
	group, err := groups.GetFirstGroup(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if group.Name != "WiChat" {
		t.Fatalf("group name: %q", group.Name)
	}
	members := groupmembermock.NewPgRepositoryMock(st)
	member, err := members.FindGroupMember(ctx, group.ID, owner.User.ID)
	if err != nil || member.Role != constants.GroupRoleOwner {
		t.Fatalf("membership: %+v err %v", member, err)
	}

	status, err := userSvc.SetupStatus(ctx)
	if err != nil || status.RegistrationMode != constants.RegistrationInviteOnly {
		t.Fatalf("status: %v %+v", err, status)
	}
}

func TestCompleteInstanceSetupOwnerOnly(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, userSvc, _, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, registerInput("admin@example.com"))
	if err := completeInstanceSetup(t, userSvc, owner.AccessToken); err != nil {
		t.Fatal(err)
	}

	err := userSvc.CompleteInstanceSetup(ctx, owner.AccessToken, models.CompleteInstanceSetupInput{
		Name:             "Again",
		RegistrationMode: constants.RegistrationOpen,
	})
	if codeOf(t, err) != string(exception.MsgSetupClosed) {
		t.Fatalf("second complete: %v", err)
	}
}

func completeInstanceSetup(t *testing.T, svc userservice.Service, token string) error {
	t.Helper()

	return svc.CompleteInstanceSetup(context.Background(), token, models.CompleteInstanceSetupInput{
		Name:             "WiChat",
		RegistrationMode: constants.RegistrationInviteOnly,
	})
}

func TestSearchUsersInGroup(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, userSvc, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, registerInput("owner-search@example.com"))
	created, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "Lookup")
	if err != nil {
		t.Fatal(err)
	}
	target := testSupport.RegisterSession(t, authSvc, notify, registerInput("findme@example.com"))

	hits, err := userSvc.SearchUsers(ctx, owner.AccessToken, "find", 10, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, h := range hits {
		if h.UserID == target.User.ID {
			found = true
		}
	}
	if !found {
		t.Fatalf("hits: %+v", hits)
	}
}

func TestSearchUsersEmptyQuery(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, userSvc, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, registerInput("search-empty@example.com"))
	created, _ := groupSvc.CreateGroup(ctx, owner.AccessToken, "Search")

	hits, err := userSvc.SearchUsers(ctx, owner.AccessToken, "  ", 10, created.ID)
	if err != nil || len(hits) != 0 {
		t.Fatalf("%+v %v", hits, err)
	}
}
