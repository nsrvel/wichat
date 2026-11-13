package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	channelservice "github.com/wichat/wichat/backend/ms-user/internal/channel/service"
	groupmembermock "github.com/wichat/wichat/backend/ms-user/internal/group-member/repository/mock"
	grouprolemock "github.com/wichat/wichat/backend/ms-user/internal/group-role/repository/mock"
	invitemock "github.com/wichat/wichat/backend/ms-user/internal/invite/repository/mock"
	inviteservice "github.com/wichat/wichat/backend/ms-user/internal/invite/service"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	testSupport "github.com/wichat/wichat/backend/ms-user/pkg/test-support"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	notificationv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/notification/v1"
)

func TestPreviewInviteNotFound(t *testing.T) {
	t.Parallel()

	st := usermock.NewState()
	svc := newInviteService(t, st)

	out, err := svc.PreviewInvite(context.Background(), "missing")
	if err != nil {
		t.Fatal(err)
	}
	if out.Valid || out.Reason != constants.InvitePreviewNotFound {
		t.Fatalf("%+v", out)
	}
}

func TestPreviewInviteValid(t *testing.T) {
	t.Parallel()

	st := usermock.NewState()
	st.Groups = []models.Group{{ID: "g1", Name: "Team", LogoURL: "https://cdn.example/icon.png"}}
	st.Members = []models.GroupMember{{GroupID: "g1", UserID: "u1", Role: constants.GroupRoleOwner}}
	st.Invites = []models.GroupInviteRow{
		{
			ID:        "inv1",
			Token:     "12345678",
			GroupID:   "g1",
			Kind:      constants.GroupInviteKindShareLink,
			ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
			CreatedAt: time.Now().UTC(),
		},
	}

	svc := newInviteService(t, st)
	out, err := svc.PreviewInvite(context.Background(), "12345678")
	if err != nil {
		t.Fatal(err)
	}
	if !out.Valid || out.GroupName != "Team" || out.GroupLogoURL != "https://cdn.example/icon.png" {
		t.Fatalf("%+v", out)
	}
	if out.MemberCount != 1 || out.OnlineCount != 0 {
		t.Fatalf("counts: %+v", out)
	}
}

func TestGetInviteLinkAndRedeemShareLink(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("owner-link@example.com"))
	created, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "Invite Team")
	if err != nil {
		t.Fatal(err)
	}

	svc := newInviteService(t, st)
	link, err := svc.GetInviteLink(ctx, owner.AccessToken, created.ID, "")
	if err != nil || link.Token == "" {
		t.Fatalf("link: %+v %v", link, err)
	}

	member := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("member-link@example.com"))
	out, err := svc.RedeemInvite(ctx, member.AccessToken, link.Token)
	if err != nil || out.GroupPublicID != created.ID {
		t.Fatalf("redeem: %+v %v", out, err)
	}
}

func TestSendGroupInvitesAddsUser(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("owner-invite@example.com"))
	created, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "Batch")
	if err != nil {
		t.Fatal(err)
	}
	target := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("target@example.com"))

	svc := newInviteService(t, st)
	res, err := svc.SendGroupInvites(ctx, owner.AccessToken, models.SendGroupInvitesInput{
		GroupPublicID: created.ID,
		UserIDs:       []string{target.User.ID},
	})
	if err != nil || res.Added != 1 {
		t.Fatalf("send: %+v %v", res, err)
	}
}

func TestPreviewInviteExpired(t *testing.T) {
	st := usermock.NewState()
	st.Groups = []models.Group{{ID: "g1", Name: "Old"}}
	st.Invites = []models.GroupInviteRow{{
		ID: "i1", Token: "expired1", GroupID: "g1", Kind: constants.GroupInviteKindShareLink,
		ExpiresAt: time.Now().UTC().Add(-time.Hour),
	}}
	svc := newInviteService(t, st)
	out, err := svc.PreviewInvite(context.Background(), "expired1")
	if err != nil || out.Valid || out.Reason != constants.InvitePreviewExpired {
		t.Fatalf("%+v %v", out, err)
	}
}

func TestSendGroupInvitesEmailsUnknownUser(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("owner-mail@example.com"))
	created, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "Mailers")
	if err != nil {
		t.Fatal(err)
	}

	notifier := &testSupport.FakeNotifier{}
	users, groups, sessions := testSupport.TestRepos(st)
	svc := inviteservice.NewService(
		groups,
		groupmembermock.NewPgRepositoryMock(st),
		grouprolemock.NewPgRepositoryMock(st),
		testSupport.ChannelRepo(st),
		invitemock.NewPgRepositoryMock(st),
		users,
		sessions,
		notifier,
		[]byte(testSupport.TestJWTSecret),
		"https://app.test",
		"",
		logger.New(logger.Options{Service: "invite-test", Env: "test", Level: "error"}),
	)

	res, err := svc.SendGroupInvites(ctx, owner.AccessToken, models.SendGroupInvitesInput{
		GroupPublicID: created.ID,
		Emails:        []string{"newfriend@example.com"},
	})
	if err != nil || res.Emailed != 1 || notifier.Sends != 1 {
		t.Fatalf("send: %+v sends=%d %v", res, notifier.Sends, err)
	}
}

func TestPreviewInviteEmptyToken(t *testing.T) {
	st := usermock.NewState()
	svc := newInviteService(t, st)
	_, err := svc.PreviewInvite(context.Background(), "  ")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRedeemEmailInviteWrongAccount(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("owner-email-inv@example.com"))
	created, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "EmailOnly")
	if err != nil {
		t.Fatal(err)
	}

	_, groups, _ := testSupport.TestRepos(st)
	group, err := groups.FindGroupByPublicID(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	email := "invited@example.com"
	tokenStr := "emailtok"
	st.Invites = []models.GroupInviteRow{{
		ID: "inv-email", Token: tokenStr, GroupID: group.ID, Kind: constants.GroupInviteKindEmail,
		Email: &email, ExpiresAt: time.Now().UTC().Add(24 * time.Hour), CreatedAt: time.Now().UTC(),
	}}

	wrong := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("wrong@example.com"))
	svc := newInviteService(t, st)
	_, err = svc.RedeemInvite(ctx, wrong.AccessToken, tokenStr)
	if err == nil {
		t.Fatal("expected forbidden")
	}
	ex, ok := exception.As(err)
	if !ok || ex.Kind != exception.KindPermissionDenied {
		t.Fatalf("got %v", err)
	}
}

func TestPreviewInviteRedeemed(t *testing.T) {
	st := usermock.NewState()
	redeemed := time.Now().UTC()
	st.Groups = []models.Group{{ID: "g1", Name: "G"}}
	st.Invites = []models.GroupInviteRow{{
		ID: "i1", Token: "used1", GroupID: "g1", Kind: constants.GroupInviteKindShareLink,
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour), RedeemedAt: &redeemed,
	}}

	svc := newInviteService(t, st)
	out, err := svc.PreviewInvite(context.Background(), "used1")
	if err != nil || out.Valid || out.Reason != constants.InvitePreviewRedeemed {
		t.Fatalf("%+v %v", out, err)
	}
}

func TestSendGroupInvitesRejectsEmptyRecipients(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("owner-empty@example.com"))
	created, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "Empty")
	if err != nil {
		t.Fatal(err)
	}

	svc := newInviteService(t, st)
	_, err = svc.SendGroupInvites(ctx, owner.AccessToken, models.SendGroupInvitesInput{GroupPublicID: created.ID})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRedeemEmailInviteMatchingUser(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("owner-redeem@example.com"))
	created, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "Redeem")
	if err != nil {
		t.Fatal(err)
	}

	invitee := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("invited-redeem@example.com"))
	_, groups, _ := testSupport.TestRepos(st)
	group, err := groups.FindGroupByPublicID(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}

	email := invitee.User.Email
	tokenStr := "emailtok2"
	st.Invites = []models.GroupInviteRow{{
		ID: "inv2", Token: tokenStr, GroupID: group.ID, Kind: constants.GroupInviteKindEmail,
		Email: &email, ExpiresAt: time.Now().UTC().Add(24 * time.Hour), CreatedAt: time.Now().UTC(),
	}}

	svc := newInviteService(t, st)
	out, err := svc.RedeemInvite(ctx, invitee.AccessToken, tokenStr)
	if err != nil || out.GroupPublicID != created.ID {
		t.Fatalf("redeem %+v %v", out, err)
	}
}

func TestSendGroupInvitesSkipsInvalidEmailAndDupes(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("owner-skip@example.com"))
	created, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "Skips")
	if err != nil {
		t.Fatal(err)
	}

	svc := newInviteService(t, st)
	res, err := svc.SendGroupInvites(ctx, owner.AccessToken, models.SendGroupInvitesInput{
		GroupPublicID: created.ID,
		Emails:        []string{"not-an-email", "a@example.com", "a@example.com"},
	})
	if err != nil || res.Skipped < 2 {
		t.Fatalf("res %+v %v", res, err)
	}
}

func TestSendGroupInvitesRejectsOversizeBatch(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("owner-batch@example.com"))
	created, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "Batch")
	if err != nil {
		t.Fatal(err)
	}

	emails := make([]string, 0, constants.MaxInviteBatchSize+1)
	for i := 0; i < constants.MaxInviteBatchSize+1; i++ {
		emails = append(emails, "e"+strings.Repeat("x", 4)+"@example.com")
	}
	svc := newInviteService(t, st)
	_, err = svc.SendGroupInvites(ctx, owner.AccessToken, models.SendGroupInvitesInput{
		GroupPublicID: created.ID,
		Emails:        emails,
	})
	if err == nil {
		t.Fatal("expected batch limit error")
	}
}

func TestGetInviteLinkWithChannel(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("owner-ch@example.com"))
	created, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "ChLink")
	if err != nil {
		t.Fatal(err)
	}

	_, groups, sessions := testSupport.TestRepos(st)
	chSvc := channelservice.NewService(testSupport.ChannelRepo(st), groups, groupmembermock.NewPgRepositoryMock(st), grouprolemock.NewPgRepositoryMock(st), sessions, []byte(testSupport.TestJWTSecret),
		logger.New(logger.Options{Service: "t", Env: "test", Level: "error"}))
	channels, err := chSvc.ListGroupChannels(ctx, owner.AccessToken, created.ID)
	if err != nil || len(channels) == 0 {
		t.Fatal(err)
	}

	svc := newInviteService(t, st)
	_, err = svc.GetInviteLink(ctx, owner.AccessToken, created.ID, channels[0].ID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetInviteLinkReusesOpenShareLink(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("owner-reuse@example.com"))
	created, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "Reuse")
	if err != nil {
		t.Fatal(err)
	}

	svc := newInviteService(t, st)
	first, err := svc.GetInviteLink(ctx, owner.AccessToken, created.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.GetInviteLink(ctx, owner.AccessToken, created.ID, "")
	if err != nil || second.Token != first.Token {
		t.Fatalf("tokens %q %q", first.Token, second.Token)
	}
}

func TestValidEmailForChip(t *testing.T) {
	t.Parallel()

	if !inviteservice.ValidEmailForChip("ada@example.com") {
		t.Fatal("valid chip")
	}
	if inviteservice.ValidEmailForChip("not-email") {
		t.Fatal("invalid chip")
	}
}

func testSupportRegisterInput(email string) models.RegisterInput {
	return models.RegisterInput{
		Email:       email,
		Password:    "correct-horse-battery",
		DisplayName: "Ada",
		BirthDate:   "2000-01-15",
	}
}

func newInviteService(t *testing.T, st *usermock.State) inviteservice.Service {
	t.Helper()
	users, groups, sessions := testSupport.TestRepos(st)
	return inviteservice.NewService(
		groups,
		groupmembermock.NewPgRepositoryMock(st),
		grouprolemock.NewPgRepositoryMock(st),
		testSupport.ChannelRepo(st),
		invitemock.NewPgRepositoryMock(st),
		users,
		sessions,
		&testSupport.FakeNotifier{},
		[]byte(testSupport.TestJWTSecret),
		"https://app.test",
		"",
		logger.New(logger.Options{Service: "invite-test", Env: "test", Level: "error"}),
	)
}

func newInviteServiceNoWebBase(t *testing.T, st *usermock.State) inviteservice.Service {
	t.Helper()
	users, groups, sessions := testSupport.TestRepos(st)
	return inviteservice.NewService(groups, groupmembermock.NewPgRepositoryMock(st), grouprolemock.NewPgRepositoryMock(st), testSupport.ChannelRepo(st), invitemock.NewPgRepositoryMock(st), users, sessions,
		&testSupport.FakeNotifier{}, []byte(testSupport.TestJWTSecret), "",
		"",
		logger.New(logger.Options{Service: "invite", Env: "test", Level: "error"}))
}

func codeOf(t *testing.T, err error) string {
	t.Helper()
	ex, ok := exception.As(err)
	if !ok {
		t.Fatalf("not an exception: %v", err)
	}
	return string(ex.Code)
}

func TestInviteURLWithoutWebBase(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("webbase@example.com"))
	created, _ := groupSvc.CreateGroup(ctx, owner.AccessToken, "URL")

	svc := newInviteServiceNoWebBase(t, st)
	link, err := svc.GetInviteLink(ctx, owner.AccessToken, created.ID, "")
	if err != nil || link.Token == "" {
		t.Fatalf("%+v %v", link, err)
	}
}

func TestRedeemInviteAlreadyMember(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("redeem-member@example.com"))
	created, _ := groupSvc.CreateGroup(ctx, owner.AccessToken, "Again")
	member := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("already-in@example.com"))
	_ = testSupport.NewGroupMemberService(t, st).JoinGroup(ctx, member.AccessToken, created.ID)

	svc := newInviteService(t, st)
	link, err := svc.GetInviteLink(ctx, owner.AccessToken, created.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	out, err := svc.RedeemInvite(ctx, member.AccessToken, link.Token)
	if err != nil || out.GroupPublicID != created.ID {
		t.Fatalf("%+v %v", out, err)
	}
}

func TestSendGroupInvitesInvalidChannel(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("bad-ch@example.com"))
	created, _ := groupSvc.CreateGroup(ctx, owner.AccessToken, "Ch")

	svc := newInviteService(t, st)
	_, err := svc.SendGroupInvites(ctx, owner.AccessToken, models.SendGroupInvitesInput{
		GroupPublicID:   created.ID,
		Emails:          []string{"new@example.com"},
		ChannelPublicID: "eeeeeeeeeeeeeeeeeeeeeeee",
	})
	if err == nil {
		t.Fatal("expected invalid channel")
	}
}

func TestRedeemInviteNotFound(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, _, _ := testSupport.NewAuthService(t, st, nil)
	member := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("redeem-miss@example.com"))

	svc := newInviteService(t, st)
	_, err := svc.RedeemInvite(context.Background(), member.AccessToken, "badtoken")
	if codeOf(t, err) != string(exception.MsgInvalidInput) {
		t.Fatalf("%v", err)
	}
}

func TestSendGroupInvitesSkipsDuplicateUserIDs(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("owner-dup-id@example.com"))
	created, _ := groupSvc.CreateGroup(ctx, owner.AccessToken, "DupIDs")
	target := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("target-dup@example.com"))

	svc := newInviteService(t, st)
	res, err := svc.SendGroupInvites(ctx, owner.AccessToken, models.SendGroupInvitesInput{
		GroupPublicID: created.ID,
		UserIDs:       []string{target.User.ID, target.User.ID, "  "},
	})
	if err != nil || res.Added != 1 || res.Skipped < 2 {
		t.Fatalf("%+v %v", res, err)
	}
}

func TestSendGroupInvitesAddsUserByEmail(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("owner-email-add@example.com"))
	created, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "EmailAdds")
	if err != nil {
		t.Fatal(err)
	}
	target := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("target-email@example.com"))

	svc := newInviteService(t, st)
	res, err := svc.SendGroupInvites(ctx, owner.AccessToken, models.SendGroupInvitesInput{
		GroupPublicID: created.ID,
		Emails:        []string{target.User.Email},
	})
	if err != nil || res.Added != 1 {
		t.Fatalf("%+v %v", res, err)
	}
}

func TestSendGroupInvitesSkipsExistingMemberEmail(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("owner-skip@example.com"))
	created, _ := groupSvc.CreateGroup(ctx, owner.AccessToken, "Skips")
	member := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("member-skip@example.com"))
	_ = testSupport.NewGroupMemberService(t, st).JoinGroup(ctx, member.AccessToken, created.ID)

	svc := newInviteService(t, st)
	res, err := svc.SendGroupInvites(ctx, owner.AccessToken, models.SendGroupInvitesInput{
		GroupPublicID: created.ID,
		Emails:        []string{member.User.Email},
	})
	if err != nil || res.Skipped != 1 || res.Added != 0 {
		t.Fatalf("%+v %v", res, err)
	}
}

func TestGetInviteLinkForbiddenForOutsider(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("owner-forbid@example.com"))
	created, _ := groupSvc.CreateGroup(ctx, owner.AccessToken, "Private")
	outsider := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("outsider-forbid@example.com"))

	svc := newInviteService(t, st)
	_, err := svc.GetInviteLink(ctx, outsider.AccessToken, created.ID, "")
	if codeOf(t, err) != string(exception.MsgForbidden) {
		t.Fatalf("%v", err)
	}
}

func TestPreviewInviteIncludesChannelName(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("owner-ch@example.com"))
	created, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "ChannelPrev")
	if err != nil {
		t.Fatal(err)
	}
	channels, err := testSupport.ChannelRepo(st).ListByGroupID(ctx, st.Groups[0].ID)
	if err != nil || len(channels) == 0 {
		t.Fatalf("channels %+v %v", channels, err)
	}

	svc := newInviteService(t, st)
	link, err := svc.GetInviteLink(ctx, owner.AccessToken, created.ID, channels[0].PublicID)
	if err != nil {
		t.Fatal(err)
	}

	out, err := svc.PreviewInvite(ctx, link.Token)
	if err != nil || !out.Valid {
		t.Fatalf("%+v %v", out, err)
	}
}

func TestRedeemEmailInviteWithChannel(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("owner-redeem-ch@example.com"))
	created, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "RedeemCh")
	if err != nil {
		t.Fatal(err)
	}
	channels, _ := testSupport.ChannelRepo(st).ListByGroupID(ctx, st.Groups[0].ID)
	chID := channels[0].ID
	chPub := channels[0].PublicID

	svc := newInviteService(t, st)
	email := "redeem-ch@example.com"
	res, err := svc.SendGroupInvites(ctx, owner.AccessToken, models.SendGroupInvitesInput{
		GroupPublicID:   created.ID,
		Emails:          []string{email},
		ChannelPublicID: chPub,
	})
	if err != nil || res.Emailed != 1 {
		t.Fatalf("%+v %v", res, err)
	}
	token := st.Invites[len(st.Invites)-1].Token
	_ = chID

	redeemer := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput(email))
	out, err := svc.RedeemInvite(ctx, redeemer.AccessToken, token)
	if err != nil || out.GroupPublicID != created.ID {
		t.Fatalf("%+v %v", out, err)
	}
}

func TestPreviewInviteExpiredShareLink(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("exp-preview@example.com"))
	created, _ := groupSvc.CreateGroup(ctx, owner.AccessToken, "Expired")
	svc := newInviteService(t, st)
	link, err := svc.GetInviteLink(ctx, owner.AccessToken, created.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	for i := range st.Invites {
		st.Invites[i].ExpiresAt = time.Now().UTC().Add(-time.Minute)
	}

	out, err := svc.PreviewInvite(ctx, link.Token)
	if err != nil || out.Valid || out.Reason != constants.InvitePreviewExpired {
		t.Fatalf("%+v %v", out, err)
	}
}

func TestRedeemShareLinkAddsMember(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("share-redeem@example.com"))
	created, _ := groupSvc.CreateGroup(ctx, owner.AccessToken, "Share")
	member := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("share-joiner@example.com"))

	svc := newInviteService(t, st)
	link, err := svc.GetInviteLink(ctx, owner.AccessToken, created.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	out, err := svc.RedeemInvite(ctx, member.AccessToken, link.Token)
	if err != nil || out.GroupPublicID != created.ID {
		t.Fatalf("%+v %v", out, err)
	}
	if err := testSupport.NewGroupMemberService(t, st).JoinGroup(ctx, member.AccessToken, created.ID); err != nil {
		t.Fatal("idempotent member")
	}
}

type failInviteNotifier struct{}

func (failInviteNotifier) Send(context.Context, *notificationv1.SendRequest) (*notificationv1.SendResponse, error) {
	return nil, errors.New("smtp down")
}

func TestSendGroupInvitesSkipsUnknownUserID(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("skip-unknown@example.com"))
	created, _ := groupSvc.CreateGroup(ctx, owner.AccessToken, "Skip")

	svc := newInviteService(t, st)
	res, err := svc.SendGroupInvites(ctx, owner.AccessToken, models.SendGroupInvitesInput{
		GroupPublicID: created.ID,
		UserIDs:       []string{"no-such-user-id"},
	})
	if err != nil || res.Skipped != 1 {
		t.Fatalf("%+v %v", res, err)
	}
}

func TestSendGroupInvitesEmailStillCountsWhenNotifyFails(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("fail-email@example.com"))
	created, _ := groupSvc.CreateGroup(ctx, owner.AccessToken, "Mail")

	users, groups, sessions := testSupport.TestRepos(st)
	svc := inviteservice.NewService(groups, groupmembermock.NewPgRepositoryMock(st), grouprolemock.NewPgRepositoryMock(st), testSupport.ChannelRepo(st), invitemock.NewPgRepositoryMock(st), users, sessions,
		failInviteNotifier{}, []byte(testSupport.TestJWTSecret), "https://app.test", "",
		logger.New(logger.Options{Service: "invite", Env: "test", Level: "error"}))

	res, err := svc.SendGroupInvites(ctx, owner.AccessToken, models.SendGroupInvitesInput{
		GroupPublicID: created.ID,
		Emails:        []string{"brand-new-invite@example.com"},
	})
	if err != nil || res.Emailed != 1 {
		t.Fatalf("%+v %v", res, err)
	}
}

func TestSendGroupInvitesAddsByUserID(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("owner-uid@example.com"))
	created, _ := groupSvc.CreateGroup(ctx, owner.AccessToken, "UID")
	target := testSupport.RegisterSession(t, authSvc, notify, testSupportRegisterInput("target-uid@example.com"))

	svc := newInviteService(t, st)
	res, err := svc.SendGroupInvites(ctx, owner.AccessToken, models.SendGroupInvitesInput{
		GroupPublicID: created.ID,
		UserIDs:       []string{target.User.ID, "", target.User.ID},
	})
	if err != nil || res.Added != 1 || res.Skipped != 2 {
		t.Fatalf("%+v %v", res, err)
	}
}
