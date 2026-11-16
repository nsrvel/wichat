package test

import (
	"context"
	"testing"

	groupmemberdelivery "github.com/wichat/wichat/backend/ms-user/internal/group-member/delivery/grpc"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	testSupport "github.com/wichat/wichat/backend/ms-user/pkg/test-support"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

func TestGroupMemberGRPCServer(t *testing.T) {
	ctx := context.Background()
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	memberSvc := testSupport.NewGroupMemberService(t, st)
	srv := groupmemberdelivery.NewServer(memberSvc)

	owner := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "owner-mem-grpc@example.com", Password: "correct-horse-battery", DisplayName: "Owner", BirthDate: "2000-01-01",
	})
	mem := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "mem-grpc@example.com", Password: "correct-horse-battery", DisplayName: "Mem", BirthDate: "2000-01-01",
	})

	group, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "Member Group")
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}

	joinResp, err := srv.JoinGroup(ctx, &userv1.JoinGroupRequest{
		AccessToken: mem.AccessToken,
		GroupId:     group.ID,
	})
	if err != nil || !joinResp.GetOk() {
		t.Fatalf("JoinGroup: %v", err)
	}

	listResp, err := srv.ListGroupMembers(ctx, &userv1.ListGroupMembersRequest{
		AccessToken: owner.AccessToken,
		GroupId:     group.ID,
		Limit:       10,
	})
	if err != nil || listResp.GetTotal() != 2 {
		t.Fatalf("ListGroupMembers: total=%d err=%v", listResp.GetTotal(), err)
	}

	timeoutResp, err := srv.TimeoutGroupMember(ctx, &userv1.TimeoutGroupMemberRequest{
		AccessToken:     owner.AccessToken,
		GroupId:         group.ID,
		UserId:          mem.User.ID,
		DurationSeconds: 60,
		Reason:          "spam",
	})
	if err != nil || timeoutResp.GetTimeoutUntil() == "" {
		t.Fatalf("TimeoutGroupMember set: resp=%v err=%v", timeoutResp, err)
	}

	clearResp, err := srv.TimeoutGroupMember(ctx, &userv1.TimeoutGroupMemberRequest{
		AccessToken:     owner.AccessToken,
		GroupId:         group.ID,
		UserId:          mem.User.ID,
		DurationSeconds: 0,
	})
	if err != nil || clearResp.GetTimeoutUntil() != "" {
		t.Fatalf("TimeoutGroupMember clear: resp=%v err=%v", clearResp, err)
	}

	kickResp, err := srv.KickGroupMember(ctx, &userv1.KickGroupMemberRequest{
		AccessToken: owner.AccessToken,
		GroupId:     group.ID,
		UserId:      mem.User.ID,
	})
	if err != nil || !kickResp.GetSuccess() {
		t.Fatalf("KickGroupMember: %v", err)
	}

	if _, err := srv.BanGroupMembers(ctx, &userv1.BanGroupMembersRequest{
		AccessToken: owner.AccessToken,
		GroupId:     group.ID,
		UserIds:     []string{mem.User.ID},
		Reason:      "spamming",
	}); err != nil {
		t.Fatalf("BanGroupMembers: %v", err)
	}
	bansResp, err := srv.ListGroupBans(ctx, &userv1.ListGroupBansRequest{
		AccessToken: owner.AccessToken,
		GroupId:     group.ID,
		Search:      "spam",
		Limit:       10,
	})
	if err != nil || bansResp.GetTotal() != 1 || len(bansResp.GetBans()) != 1 {
		t.Fatalf("ListGroupBans: %+v err=%v", bansResp, err)
	}
	if _, err := srv.UnbanGroupMember(ctx, &userv1.UnbanGroupMemberRequest{
		AccessToken: owner.AccessToken,
		GroupId:     group.ID,
		UserId:      mem.User.ID,
	}); err != nil {
		t.Fatalf("UnbanGroupMember: %v", err)
	}
}
