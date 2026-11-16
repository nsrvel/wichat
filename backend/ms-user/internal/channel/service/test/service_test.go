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
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
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
		categorymock.NewPgRepositoryMock(st),
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
		categorymock.NewPgRepositoryMock(st),
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
		categorymock.NewPgRepositoryMock(st),
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
		categorymock.NewPgRepositoryMock(st),
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
		categorymock.NewPgRepositoryMock(st),
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
	catRepo := categorymock.NewPgRepositoryMock(st)
	log := logger.New(logger.Options{Service: "channel-test", Env: "test", Level: "error"})
	svc := channelservice.NewService(
		testSupport.ChannelRepo(st),
		catRepo,
		groups,
		groupmembermock.NewPgRepositoryMock(st),
		grouprolemock.NewPgRepositoryMock(st),
		sessions,
		[]byte(testSupport.TestJWTSecret),
		log,
	)
	catSvc := categoryservice.NewService(
		catRepo,
		groups,
		groupmembermock.NewPgRepositoryMock(st),
		grouprolemock.NewPgRepositoryMock(st),
		sessions,
		[]byte(testSupport.TestJWTSecret),
		log,
	)

	catA, err := catSvc.CreateGroupCategory(ctx, owner.AccessToken, created.ID, "Alpha", false)
	if err != nil {
		t.Fatal(err)
	}
	catB, err := catSvc.CreateGroupCategory(ctx, owner.AccessToken, created.ID, "Beta", false)
	if err != nil {
		t.Fatal(err)
	}

	if err := catSvc.ReorderChannelCategories(ctx, owner.AccessToken, created.ID, []string{catB.ID, catA.ID}); err != nil {
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
	if err := catSvc.ReorderChannelCategories(ctx, outsider.AccessToken, created.ID, []string{catA.ID, catB.ID}); err == nil {
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
	if err := catSvc.ReorderChannelCategories(ctx, owner.AccessToken, created.ID, []string{catA.ID, catA.ID}); err == nil {
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
	catRepo := categorymock.NewPgRepositoryMock(st)
	log := logger.New(logger.Options{Service: "channel-test", Env: "test", Level: "error"})
	svc := channelservice.NewService(
		testSupport.ChannelRepo(st),
		catRepo,
		groups,
		groupmembermock.NewPgRepositoryMock(st),
		grouprolemock.NewPgRepositoryMock(st),
		sessions,
		[]byte(testSupport.TestJWTSecret),
		log,
	)
	catSvc := categoryservice.NewService(
		catRepo,
		groups,
		groupmembermock.NewPgRepositoryMock(st),
		grouprolemock.NewPgRepositoryMock(st),
		sessions,
		[]byte(testSupport.TestJWTSecret),
		log,
	)

	cat, err := catSvc.CreateGroupCategory(ctx, owner.AccessToken, created.ID, "Engineering", false)
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

func TestCategoryPermissionEnforcementAndPrecedence(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "owner-enforce@example.com", Password: "correct-horse-battery", DisplayName: "Owner", BirthDate: "2000-01-15",
	})
	member := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "member-enforce@example.com", Password: "correct-horse-battery", DisplayName: "Member", BirthDate: "2000-01-15",
	})

	created, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "Enforce Group")
	if err != nil {
		t.Fatal(err)
	}
	st.Lock()
	groupUUID := st.Groups[0].ID
	memberUserID := member.User.ID
	st.Members = append(st.Members, models.GroupMember{
		GroupID: groupUUID,
		UserID:  memberUserID,
		Role:    constants.GroupRoleMember,
	})
	st.Unlock()

	_, groups, sessions := testSupport.TestRepos(st)
	roleRepo := grouprolemock.NewPgRepositoryMock(st)
	catRepo := categorymock.NewPgRepositoryMock(st)
	log := logger.New(logger.Options{Service: "channel-test", Env: "test", Level: "error"})
	svc := channelservice.NewService(
		testSupport.ChannelRepo(st),
		catRepo,
		groups,
		groupmembermock.NewPgRepositoryMock(st),
		roleRepo,
		sessions,
		[]byte(testSupport.TestJWTSecret),
		log,
	)
	catSvc := categoryservice.NewService(
		catRepo,
		groups,
		groupmembermock.NewPgRepositoryMock(st),
		roleRepo,
		sessions,
		[]byte(testSupport.TestJWTSecret),
		log,
	)

	// 1. Owner creates a private category & channel inside it.
	secretCat, err := catSvc.CreateGroupCategory(ctx, owner.AccessToken, created.ID, "Secret", true)
	if err != nil {
		t.Fatal(err)
	}
	secretCh, err := svc.CreateGroupChannel(ctx, owner.AccessToken, created.ID, "secret-room", constants.ChannelKindTextOnly, secretCat.ID, false)
	if err != nil {
		t.Fatal(err)
	}

	// Owner sees private category and channel.
	ownerChs, ownerCats, err := svc.ListGroupChannels(ctx, owner.AccessToken, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(ownerCats) == 0 || len(ownerChs) == 0 {
		t.Fatalf("expected owner to see categories and channels")
	}

	// 2. Regular member does NOT see private category or its channel, and AssertGroupChannelAccess rejects.
	memChs, memCats, err := svc.ListGroupChannels(ctx, member.AccessToken, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range memCats {
		if c.ID == secretCat.ID {
			t.Fatalf("expected private category %s to be hidden from regular member", secretCat.ID)
		}
	}
	for _, ch := range memChs {
		if ch.ID == secretCh.ID {
			t.Fatalf("expected channel %s in private category to be hidden from regular member", secretCh.ID)
		}
	}
	_, err = svc.AssertGroupChannelAccess(ctx, member.AccessToken, created.ID, secretCh.ID)
	if ex, ok := exception.As(err); !ok || ex.Kind != exception.KindPermissionDenied {
		t.Fatalf("expected forbidden for channel in private category, got %v", err)
	}

	// 3. Grant view_channels to a custom role assigned to member -> member sees category & channel.
	vipRole, err := roleRepo.CreateRole(ctx, groupUUID, "VIP", "#3b82f6", false, false, []string{constants.PermViewChannels}, st.Users[0].CreatedAt)
	if err != nil {
		t.Fatal(err)
	}
	if err := roleRepo.AddMemberRoles(ctx, groupUUID, vipRole.ID, []string{memberUserID}); err != nil {
		t.Fatal(err)
	}
	if _, err := catSvc.UpsertCategoryPermission(ctx, owner.AccessToken, created.ID, secretCat.ID, "role", vipRole.PublicID, []string{constants.PermViewChannels}, nil); err != nil {
		t.Fatal(err)
	}

	memChs, memCats, err = svc.ListGroupChannels(ctx, member.AccessToken, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	foundCat := false
	for _, c := range memCats {
		if c.ID == secretCat.ID {
			foundCat = true
			break
		}
	}
	if !foundCat {
		t.Fatalf("expected private category %s to be visible with role allow override", secretCat.ID)
	}
	foundCh := false
	for _, ch := range memChs {
		if ch.ID == secretCh.ID {
			foundCh = true
			break
		}
	}
	if !foundCh {
		t.Fatalf("expected channel %s to be visible with role allow override", secretCh.ID)
	}
	if _, err := svc.AssertGroupChannelAccess(ctx, member.AccessToken, created.ID, secretCh.ID); err != nil {
		t.Fatalf("expected channel access allowed via role override, got %v", err)
	}

	// 4. Member-level deny overrides role-level allow.
	if _, err := catSvc.UpsertCategoryPermission(ctx, owner.AccessToken, created.ID, secretCat.ID, "member", memberUserID, nil, []string{constants.PermViewChannels}); err != nil {
		t.Fatal(err)
	}
	_, err = svc.AssertGroupChannelAccess(ctx, member.AccessToken, created.ID, secretCh.ID)
	if ex, ok := exception.As(err); !ok || ex.Kind != exception.KindPermissionDenied {
		t.Fatalf("expected member deny to override role allow, got %v", err)
	}

	// 5. @everyone override on public category & member allow override precedence + Administrator bypass.
	pubCat, err := catSvc.CreateGroupCategory(ctx, owner.AccessToken, created.ID, "PublicLocked", false)
	if err != nil {
		t.Fatal(err)
	}
	pubCh, err := svc.CreateGroupChannel(ctx, owner.AccessToken, created.ID, "locked-room", constants.ChannelKindTextOnly, pubCat.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	allRoles, err := roleRepo.ListRolesForGroup(ctx, groupUUID)
	if err != nil {
		t.Fatal(err)
	}
	var everyonePubID string
	for _, r := range allRoles {
		if r.IsDefault {
			everyonePubID = r.PublicID
			break
		}
	}
	if _, err := catSvc.UpsertCategoryPermission(ctx, owner.AccessToken, created.ID, pubCat.ID, "role", everyonePubID, nil, []string{constants.PermViewChannels}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AssertGroupChannelAccess(ctx, member.AccessToken, created.ID, pubCh.ID); err == nil {
		t.Fatal("expected @everyone deny to block member access")
	}

	// Member allow overrides @everyone deny.
	if _, err := catSvc.UpsertCategoryPermission(ctx, owner.AccessToken, created.ID, pubCat.ID, "member", memberUserID, []string{constants.PermViewChannels}, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AssertGroupChannelAccess(ctx, member.AccessToken, created.ID, pubCh.ID); err != nil {
		t.Fatalf("expected member allow to override @everyone deny, got %v", err)
	}

	// Administrator bypasses all category restrictions.
	admin := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "admin-enforce@example.com", Password: "correct-horse-battery", DisplayName: "Admin", BirthDate: "2000-01-15",
	})
	st.Lock()
	st.Members = append(st.Members, models.GroupMember{
		GroupID: groupUUID,
		UserID:  admin.User.ID,
		Role:    constants.GroupRoleMember,
	})
	st.Unlock()
	adminRole, err := roleRepo.CreateRole(ctx, groupUUID, "AdminRole", "#ef4444", false, false, []string{constants.PermAdministrator}, st.Users[0].CreatedAt)
	if err != nil {
		t.Fatal(err)
	}
	if err := roleRepo.AddMemberRoles(ctx, groupUUID, adminRole.ID, []string{admin.User.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AssertGroupChannelAccess(ctx, admin.AccessToken, created.ID, secretCh.ID); err != nil {
		t.Fatalf("expected administrator to bypass private category, got %v", err)
	}
	adminChs, adminCats, err := svc.ListGroupChannels(ctx, admin.AccessToken, created.ID)
	if err != nil || len(adminCats) == 0 || len(adminChs) == 0 {
		t.Fatalf("expected administrator to list all categories and channels, got err=%v", err)
	}

	// Verify WithGroupActivityLog records category mutations cleanly.
	rec := &recordingActivityLog{}
	auditedCatSvc := catSvc.WithGroupActivityLog(rec)
	auditedCat, err := auditedCatSvc.CreateGroupCategory(ctx, owner.AccessToken, created.ID, "Audited", false)
	if err != nil {
		t.Fatal(err)
	}
	newName := "AuditedRenamed"
	if _, err := auditedCatSvc.UpdateGroupCategory(ctx, owner.AccessToken, created.ID, auditedCat.ID, &newName, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := auditedCatSvc.UpsertCategoryPermission(ctx, owner.AccessToken, created.ID, auditedCat.ID, "member", memberUserID, []string{constants.PermViewChannels}, nil); err != nil {
		t.Fatal(err)
	}
	if err := auditedCatSvc.DeleteCategoryPermission(ctx, owner.AccessToken, created.ID, auditedCat.ID, "member", memberUserID); err != nil {
		t.Fatal(err)
	}
	if err := auditedCatSvc.DeleteGroupCategory(ctx, owner.AccessToken, created.ID, auditedCat.ID); err != nil {
		t.Fatal(err)
	}
	if len(rec.entries) != 5 {
		t.Fatalf("expected 5 audit entries, got %d", len(rec.entries))
	}
}

func TestChannelSettingsAndPermissions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	owner := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email:       "chsettings@example.com",
		Password:    "correct-horse-battery",
		DisplayName: "ChanSettingsOwner",
		BirthDate:   "2000-01-15",
	})
	_, groups, sessions := testSupport.TestRepos(st)
	members := groupmembermock.NewPgRepositoryMock(st)
	roleRepo := grouprolemock.NewPgRepositoryMock(st)
	chanRepo := testSupport.ChannelRepo(st)
	catRepo := categorymock.NewPgRepositoryMock(st)
	log := logger.New(logger.Options{Service: "channel-test", Env: "test", Level: "error"})
	svc := channelservice.NewService(chanRepo, catRepo, groups, members, roleRepo, sessions, []byte(testSupport.TestJWTSecret), log)
	catSvc := categoryservice.NewService(catRepo, groups, members, roleRepo, sessions, []byte(testSupport.TestJWTSecret), log)

	created, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "ChanSettingsGroup")
	if err != nil {
		t.Fatal(err)
	}
	st.Lock()
	groupUUID := st.Groups[0].ID
	st.Unlock()

	// Create a channel
	ch, err := svc.CreateGroupChannel(ctx, owner.AccessToken, created.ID, "lounge", constants.ChannelKindVoiceOnly, "", false)
	if err != nil {
		t.Fatal(err)
	}

	// Empty update -> invalid input
	if _, err := svc.UpdateGroupChannel(ctx, owner.AccessToken, created.ID, ch.ID, models.UpdateChannelInput{}); err == nil {
		t.Fatal("expected error for empty update input")
	}

	// Invalid bitrate -> error
	badBitrate := int32(256)
	if _, err := svc.UpdateGroupChannel(ctx, owner.AccessToken, created.ID, ch.ID, models.UpdateChannelInput{BitrateKbps: &badBitrate}); err == nil {
		t.Fatal("expected error for invalid bitrate")
	}

	// Invalid content visibility -> error
	badVis := "unknown"
	if _, err := svc.UpdateGroupChannel(ctx, owner.AccessToken, created.ID, ch.ID, models.UpdateChannelInput{ContentVisibility: &badVis}); err == nil {
		t.Fatal("expected error for invalid content_visibility")
	}

	// Valid update across settings fields
	newName := "lounge-hq"
	isPriv := true
	slow := int32(15)
	threadSlow := int32(30)
	vis := "spoiler"
	layout := "gallery"
	sortOrder := "creation_time"
	tagMatch := "match_all"
	bitrate := int32(96)
	videoQual := "720p"
	userLimit := int32(12)
	region := "singapore"

	updated, err := svc.UpdateGroupChannel(ctx, owner.AccessToken, created.ID, ch.ID, models.UpdateChannelInput{
		Name:                  &newName,
		IsPrivate:             &isPriv,
		SlowmodeSeconds:       &slow,
		ThreadSlowmodeSeconds: &threadSlow,
		ContentVisibility:     &vis,
		ForumDefaultLayout:    &layout,
		ForumSortOrder:        &sortOrder,
		ForumTagMatching:      &tagMatch,
		BitrateKbps:           &bitrate,
		VideoQuality:          &videoQual,
		UserLimit:             &userLimit,
		RTCRegion:             &region,
	})
	if err != nil {
		t.Fatalf("UpdateGroupChannel failed: %v", err)
	}
	if updated.Name != newName || !updated.IsPrivate || updated.SlowmodeSeconds != 15 || updated.BitrateKbps != 96 || updated.RTCRegion != "singapore" {
		t.Fatalf("unexpected updated channel: %+v", updated)
	}

	// Channel permissions upsert / list / delete
	modRole, err := roleRepo.CreateRole(ctx, groupUUID, "Moderator", "#22c55e", false, false, []string{constants.PermViewChannels}, st.Users[0].CreatedAt)
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}
	everyoneID := modRole.PublicID
	if _, err := svc.UpsertChannelPermission(ctx, owner.AccessToken, created.ID, ch.ID, "role", everyoneID, []string{constants.PermViewChannels}, nil); err != nil {
		t.Fatalf("UpsertChannelPermission failed: %v", err)
	}
	// Conflict between allow & deny -> error
	if _, err := svc.UpsertChannelPermission(ctx, owner.AccessToken, created.ID, ch.ID, "role", everyoneID, []string{constants.PermViewChannels}, []string{constants.PermViewChannels}); err == nil {
		t.Fatal("expected error when allow and deny overlap")
	}
	overrides, err := svc.ListChannelPermissions(ctx, owner.AccessToken, created.ID, ch.ID)
	if err != nil || len(overrides) != 1 {
		t.Fatalf("ListChannelPermissions unexpected: len=%d err=%v", len(overrides), err)
	}
	if err := svc.DeleteChannelPermission(ctx, owner.AccessToken, created.ID, ch.ID, "role", everyoneID); err != nil {
		t.Fatalf("DeleteChannelPermission failed: %v", err)
	}
	overrides, err = svc.ListChannelPermissions(ctx, owner.AccessToken, created.ID, ch.ID)
	if err != nil || len(overrides) != 0 {
		t.Fatalf("expected 0 overrides after delete, got len=%d err=%v", len(overrides), err)
	}

	// Delete channel
	if err := svc.DeleteGroupChannel(ctx, owner.AccessToken, created.ID, ch.ID); err != nil {
		t.Fatalf("DeleteGroupChannel failed: %v", err)
	}
	if err := svc.DeleteGroupChannel(ctx, owner.AccessToken, created.ID, ch.ID); err == nil {
		t.Fatal("expected error deleting already-deleted channel")
	}

	// SyncChannelPermissions copies category privacy & overrides to channel
	cat, err := catSvc.CreateGroupCategory(ctx, owner.AccessToken, created.ID, "SyncedCat", true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := catSvc.UpsertCategoryPermission(ctx, owner.AccessToken, created.ID, cat.ID, "role", everyoneID, []string{constants.PermViewChannels}, nil); err != nil {
		t.Fatal(err)
	}
	catCh, err := svc.CreateGroupChannel(ctx, owner.AccessToken, created.ID, "synced-room", constants.ChannelKindTextOnly, cat.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UpsertChannelPermission(ctx, owner.AccessToken, created.ID, catCh.ID, "role", everyoneID, []string{constants.PermViewChannels}, nil); err != nil {
		t.Fatal(err)
	}
	synced, err := svc.SyncChannelPermissions(ctx, owner.AccessToken, created.ID, catCh.ID)
	if err != nil {
		t.Fatalf("SyncChannelPermissions failed: %v", err)
	}
	if !synced.IsPrivate {
		t.Fatal("expected synced channel to inherit category IsPrivate=true")
	}
	syncedOverrides, err := svc.ListChannelPermissions(ctx, owner.AccessToken, created.ID, catCh.ID)
	if err != nil || len(syncedOverrides) != 0 {
		t.Fatalf("expected 0 overrides after sync (inherit all), got len=%d err=%v", len(syncedOverrides), err)
	}
}

type recordingActivityLog struct {
	entries []models.GroupActivityLogEntry
}

func (r *recordingActivityLog) Record(_ context.Context, entry models.GroupActivityLogEntry) {
	r.entries = append(r.entries, entry)
}
