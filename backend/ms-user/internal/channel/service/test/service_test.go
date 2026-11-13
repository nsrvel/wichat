package service_test

import (
	"context"
	"testing"

	channelservice "github.com/wichat/wichat/backend/ms-user/internal/channel/service"
	groupmembermock "github.com/wichat/wichat/backend/ms-user/internal/group-member/repository/mock"
	grouprolemock "github.com/wichat/wichat/backend/ms-user/internal/group-role/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	testSupport "github.com/wichat/wichat/backend/ms-user/pkg/test-support"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

func TestListGroupChannelsRejectsInvalidGroupPublicID(t *testing.T) {
	t.Parallel()

	st := usermock.NewState()
	authSvc, notify, _, _, _, _ := testSupport.NewAuthService(t, st, nil)
	sess := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email:       "chan@example.com",
		Password:    "correct-horse-battery",
		DisplayName: "Chan",
		BirthDate:   "2000-01-15",
	})

	_, groups, sessions := testSupport.TestRepos(st)
	svc := channelservice.NewService(
		testSupport.ChannelRepo(st),
		groups,
		groupmembermock.NewPgRepositoryMock(st),
		grouprolemock.NewPgRepositoryMock(st),
		sessions,
		[]byte(testSupport.TestJWTSecret),
		logger.New(logger.Options{Service: "channel-test", Env: "test", Level: "error"}),
	)

	_, err := svc.ListGroupChannels(context.Background(), sess.AccessToken, "not-a-group-id")
	if err == nil {
		t.Fatal("expected error")
	}
	ex, ok := exception.As(err)
	if !ok || ex.Kind != exception.KindInvalidInput {
		t.Fatalf("got %v", err)
	}
}

func TestListGroupChannelsForMember(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "owner-ch@example.com", Password: "correct-horse-battery", DisplayName: "Owner", BirthDate: "2000-01-15",
	})
	created, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "Channel Guild")
	if err != nil {
		t.Fatal(err)
	}

	_, groups, sessions := testSupport.TestRepos(st)
	svc := channelservice.NewService(
		testSupport.ChannelRepo(st),
		groups,
		groupmembermock.NewPgRepositoryMock(st),
		grouprolemock.NewPgRepositoryMock(st),
		sessions,
		[]byte(testSupport.TestJWTSecret),
		logger.New(logger.Options{Service: "channel-test", Env: "test", Level: "error"}),
	)

	channels, err := svc.ListGroupChannels(ctx, owner.AccessToken, created.ID)
	if err != nil || len(channels) == 0 {
		t.Fatalf("channels %+v %v", channels, err)
	}

	outsider := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "outsider-ch@example.com", Password: "correct-horse-battery", DisplayName: "Out", BirthDate: "2000-01-15",
	})
	_, err = svc.ListGroupChannels(ctx, outsider.AccessToken, created.ID)
	if err == nil {
		t.Fatal("expected forbidden")
	}
	ex, ok := exception.As(err)
	if !ok || ex.Kind != exception.KindPermissionDenied {
		t.Fatalf("got %v", err)
	}
}

func TestAssertGroupChannelAccess(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "access@example.com", Password: "correct-horse-battery", DisplayName: "Owner", BirthDate: "2000-01-15",
	})
	created, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "Access")
	if err != nil {
		t.Fatal(err)
	}

	_, groups, sessions := testSupport.TestRepos(st)
	svc := channelservice.NewService(
		testSupport.ChannelRepo(st),
		groups,
		groupmembermock.NewPgRepositoryMock(st),
		grouprolemock.NewPgRepositoryMock(st),
		sessions,
		[]byte(testSupport.TestJWTSecret),
		logger.New(logger.Options{Service: "channel-test", Env: "test", Level: "error"}),
	)

	channels, err := svc.ListGroupChannels(ctx, owner.AccessToken, created.ID)
	if err != nil || len(channels) == 0 {
		t.Fatalf("channels %+v %v", channels, err)
	}

	access, err := svc.AssertGroupChannelAccess(ctx, owner.AccessToken, created.ID, channels[0].ID)
	if err != nil || access.UserID != owner.User.ID || access.ChannelID == "" {
		t.Fatalf("access: %+v %v", access, err)
	}

	_, err = svc.AssertGroupChannelAccess(ctx, owner.AccessToken, created.ID, "not-a-channel-id")
	if err == nil {
		t.Fatal("expected invalid channel")
	}
	ex, ok := exception.As(err)
	if !ok || ex.Kind != exception.KindInvalidInput {
		t.Fatalf("%v", err)
	}
}

func TestAssertGroupChannelAccessRejectsNonMember(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "owner-assert@example.com", Password: "correct-horse-battery", DisplayName: "Owner", BirthDate: "2000-01-15",
	})
	created, _ := groupSvc.CreateGroup(ctx, owner.AccessToken, "Restricted")
	outsider := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "outsider-assert@example.com", Password: "correct-horse-battery", DisplayName: "Out", BirthDate: "2000-01-15",
	})

	_, groups, sessions := testSupport.TestRepos(st)
	svc := channelservice.NewService(
		testSupport.ChannelRepo(st),
		groups,
		groupmembermock.NewPgRepositoryMock(st),
		grouprolemock.NewPgRepositoryMock(st),
		sessions,
		[]byte(testSupport.TestJWTSecret),
		logger.New(logger.Options{Service: "channel-test", Env: "test", Level: "error"}),
	)
	channels, _ := svc.ListGroupChannels(ctx, owner.AccessToken, created.ID)

	_, err := svc.AssertGroupChannelAccess(ctx, outsider.AccessToken, created.ID, channels[0].ID)
	if err == nil {
		t.Fatal("expected forbidden")
	}
	ex, ok := exception.As(err)
	if !ok || ex.Kind != exception.KindPermissionDenied {
		t.Fatalf("%v", err)
	}
}

func TestAssertGroupChannelAccessUnknownGroup(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, _, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "nogroup@example.com", Password: "correct-horse-battery", DisplayName: "A", BirthDate: "2000-01-15",
	})
	_, groups, sessions := testSupport.TestRepos(st)
	svc := channelservice.NewService(
		testSupport.ChannelRepo(st),
		groups,
		groupmembermock.NewPgRepositoryMock(st),
		grouprolemock.NewPgRepositoryMock(st),
		sessions,
		[]byte(testSupport.TestJWTSecret),
		logger.New(logger.Options{Service: "channel-test", Env: "test", Level: "error"}),
	)

	_, err := svc.AssertGroupChannelAccess(ctx, owner.AccessToken, "cccccccccccccccccccccccc", "dddddddddddddddddddddddd")
	if err == nil {
		t.Fatal("expected error")
	}
}
