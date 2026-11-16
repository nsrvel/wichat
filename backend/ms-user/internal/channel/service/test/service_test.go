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

	_, _, err := svc.ListGroupChannels(context.Background(), sess.AccessToken, "not-a-group-id")
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

	channels, categories, err := svc.ListGroupChannels(ctx, owner.AccessToken, created.ID)
	if err != nil || len(channels) == 0 {
		t.Fatalf("channels %+v categories %+v %v", channels, categories, err)
	}

	outsider := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "outsider-ch@example.com", Password: "correct-horse-battery", DisplayName: "Out", BirthDate: "2000-01-15",
	})
	_, _, err = svc.ListGroupChannels(ctx, outsider.AccessToken, created.ID)
	if err == nil {
		t.Fatal("expected forbidden")
	}
	ex, ok := exception.As(err)
	if !ok || ex.Kind != exception.KindPermissionDenied {
		t.Fatalf("got %v", err)
	}
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

	// Add member directly to state
	st.Members = append(st.Members, models.GroupMember{
		GroupID: st.Groups[0].ID,
		UserID:  member.User.ID,
		Role:    "member",
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

	// Invalid empty name
	if _, err := svc.CreateGroupCategory(ctx, owner.AccessToken, created.ID, "   ", false); err == nil {
		t.Fatal("expected invalid input for empty name")
	}

	// Regular member without manage_channels is forbidden
	if _, err := svc.CreateGroupCategory(ctx, member.AccessToken, created.ID, "Blocked", false); err == nil {
		t.Fatal("expected permission denied for regular member")
	}

	// Owner creates public & private categories
	cat1, err := svc.CreateGroupCategory(ctx, owner.AccessToken, created.ID, "  Engineering  ", false)
	if err != nil || cat1.Name != "Engineering" || cat1.IsPrivate {
		t.Fatalf("cat1 %+v err %v", cat1, err)
	}
	cat2, err := svc.CreateGroupCategory(ctx, owner.AccessToken, created.ID, "Leadership", true)
	if err != nil || cat2.Name != "Leadership" || !cat2.IsPrivate || cat2.Position <= cat1.Position {
		t.Fatalf("cat2 %+v err %v", cat2, err)
	}

	_, listedCats, err := svc.ListGroupChannels(ctx, owner.AccessToken, created.ID)
	if err != nil || len(listedCats) != 2 {
		t.Fatalf("listedCats %+v err %v", listedCats, err)
	}

	// Member with manage_channels role can create category
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
	cat3, err := svc.CreateGroupCategory(ctx, member.AccessToken, created.ID, "ModOps", false)
	if err != nil || cat3.Name != "ModOps" {
		t.Fatalf("cat3 %+v err %v", cat3, err)
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

	channels, _, err := svc.ListGroupChannels(ctx, owner.AccessToken, created.ID)
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
	channels, _, _ := svc.ListGroupChannels(ctx, owner.AccessToken, created.ID)

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
		Email: "nogroup@example.com", Password: "correct-horse-battery", DisplayName: "Ada", BirthDate: "2000-01-15",
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

func TestReorderChannelCategoriesAndChannels(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "owner-reorder@example.com", Password: "correct-horse-battery", DisplayName: "Owner", BirthDate: "2000-01-15",
	})
	created, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "Reorder Group")
	if err != nil {
		t.Fatal(err)
	}
	outsider := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "outsider-reorder@example.com", Password: "correct-horse-battery", DisplayName: "Out", BirthDate: "2000-01-15",
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

	catA, err := svc.CreateGroupCategory(ctx, owner.AccessToken, created.ID, "Alpha", false)
	if err != nil {
		t.Fatal(err)
	}
	catB, err := svc.CreateGroupCategory(ctx, owner.AccessToken, created.ID, "Beta", false)
	if err != nil {
		t.Fatal(err)
	}

	if err := svc.ReorderChannelCategories(ctx, owner.AccessToken, created.ID, []string{catB.ID, catA.ID}); err != nil {
		t.Fatal(err)
	}
	channels, cats, err := svc.ListGroupChannels(ctx, owner.AccessToken, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(cats) != 2 || cats[0].ID != catB.ID || cats[1].ID != catA.ID {
		t.Fatalf("unexpected category order: %+v", cats)
	}

	// Permission denied for non-member.
	if err := svc.ReorderChannelCategories(ctx, outsider.AccessToken, created.ID, []string{catA.ID, catB.ID}); err == nil {
		t.Fatal("expected forbidden")
	}

	// Reorder channels and move into another category and into favorites.
	if len(channels) == 1 {
		if err := svc.ReorderGroupChannels(ctx, owner.AccessToken, created.ID, "", catB.ID, []string{channels[0].ID}); err != nil {
			t.Fatal(err)
		}
		updatedChannels, _, err := svc.ListGroupChannels(ctx, owner.AccessToken, created.ID)
		if err != nil {
			t.Fatal(err)
		}
		if updatedChannels[0].CategoryID != catB.ID {
			t.Fatalf("expected channel moved to %s, got %s", catB.ID, updatedChannels[0].CategoryID)
		}

		if err := svc.ReorderGroupChannels(ctx, owner.AccessToken, created.ID, "favorites", "", []string{channels[0].ID}); err != nil {
			t.Fatal(err)
		}
		favChannels, _, err := svc.ListGroupChannels(ctx, owner.AccessToken, created.ID)
		if err != nil {
			t.Fatal(err)
		}
		if favChannels[0].SidebarSection != "favorites" || favChannels[0].CategoryID != "" {
			t.Fatalf("expected channel in favorites, got section=%s cat=%s", favChannels[0].SidebarSection, favChannels[0].CategoryID)
		}
	}

	// Rejects unknown category_id.
	if err := svc.ReorderGroupChannels(ctx, owner.AccessToken, created.ID, "", "eeeeeeeeeeeeeeeeeeeeeeee", []string{channels[0].ID}); err == nil {
		t.Fatal("expected invalid input for unknown category_id")
	}

	// Rejects empty channel_ids.
	if err := svc.ReorderGroupChannels(ctx, owner.AccessToken, created.ID, "", "", nil); err == nil {
		t.Fatal("expected invalid input for empty channel_ids")
	}

	// Rejects duplicate category_ids.
	if err := svc.ReorderChannelCategories(ctx, owner.AccessToken, created.ID, []string{catA.ID, catA.ID}); err == nil {
		t.Fatal("expected invalid input for duplicate category_ids")
	}
}

func TestCreateGroupChannelLifecycleAndPermissions(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "owner-ch@example.com", Password: "correct-horse-battery", DisplayName: "Owner", BirthDate: "2000-01-15",
	})
	created, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "Channels Group")
	if err != nil {
		t.Fatal(err)
	}
	outsider := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "outsider-ch@example.com", Password: "correct-horse-battery", DisplayName: "Out", BirthDate: "2000-01-15",
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

	cat, err := svc.CreateGroupCategory(ctx, owner.AccessToken, created.ID, "Engineering", false)
	if err != nil {
		t.Fatal(err)
	}

	// Validation errors.
	if _, err := svc.CreateGroupChannel(ctx, owner.AccessToken, created.ID, "   ", "text_only", cat.ID, false); err == nil {
		t.Fatal("expected invalid input for blank name")
	}
	if _, err := svc.CreateGroupChannel(ctx, owner.AccessToken, created.ID, "bad-type", "text_and_voice", cat.ID, false); err == nil {
		t.Fatal("expected invalid input for unsupported type")
	}
	if _, err := svc.CreateGroupChannel(ctx, owner.AccessToken, created.ID, "bad-cat", "text_only", "ffffffffffffffffffffffff", false); err == nil {
		t.Fatal("expected invalid input for unknown category")
	}

	// Permission denied for non-member.
	if _, err := svc.CreateGroupChannel(ctx, outsider.AccessToken, created.ID, "blocked", "text_only", cat.ID, false); err == nil {
		t.Fatal("expected forbidden")
	}

	// Create all 4 creatable channel types.
	for _, kind := range []string{"text_only", "voice_only", "forum", "broadcast"} {
		ch, err := svc.CreateGroupChannel(ctx, owner.AccessToken, created.ID, "ch-"+kind, kind, cat.ID, kind == "broadcast")
		if err != nil {
			t.Fatalf("CreateGroupChannel(%s): %v", kind, err)
		}
		if ch.CategoryID != cat.ID || ch.Type != kind {
			t.Fatalf("unexpected created channel: %+v", ch)
		}
	}
}
