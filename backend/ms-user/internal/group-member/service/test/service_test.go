package test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	testSupport "github.com/wichat/wichat/backend/ms-user/pkg/test-support"
)

type failingPresenceReader struct{}

func (failingPresenceReader) BatchGetEffectivePresence(context.Context, []string) (map[string]models.UserPresenceSnapshot, error) {
	return nil, errors.New("presence store unavailable")
}

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

	if err := memberSvc.KickGroupMember(ctx, mem1.AccessToken, group.ID, mem2.User.ID, ""); err == nil {
		t.Fatalf("expected forbidden for non-owner kick")
	}

	if err := memberSvc.KickGroupMember(ctx, owner.AccessToken, group.ID, owner.User.ID, ""); err == nil {
		t.Fatalf("expected invalid input for self kick")
	}

	if _, err := memberSvc.TimeoutGroupMember(ctx, owner.AccessToken, group.ID, mem1.User.ID, -1, ""); err == nil {
		t.Fatalf("expected invalid input for negative timeout duration")
	}
	if _, err := memberSvc.TimeoutGroupMember(ctx, mem1.AccessToken, group.ID, mem2.User.ID, 60, "spam"); err == nil {
		t.Fatalf("expected forbidden for non-owner timeout")
	}
	until, err := memberSvc.TimeoutGroupMember(ctx, owner.AccessToken, group.ID, mem1.User.ID, 60, "spam")
	if err != nil || until == nil {
		t.Fatalf("TimeoutGroupMember set: until=%v err=%v", until, err)
	}
	if cleared, err := memberSvc.TimeoutGroupMember(ctx, owner.AccessToken, group.ID, mem1.User.ID, 0, "forgiven"); err != nil || cleared != nil {
		t.Fatalf("TimeoutGroupMember clear: cleared=%v err=%v", cleared, err)
	}

	if err := memberSvc.KickGroupMember(ctx, owner.AccessToken, group.ID, mem1.User.ID, "spam"); err != nil {
		t.Fatalf("KickGroupMember: %v", err)
	}

	if _, err := memberSvc.BanGroupMembers(ctx, owner.AccessToken, group.ID, []string{mem2.User.ID}, "bad", 0, -1); err == nil {
		t.Fatalf("expected invalid input for negative ban duration")
	}
	cnt, err := memberSvc.BanGroupMembers(ctx, owner.AccessToken, group.ID, []string{mem2.User.ID}, "bad", 3600, 0)
	if err != nil || cnt != 1 {
		t.Fatalf("BanGroupMembers: cnt=%d err=%v", cnt, err)
	}

	if err := memberSvc.JoinGroup(ctx, mem2.AccessToken, group.ID); err == nil {
		t.Fatalf("expected permission denied for banned user")
	}

	bansPage, err := memberSvc.ListGroupBans(ctx, owner.AccessToken, group.ID, "bad", 10, nil)
	if err != nil || bansPage.Total != 1 || len(bansPage.Bans) != 1 || bansPage.Bans[0].UserID != mem2.User.ID {
		t.Fatalf("ListGroupBans: %+v err=%v", bansPage, err)
	}
	if err := memberSvc.UnbanGroupMember(ctx, owner.AccessToken, group.ID, mem2.User.ID); err != nil {
		t.Fatalf("UnbanGroupMember: %v", err)
	}
	if err := memberSvc.JoinGroup(ctx, mem2.AccessToken, group.ID); err != nil {
		t.Fatalf("expected unbanned user to rejoin, got %v", err)
	}

	if err := memberSvc.LeaveGroup(ctx, owner.AccessToken, group.ID); err == nil {
		t.Fatalf("expected forbidden for owner leaving")
	}
}

func TestListGroupMembersWithoutPresenceEnrichment(t *testing.T) {
	ctx := context.Background()
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	memberSvc := testSupport.NewGroupMemberService(t, st).WithPresence(failingPresenceReader{})

	owner := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "owner-presence-degrade@example.com", Password: "correct-horse-battery", DisplayName: "Owner", BirthDate: "2000-01-01",
	})
	group, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "Degrade Members")
	require.NoError(t, err)

	page, err := memberSvc.ListGroupMembers(ctx, owner.AccessToken, group.ID, 10, nil)
	require.NoError(t, err)
	require.Equal(t, 1, page.Total)
	require.Len(t, page.Members, 1)
	require.False(t, page.Members[0].IsOnline)
	require.Equal(t, models.PresenceStatusOffline, page.Members[0].PresenceStatus)
}
