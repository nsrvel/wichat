package test

import (
	"context"
	"testing"

	channelmock "github.com/wichat/wichat/backend/ms-user/internal/channel/repository/mock"
	groupmembermock "github.com/wichat/wichat/backend/ms-user/internal/group-member/repository/mock"
	grouprolemock "github.com/wichat/wichat/backend/ms-user/internal/group-role/repository/mock"
	invitegrpc "github.com/wichat/wichat/backend/ms-user/internal/invite/delivery/grpc"
	invitemock "github.com/wichat/wichat/backend/ms-user/internal/invite/repository/mock"
	inviteservice "github.com/wichat/wichat/backend/ms-user/internal/invite/service"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	testSupport "github.com/wichat/wichat/backend/ms-user/pkg/test-support"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

func TestInviteGRPCServer(t *testing.T) {
	ctx := context.Background()
	st := usermock.NewState()
	authSvc, notify, users, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	_, groups, sessions := testSupport.TestRepos(st)
	log := logger.New(logger.Options{Service: "invite-grpc-test", Env: "test", Level: "error"})

	invSvc := inviteservice.NewService(
		groups,
		groupmembermock.NewPgRepositoryMock(st),
		grouprolemock.NewPgRepositoryMock(st),
		channelmock.NewPgRepositoryMock(st),
		invitemock.NewPgRepositoryMock(st),
		users,
		sessions,
		notify,
		[]byte(testSupport.TestJWTSecret),
		"http://localhost:3000",
		"http://localhost:9000/wichat",
		log,
	)
	srv := invitegrpc.NewServer(invSvc)

	owner := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "inv-grpc@example.com", Password: "correct-horse-battery", DisplayName: "Owner", BirthDate: "2000-01-01",
	})
	group, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "Invite Group")
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}

	link, err := srv.GetGroupInviteLink(ctx, &userv1.GetGroupInviteLinkRequest{
		AccessToken: owner.AccessToken,
		GroupId:     group.ID,
	})
	if err != nil || link.GetToken() == "" {
		t.Fatalf("GetGroupInviteLink: %v, token=%q", err, link.GetToken())
	}

	preview, err := srv.PreviewGroupInvite(ctx, &userv1.PreviewGroupInviteRequest{
		Token: link.GetToken(),
	})
	if err != nil || !preview.GetValid() || preview.GetGroupName() != "Invite Group" {
		t.Fatalf("PreviewGroupInvite: %v, preview=%+v", err, preview)
	}
}
