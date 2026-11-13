package test

import (
	"context"
	"testing"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	testSupport "github.com/wichat/wichat/backend/ms-user/pkg/test-support"
)

func TestGroupMemberService(t *testing.T) {
	ctx := context.Background()
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	memberSvc := testSupport.NewGroupMemberService(t, st)

	owner := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "owner-mem@example.com", Password: "correct-horse-battery", DisplayName: "Owner", BirthDate: "2000-01-01",
	})
	mem1 := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "mem1@example.com", Password: "correct-horse-battery", DisplayName: "Mem1", BirthDate: "2000-01-01",
	})
	mem2 := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "mem2@example.com", Password: "correct-horse-battery", DisplayName: "Mem2", BirthDate: "2000-01-01",
	})

	group, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "Member Group")
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}

	if err := memberSvc.JoinGroup(ctx, mem1.AccessToken, group.ID); err != nil {
		t.Fatalf("JoinGroup mem1: %v", err)
	}
	if err := memberSvc.JoinGroup(ctx, mem2.AccessToken, group.ID); err != nil {
		t.Fatalf("JoinGroup mem2: %v", err)
	}

	page, err := memberSvc.ListGroupMembers(ctx, owner.AccessToken, group.ID, 10, nil)
	if err != nil || page.Total != 3 {
		t.Fatalf("ListGroupMembers: total=%d err=%v", page.Total, err)
	}

	if err := memberSvc.KickGroupMember(ctx, mem1.AccessToken, group.ID, mem2.User.ID); err == nil {
		t.Fatalf("expected forbidden for non-owner kick")
	}

	if err := memberSvc.KickGroupMember(ctx, owner.AccessToken, group.ID, owner.User.ID); err == nil {
		t.Fatalf("expected invalid input for self kick")
	}

	if err := memberSvc.KickGroupMember(ctx, owner.AccessToken, group.ID, mem1.User.ID); err != nil {
		t.Fatalf("KickGroupMember: %v", err)
	}

	cnt, err := memberSvc.BanGroupMembers(ctx, owner.AccessToken, group.ID, []string{mem2.User.ID}, "bad", 3600)
	if err != nil || cnt != 1 {
		t.Fatalf("BanGroupMembers: cnt=%d err=%v", cnt, err)
	}

	if err := memberSvc.JoinGroup(ctx, mem2.AccessToken, group.ID); err == nil {
		t.Fatalf("expected permission denied for banned user")
	}

	if err := memberSvc.LeaveGroup(ctx, owner.AccessToken, group.ID); err == nil {
		t.Fatalf("expected forbidden for owner leaving")
	}
}
