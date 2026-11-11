package mock_test

import (
	"context"
	"testing"
	"time"

	groupmock "github.com/wichat/wichat/backend/ms-user/internal/group/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

func TestGroupMockUpdatesAndMembersPage(t *testing.T) {
	st := usermock.NewState()
	st.Groups = []models.Group{{ID: "g1", PublicID: "a1b2c3d4e5f6789012345678", Name: "Old"}}
	st.Members = []models.GroupMember{{GroupID: "g1", UserID: "u1", Role: constants.GroupRoleOwner, CreatedAt: time.Now().UTC()}}

	repo := groupmock.NewPgRepositoryMock(st)
	ctx := context.Background()
	now := time.Now().UTC()

	if err := repo.UpdateGroupDisplayName(ctx, "g1", "New", now); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpdateGroupLogoURL(ctx, "g1", "https://cdn/x.png", now); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpdateGroupProfile(ctx, "g1", models.GroupProfilePatch{
		Name: "New", Description: "About", IsPrivate: true,
		BannerColor: "volcano", BannerURL: "", LogoURL: "https://cdn/x.png",
	}, now); err != nil {
		t.Fatal(err)
	}
	got, err := repo.FindGroupByID(ctx, "g1")
	if err != nil || got.Description != "About" || !got.IsPrivate || got.BannerColor != "volcano" {
		t.Fatalf("profile %+v %v", got, err)
	}
	if err := repo.SetRailPositions(ctx, "u1", []string{"g1"}); err != nil {
		t.Fatal(err)
	}
	if n, err := repo.CountGroupMembers(ctx, "g1"); err != nil || n != 1 {
		t.Fatalf("count %d %v", n, err)
	}
	members, _, err := repo.ListGroupMembersPage(ctx, "g1", 10, nil)
	if err != nil || len(members) != 1 {
		t.Fatalf("members %+v %v", members, err)
	}
}

func TestGroupMockJoinableAndCRUD(t *testing.T) {
	st := usermock.NewState()
	repo := groupmock.NewPgRepositoryMock(st)
	ctx := context.Background()
	now := time.Now().UTC()

	g := models.Group{ID: "g2", PublicID: "bbbbbbbbbbbbbbbbbbbbbbbb", Name: "Open", CreatedAt: now, UpdatedAt: now}
	m := models.GroupMember{GroupID: "g2", UserID: "u1", Role: constants.GroupRoleOwner, CreatedAt: now}
	if err := repo.CreateGroupWithOwner(ctx, g, m); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetFirstGroup(ctx); err != nil {
		t.Fatal(err)
	}
	joinable, err := repo.ListJoinableGroups(ctx, "u2")
	if err != nil || len(joinable) != 1 {
		t.Fatalf("%+v %v", joinable, err)
	}
	if err := repo.InsertGroupMember(ctx, models.GroupMember{GroupID: "g2", UserID: "u2", Role: constants.GroupRoleMember, CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := repo.DeleteGroupMember(ctx, "g2", "u2"); err != nil {
		t.Fatal(err)
	}
	if err := repo.DeleteGroup(ctx, "g2"); err != nil {
		t.Fatal(err)
	}
	if err := repo.DeleteGroup(ctx, "missing"); err == nil {
		t.Fatal("not found")
	}
}

func TestGroupMockNotFoundPaths(t *testing.T) {
	st := usermock.NewState()
	repo := groupmock.NewPgRepositoryMock(st)
	ctx := context.Background()

	if _, err := repo.FindGroupByID(ctx, "missing"); err == nil {
		t.Fatal("find group by id")
	}
	if _, err := repo.FindGroupByPublicID(ctx, "bbbbbbbbbbbbbbbbbbbbbbbb"); err == nil {
		t.Fatal("find by public id")
	}
	if _, err := repo.FindGroupMember(ctx, "g1", "u1"); err == nil {
		t.Fatal("find member")
	}
	if _, err := repo.GetFirstGroup(ctx); err == nil {
		t.Fatal("first group empty")
	}
}

func TestGroupMockListGroupsForUserWithRailOrder(t *testing.T) {
	st := usermock.NewState()
	now := time.Now().UTC()
	st.Groups = []models.Group{
		{ID: "g1", PublicID: "aaaaaaaaaaaaaaaaaaaaaaaa", Name: "Zed", CreatedAt: now, UpdatedAt: now},
		{ID: "g2", PublicID: "bbbbbbbbbbbbbbbbbbbbbbbb", Name: "Acme", CreatedAt: now, UpdatedAt: now},
	}
	st.Members = []models.GroupMember{
		{GroupID: "g1", UserID: "u1", Role: constants.GroupRoleOwner, CreatedAt: now, RailPosition: intPtr(1)},
		{GroupID: "g2", UserID: "u1", Role: constants.GroupRoleMember, CreatedAt: now, RailPosition: intPtr(0)},
	}

	repo := groupmock.NewPgRepositoryMock(st)
	ctx := context.Background()

	mine, err := repo.ListGroupsForUser(ctx, "u1")
	if err != nil || len(mine) != 2 {
		t.Fatalf("mine %+v %v", mine, err)
	}
	if mine[0].ID != "g2" || mine[0].CallerRole != constants.GroupRoleMember {
		t.Fatalf("order/role: %+v", mine[0])
	}
	if mine[1].CallerRole != constants.GroupRoleOwner {
		t.Fatalf("role: %+v", mine[1])
	}
}

func intPtr(v int) *int { return &v }

func TestGroupMockSetRailPositionsNotFound(t *testing.T) {
	st := usermock.NewState()
	now := time.Now().UTC()
	st.Groups = []models.Group{{ID: "g1", PublicID: "a1b2c3d4e5f6789012345678", Name: "One", CreatedAt: now, UpdatedAt: now}}
	st.Members = []models.GroupMember{{GroupID: "g1", UserID: "u1", Role: constants.GroupRoleOwner, CreatedAt: now}}

	repo := groupmock.NewPgRepositoryMock(st)
	if err := repo.SetRailPositions(context.Background(), "u1", []string{"missing"}); err == nil {
		t.Fatal("expected not found from SetRailPositions")
	}
}
