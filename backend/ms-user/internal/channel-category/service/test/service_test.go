package service_test

import (
	"context"
	"testing"

	categorymock "github.com/wichat/wichat/backend/ms-user/internal/channel-category/repository/mock"
	categoryservice "github.com/wichat/wichat/backend/ms-user/internal/channel-category/service"
	channelservice "github.com/wichat/wichat/backend/ms-user/internal/channel/service"
	groupmembermock "github.com/wichat/wichat/backend/ms-user/internal/group-member/repository/mock"
	grouprolemock "github.com/wichat/wichat/backend/ms-user/internal/group-role/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	testSupport "github.com/wichat/wichat/backend/ms-user/pkg/test-support"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

func newCategoryAndChannelServices(st *usermock.State) (categoryservice.Service, channelservice.Service) {
	_, groups, sessions := testSupport.TestRepos(st)
	members := groupmembermock.NewPgRepositoryMock(st)
	roles := grouprolemock.NewPgRepositoryMock(st)
	log := logger.New(logger.Options{Service: "category-test", Env: "test", Level: "error"})
	catRepo := categorymock.NewPgRepositoryMock(st)
	catSvc := categoryservice.NewService(catRepo, groups, members, roles, sessions, []byte(testSupport.TestJWTSecret), log)
	chSvc := channelservice.NewService(testSupport.ChannelRepo(st), catRepo, groups, members, roles, sessions, []byte(testSupport.TestJWTSecret), log)

	return catSvc, chSvc
}

func TestCreateGroupCategoryLifecycleAndPermissions(t *testing.T) {
	t.Parallel()

	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "cat-owner@example.com", Password: "correct-horse-battery", DisplayName: "CatOwner", BirthDate: "2000-01-15",
	})
	member := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "cat-member@example.com", Password: "correct-horse-battery", DisplayName: "CatMember", BirthDate: "2000-01-15",
	})
	created, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "Category Guild")
	if err != nil {
		t.Fatal(err)
	}

	st.Members = append(st.Members, models.GroupMember{
		GroupID: st.Groups[0].ID,
		UserID:  member.User.ID,
		Role:    "member",
	})

	catSvc, chSvc := newCategoryAndChannelServices(st)

	if _, err := catSvc.CreateGroupCategory(ctx, owner.AccessToken, created.ID, "   ", false); err == nil {
		t.Fatal("expected invalid input for empty name")
	}

	if _, err := catSvc.CreateGroupCategory(ctx, member.AccessToken, created.ID, "Blocked", false); err == nil {
		t.Fatal("expected permission denied for regular member")
	}

	cat1, err := catSvc.CreateGroupCategory(ctx, owner.AccessToken, created.ID, "  Engineering  ", false)
	if err != nil || cat1.Name != "Engineering" || cat1.IsPrivate {
		t.Fatalf("cat1 %+v err %v", cat1, err)
	}
	cat2, err := catSvc.CreateGroupCategory(ctx, owner.AccessToken, created.ID, "Leadership", true)
	if err != nil || cat2.Name != "Leadership" || !cat2.IsPrivate || cat2.Position <= cat1.Position {
		t.Fatalf("cat2 %+v err %v", cat2, err)
	}

	_, listedCats, err := chSvc.ListGroupChannels(ctx, owner.AccessToken, created.ID)
	if err != nil || len(listedCats) != 2 {
		t.Fatalf("listedCats %+v err %v", listedCats, err)
	}

	st.Roles = append(st.Roles, models.GroupRole{
		ID:          "role-mod",
		GroupID:     st.Groups[0].ID,
		Name:        "Mod",
		Permissions: []string{"manage_channels"},
	})
	st.MemberRoles = append(st.MemberRoles, models.GroupMemberRoleRow{
		GroupID: st.Groups[0].ID,
		UserID:  member.User.ID,
		RoleID:  "role-mod",
	})
	cat3, err := catSvc.CreateGroupCategory(ctx, member.AccessToken, created.ID, "ModOps", false)
	if err != nil || cat3.Name != "ModOps" {
		t.Fatalf("cat3 %+v err %v", cat3, err)
	}

	if err := catSvc.ReorderChannelCategories(ctx, owner.AccessToken, created.ID, []string{cat3.ID, cat1.ID, cat2.ID}); err != nil {
		t.Fatalf("ReorderChannelCategories: %v", err)
	}
	_, reordered, err := chSvc.ListGroupChannels(ctx, owner.AccessToken, created.ID)
	if err != nil || len(reordered) != 3 || reordered[0].ID != cat3.ID {
		t.Fatalf("expected %s first after reorder, got %+v err=%v", cat3.ID, reordered, err)
	}
}

func TestUpdateAndDeleteGroupCategory(t *testing.T) {
	t.Parallel()

	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "owner-cat-edit@example.com", Password: "correct-horse-battery", DisplayName: "Owner", BirthDate: "2000-01-15",
	})
	created, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "Category Edit Group")
	if err != nil {
		t.Fatal(err)
	}

	catSvc, chSvc := newCategoryAndChannelServices(st)

	cat, err := catSvc.CreateGroupCategory(ctx, owner.AccessToken, created.ID, "Text Channels", false)
	if err != nil {
		t.Fatal(err)
	}
	ch, err := chSvc.CreateGroupChannel(ctx, owner.AccessToken, created.ID, "general", "text_only", cat.ID, false)
	if err != nil {
		t.Fatal(err)
	}

	newName := "Discussion"
	isPrivate := true
	updated, err := catSvc.UpdateGroupCategory(ctx, owner.AccessToken, created.ID, cat.ID, &newName, &isPrivate)
	if err != nil {
		t.Fatalf("UpdateGroupCategory: %v", err)
	}
	if updated.Name != "Discussion" || !updated.IsPrivate {
		t.Fatalf("unexpected updated category: %+v", updated)
	}

	if err := catSvc.DeleteGroupCategory(ctx, owner.AccessToken, created.ID, cat.ID); err != nil {
		t.Fatalf("DeleteGroupCategory: %v", err)
	}

	channels, categories, err := chSvc.ListGroupChannels(ctx, owner.AccessToken, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range categories {
		if c.ID == cat.ID {
			t.Fatalf("expected category %s to be deleted", cat.ID)
		}
	}
	found := false
	for _, item := range channels {
		if item.ID == ch.ID {
			found = true
			if item.CategoryID != "" {
				t.Fatalf("expected channel %s to become uncategorized, got %q", ch.ID, item.CategoryID)
			}
		}
	}
	if !found {
		t.Fatalf("expected channel %s to remain after category deletion", ch.ID)
	}
}

func TestCategoryPermissionOverrides(t *testing.T) {
	t.Parallel()

	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "owner-cat-perm@example.com", Password: "correct-horse-battery", DisplayName: "Owner", BirthDate: "2000-01-15",
	})
	created, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "Category Perm Group")
	if err != nil {
		t.Fatal(err)
	}

	catSvc, _ := newCategoryAndChannelServices(st)

	cat, err := catSvc.CreateGroupCategory(ctx, owner.AccessToken, created.ID, "Text Channels", false)
	if err != nil {
		t.Fatal(err)
	}

	list, err := catSvc.ListCategoryPermissions(ctx, owner.AccessToken, created.ID, cat.ID)
	if err != nil {
		t.Fatalf("ListCategoryPermissions: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected empty overrides initially, got %d", len(list))
	}

	if _, err := catSvc.UpsertCategoryPermission(ctx, owner.AccessToken, created.ID, cat.ID, "bot", "role-1", nil, nil); err == nil {
		t.Fatal("expected invalid input for target_type=bot")
	}
	if _, err := catSvc.UpsertCategoryPermission(ctx, owner.AccessToken, created.ID, cat.ID, "role", "role-1", []string{"view_channels"}, []string{"view_channels"}); err == nil {
		t.Fatal("expected invalid input for overlapping allow & deny")
	}
	if _, err := catSvc.UpsertCategoryPermission(ctx, owner.AccessToken, created.ID, cat.ID, "role", "role-1", []string{"not_a_real_perm"}, nil); err == nil {
		t.Fatal("expected invalid input for unknown permission")
	}

	upserted, err := catSvc.UpsertCategoryPermission(
		ctx, owner.AccessToken, created.ID, cat.ID, "role", "role-1",
		[]string{"view_channels", "view_channels"}, []string{"manage_webhooks"},
	)
	if err != nil {
		t.Fatalf("UpsertCategoryPermission: %v", err)
	}
	if len(upserted.AllowPermissions) != 1 || upserted.AllowPermissions[0] != "view_channels" {
		t.Fatalf("unexpected allow permissions: %+v", upserted)
	}

	if err := catSvc.DeleteCategoryPermission(ctx, owner.AccessToken, created.ID, cat.ID, "role", "role-1"); err != nil {
		t.Fatalf("DeleteCategoryPermission: %v", err)
	}
	if err := catSvc.DeleteCategoryPermission(ctx, owner.AccessToken, created.ID, cat.ID, "bad", "role-1"); err == nil {
		t.Fatal("expected invalid input on DeleteCategoryPermission")
	}
}
