package service_test

import (
	"context"
	"testing"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	userservice "github.com/wichat/wichat/backend/ms-user/internal/user/service"
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

func codeOf(t *testing.T, err error) string {
	t.Helper()

	ex, ok := exception.As(err)
	if !ok {
		t.Fatalf("not an exception: %v", err)
	}

	return string(ex.Code)
}

func TestGuestDisabledByDefault(t *testing.T) {
	st := usermock.NewState()
	authSvc, _, _, _, _, _ := testSupport.NewAuthService(t, st)
	ctx := context.Background()

	_, err := authSvc.RegisterGuest(ctx, models.GuestRegisterInput{
		DisplayName: "Visitor",
		BirthDate:   "2000-01-15",
	})
	if codeOf(t, err) != string(exception.MsgGuestDisabled) {
		t.Fatalf("guest: %v", err)
	}
}

func TestCompleteInstanceSetupCreatesFirstGroup(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, users, userSvc, _, _ := testSupport.NewAuthService(t, st)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, registerInput("admin@example.com"))
	if err := completeInstanceSetup(t, userSvc, owner.AccessToken); err != nil {
		t.Fatal(err)
	}

	group, err := users.GetFirstGroup(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if group.Name != "WiChat" {
		t.Fatalf("group name: %q", group.Name)
	}
	member, err := users.FindGroupMember(ctx, group.ID, owner.User.ID)
	if err != nil || member.Role != constants.GroupRoleOwner {
		t.Fatalf("membership: %+v err %v", member, err)
	}

	status, err := userSvc.SetupStatus(ctx)
	if err != nil || status.RegistrationMode != constants.RegistrationInviteOnly {
		t.Fatalf("status: %v %+v", err, status)
	}
}

func TestGuestRegisterWhenEnabled(t *testing.T) {
	st := usermock.NewState()
	st.SetInstanceSettings(models.InstanceSettings{
		UsernameEnabled:  true,
		RegistrationMode: constants.RegistrationInviteOnly,
		GuestEnabled:     true,
	})
	authSvc, _, _, _, _, _ := testSupport.NewAuthService(t, st)
	ctx := context.Background()

	guest, err := authSvc.RegisterGuest(ctx, models.GuestRegisterInput{
		DisplayName: "Visitor",
		BirthDate:   "2000-01-15",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !guest.User.IsGuest {
		t.Fatalf("guest user: %+v", guest.User)
	}
}

func TestCompleteInstanceSetupOwnerOnly(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, userSvc, _, _ := testSupport.NewAuthService(t, st)
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
		UsernameEnabled:  true,
		GuestEnabled:     false,
	})
}
