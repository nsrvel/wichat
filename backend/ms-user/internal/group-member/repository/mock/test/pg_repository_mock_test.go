package test

import (
	"context"
	"testing"
	"time"

	groupmembermock "github.com/wichat/wichat/backend/ms-user/internal/group-member/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

func TestPgRepositoryMock(t *testing.T) {
	ctx := context.Background()
	st := usermock.NewState()
	m := groupmembermock.NewPgRepositoryMock(st)

	groupID := "grp-1"
	ownerID := "usr-owner"
	memID := "usr-mem"
	now := time.Now().UTC()

	st.Users = append(st.Users,
		models.User{ID: ownerID, DisplayName: "Owner", Email: "o@e.com"},
		models.User{ID: memID, DisplayName: "Member", Email: "m@e.com"},
	)

	_ = m.InsertGroupMember(ctx, models.GroupMember{GroupID: groupID, UserID: ownerID, Role: constants.GroupRoleOwner, CreatedAt: now})
	_ = m.InsertGroupMember(ctx, models.GroupMember{GroupID: groupID, UserID: memID, Role: constants.GroupRoleMember, CreatedAt: now.Add(time.Second)})

	count, err := m.CountGroupMembers(ctx, groupID)
	if err != nil || count != 2 {
		t.Fatalf("expected 2 members, got %d", count)
	}

	page, next, err := m.ListGroupMembersPage(ctx, groupID, 1, nil)
	if err != nil || len(page) != 1 || next == nil {
		t.Fatalf("expected 1 member and cursor, got %d, next=%v", len(page), next)
	}

	page2, next2, err := m.ListGroupMembersPage(ctx, groupID, 10, next)
	if err != nil || len(page2) != 1 || next2 != nil {
		t.Fatalf("expected 1 remaining member, got %d", len(page2))
	}

	bannedCount, err := m.BanMembers(ctx, groupID, []string{ownerID, memID}, ownerID, "spam", 0)
	if err != nil || bannedCount != 1 {
		t.Fatalf("expected 1 banned (owner skipped), got %d, err %v", bannedCount, err)
	}

	banned, _ := m.IsMemberBanned(ctx, groupID, memID)
	if !banned {
		t.Fatalf("expected member to be banned")
	}
}
