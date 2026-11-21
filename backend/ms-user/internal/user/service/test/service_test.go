package service_test

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/pquerna/otp/totp"
	"github.com/redis/go-redis/v9"
	authrepo "github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	groupmembermock "github.com/wichat/wichat/backend/ms-user/internal/group-member/repository/mock"
	groupmock "github.com/wichat/wichat/backend/ms-user/internal/group/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	activitymock "github.com/wichat/wichat/backend/ms-user/internal/user-activity-log/repository/mock"
	activityservice "github.com/wichat/wichat/backend/ms-user/internal/user-activity-log/service"
	userrepo "github.com/wichat/wichat/backend/ms-user/internal/user/repository"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	userservice "github.com/wichat/wichat/backend/ms-user/internal/user/service"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	jwtpkg "github.com/wichat/wichat/backend/ms-user/pkg/jwt"
	testSupport "github.com/wichat/wichat/backend/ms-user/pkg/test-support"
	"github.com/wichat/wichat/backend/ms-user/pkg/username"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

const testPassword = "correct-horse-battery"

func registerInput(email string) models.RegisterInput {
	handle, _ := username.BaseFromEmail(email)
	return models.RegisterInput{
		Email:       email,
		Username:    handle,
		Password:    testPassword,
		DisplayName: "Ada",
		BirthDate:   "2000-01-15",
	}
}

func codeOf(t *testing.T, err error) string {
	t.Helper()

	ex, ok := exception.As(err)
	if !ok {
		t.Fatalf("not an exception: %v", err)
	}

	return string(ex.Code)
}

func TestSetupStatusDefaultsBeforeWizard(t *testing.T) {
	st := usermock.NewState()
	_, _, _, userSvc, _, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	status, err := userSvc.SetupStatus(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if status.RegistrationMode != constants.RegistrationOpen {
		t.Fatalf("%+v", status)
	}
}

func TestSetupStatusInviteOnlyFromEnv(t *testing.T) {
	st := usermock.NewState()
	_, _, _, userSvc, _, _ := testSupport.NewAuthService(t, st, nil, true)
	ctx := context.Background()

	status, err := userSvc.SetupStatus(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if status.RegistrationMode != constants.RegistrationInviteOnly {
		t.Fatalf("%+v", status)
	}
}

func TestCheckUsernameAvailability(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, userSvc, _, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	res, err := userSvc.CheckUsername(ctx, "  Ada.Lovelace ")
	if err != nil || !res.Available || res.Normalized != "ada.lovelace" || res.Reason != "" {
		t.Fatalf("available: %+v %v", res, err)
	}

	resBad, err := userSvc.CheckUsername(ctx, "admin")
	if err != nil || resBad.Available || resBad.Reason != username.ReasonReserved {
		t.Fatalf("reserved: %+v %v", resBad, err)
	}

	_ = testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email:       "ada-check@example.com",
		Username:    "ada.lovelace",
		Password:    testPassword,
		DisplayName: "Ada",
		BirthDate:   "2000-01-15",
	})

	resTaken, err := userSvc.CheckUsername(ctx, "ada.lovelace")
	if err != nil || resTaken.Available || resTaken.Reason != username.ReasonTaken {
		t.Fatalf("taken: %+v %v", resTaken, err)
	}
}

func TestCompleteInstanceSetupRejectsNonOwner(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, userSvc, _, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, registerInput("owner-setup@example.com"))
	if err := completeInstanceSetup(t, userSvc, owner.AccessToken); err != nil {
		t.Fatal(err)
	}

	member := testSupport.RegisterSession(t, authSvc, notify, registerInput("member-setup@example.com"))
	err := userSvc.CompleteInstanceSetup(ctx, member.AccessToken, models.CompleteInstanceSetupInput{
		Name: "X", RegistrationMode: constants.RegistrationOpen,
	})
	if codeOf(t, err) != string(exception.MsgForbidden) {
		t.Fatalf("%v", err)
	}
}

func TestCompleteInstanceSetupCreatesFirstGroup(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, userSvc, _, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, registerInput("admin@example.com"))
	if err := completeInstanceSetup(t, userSvc, owner.AccessToken); err != nil {
		t.Fatal(err)
	}

	_, groups, _ := testSupport.TestRepos(st)
	group, err := groups.GetFirstGroup(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if group.Name != "WiChat" {
		t.Fatalf("group name: %q", group.Name)
	}
	members := groupmembermock.NewPgRepositoryMock(st)
	member, err := members.FindGroupMember(ctx, group.ID, owner.User.ID)
	if err != nil || member.Role != constants.GroupRoleOwner {
		t.Fatalf("membership: %+v err %v", member, err)
	}

	status, err := userSvc.SetupStatus(ctx)
	if err != nil || status.RegistrationMode != constants.RegistrationOpen {
		t.Fatalf("status: %v %+v", err, status)
	}
}

func TestCompleteInstanceSetupOwnerOnly(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, userSvc, _, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, registerInput("admin@example.com"))
	if err := completeInstanceSetup(t, userSvc, owner.AccessToken); err != nil {
		t.Fatal(err)
	}

	err := userSvc.CompleteInstanceSetup(ctx, owner.AccessToken, models.CompleteInstanceSetupInput{
		Name:             "Again",
		RegistrationMode: constants.RegistrationOpen,
	})
	if codeOf(t, err) != string(exception.MsgSetupClosed) {
		t.Fatalf("second complete: %v", err)
	}
}

func completeInstanceSetup(t *testing.T, svc userservice.Service, token string) error {
	t.Helper()

	return svc.CompleteInstanceSetup(context.Background(), token, models.CompleteInstanceSetupInput{
		Name:             "WiChat",
		RegistrationMode: constants.RegistrationInviteOnly,
	})
}

func TestSearchUsersInGroup(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, userSvc, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, registerInput("owner-search@example.com"))
	created, err := groupSvc.CreateGroup(ctx, owner.AccessToken, "Lookup")
	if err != nil {
		t.Fatal(err)
	}
	target := testSupport.RegisterSession(t, authSvc, notify, registerInput("findme@example.com"))

	hits, err := userSvc.SearchUsers(ctx, owner.AccessToken, "find", 10, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, h := range hits {
		if h.UserID == target.User.ID {
			found = true
		}
	}
	if !found {
		t.Fatalf("hits: %+v", hits)
	}
}

func TestSearchUsersEmptyQuery(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, userSvc, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, authSvc, notify, registerInput("search-empty@example.com"))
	created, _ := groupSvc.CreateGroup(ctx, owner.AccessToken, "Search")

	hits, err := userSvc.SearchUsers(ctx, owner.AccessToken, "  ", 10, created.ID)
	if err != nil || len(hits) != 0 {
		t.Fatalf("%+v %v", hits, err)
	}
}

type fakeUserMinio struct {
	removed []string
	nextURL string
}

func (f *fakeUserMinio) PutUserAvatar(_ context.Context, _, _ string, _ io.Reader, _ int64) (string, error) {
	return f.nextURL, nil
}

func (f *fakeUserMinio) RemoveUserAvatar(_ context.Context, userID string) error {
	f.removed = append(f.removed, userID)
	return nil
}

func TestUpdateUserProfileAndAvatar(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, users, _, _, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	_, _, sessions := testSupport.TestRepos(st)
	sess := testSupport.RegisterSession(t, authSvc, notify, registerInput("profile@example.com"))
	fakeMinio := &fakeUserMinio{nextURL: "http://localhost:9000/wichat/users/u1/avatar.png"}
	log := logger.New(logger.Options{Service: "user-test", Env: "test", Level: "error"})
	svc := userservice.NewService(
		users,
		groupmock.NewPgRepositoryMock(st),
		groupmembermock.NewPgRepositoryMock(st),
		fakeMinio,
		sessions,
		nil,
		nil,
		nil,
		[]byte(testSupport.TestJWTSecret),
		log,
	)

	// UploadUserAvatar rejects invalid content type
	if _, err := svc.UploadUserAvatar(ctx, sess.AccessToken, "text/plain", strings.NewReader("x"), 1); codeOf(t, err) != string(exception.MsgInvalidInput) {
		t.Fatalf("expected invalid_input for bad mime, got %v", err)
	}

	// UploadUserAvatar succeeds with image/png
	withAvatar, err := svc.UploadUserAvatar(ctx, sess.AccessToken, "image/png", strings.NewReader("png-bytes"), 9)
	if err != nil {
		t.Fatalf("upload avatar: %v", err)
	}
	if withAvatar.AvatarURL != fakeMinio.nextURL {
		t.Fatalf("unexpected avatar url: %q", withAvatar.AvatarURL)
	}

	// UpdateUserProfile updates scalars
	updated, err := svc.UpdateUserProfile(ctx, sess.AccessToken, "Ada L.", "she/her", "Hello WiChat", "#1e1f22", false)
	if err != nil {
		t.Fatalf("update profile: %v", err)
	}
	if updated.DisplayName != "Ada L." || updated.Pronouns != "she/her" || updated.Bio != "Hello WiChat" || updated.BannerColor != "#1E1F22" || updated.AvatarURL != fakeMinio.nextURL {
		t.Fatalf("unexpected updated profile: %+v", updated)
	}

	// UpdateUserProfile clears avatar when requested
	cleared, err := svc.UpdateUserProfile(ctx, sess.AccessToken, "Ada L.", "", "", "", true)
	if err != nil {
		t.Fatalf("clear avatar: %v", err)
	}
	if cleared.AvatarURL != "" || len(fakeMinio.removed) != 1 {
		t.Fatalf("expected avatar cleared and removed from minio, got %+v removed=%v", cleared, fakeMinio.removed)
	}

	// Validation error on invalid hex color
	if _, err := svc.UpdateUserProfile(ctx, sess.AccessToken, "Ada L.", "", "", "not-hex", false); codeOf(t, err) != string(exception.MsgInvalidInput) {
		t.Fatalf("expected invalid_input for bad banner_color, got %v", err)
	}
}

func TestUpdateUserStatusNote(t *testing.T) {
	ctx := context.Background()
	st := usermock.NewState()
	authSvc, notify, users, _, _, _ := testSupport.NewAuthService(t, st, nil)
	_, _, sessions := testSupport.TestRepos(st)

	sess := testSupport.RegisterSession(t, authSvc, notify, registerInput("status-note@example.com"))
	log := logger.New(logger.Options{Service: "user-test", Env: "test", Level: "error"})
	svc := userservice.NewService(
		users,
		groupmock.NewPgRepositoryMock(st),
		groupmembermock.NewPgRepositoryMock(st),
		nil,
		sessions,
		nil,
		nil,
		nil,
		[]byte(testSupport.TestJWTSecret),
		log,
	)

	exp := time.Now().UTC().Add(time.Hour).Format(time.RFC3339)
	updated, err := svc.UpdateUserStatusNote(ctx, sess.AccessToken, "Deep work", "💻", exp)
	if err != nil {
		t.Fatalf("update status note: %v", err)
	}
	if updated.StatusNoteText != "Deep work" || updated.StatusNoteEmoji != "💻" || updated.StatusNoteExpiresAt == nil {
		t.Fatalf("unexpected status note: %+v", updated)
	}

	cleared, err := svc.UpdateUserStatusNote(ctx, sess.AccessToken, "", "", "")
	if err != nil {
		t.Fatalf("clear status note: %v", err)
	}
	if cleared.StatusNoteText != "" || cleared.StatusNoteEmoji != "" || cleared.StatusNoteExpiresAt != nil {
		t.Fatalf("expected cleared status note, got %+v", cleared)
	}

	if _, err := svc.UpdateUserStatusNote(ctx, sess.AccessToken, strings.Repeat("x", constants.MaxUserStatusNoteRunes+1), "", ""); codeOf(t, err) != string(exception.MsgInvalidInput) {
		t.Fatalf("expected invalid_input for long status note, got %v", err)
	}
}

func TestUpdateUserPreferences(t *testing.T) {
	ctx := context.Background()
	st := usermock.NewState()
	authSvc, notify, users, _, _, _ := testSupport.NewAuthService(t, st, nil)
	_, _, sessions := testSupport.TestRepos(st)

	sess := testSupport.RegisterSession(t, authSvc, notify, registerInput("prefs@example.com"))
	log := logger.New(logger.Options{Service: "user-test", Env: "test", Level: "error"})
	svc := userservice.NewService(
		users,
		groupmock.NewPgRepositoryMock(st),
		groupmembermock.NewPgRepositoryMock(st),
		nil,
		sessions,
		nil,
		nil,
		nil,
		[]byte(testSupport.TestJWTSecret),
		log,
	)

	updated, err := svc.UpdateUserPreferences(ctx, sess.AccessToken, "id", "24h")
	if err != nil {
		t.Fatalf("update preferences: %v", err)
	}
	if updated.Locale != "id" || updated.TimeFormat != "24h" {
		t.Fatalf("unexpected preferences: %+v", updated)
	}

	if _, err := svc.UpdateUserPreferences(ctx, sess.AccessToken, "fr", "auto"); codeOf(t, err) != string(exception.MsgInvalidInput) {
		t.Fatalf("expected invalid_input for bad locale, got %v", err)
	}
}

func TestUpdateUsername(t *testing.T) {
	ctx := context.Background()
	st := usermock.NewState()
	authSvc, notify, users, _, _, _ := testSupport.NewAuthService(t, st, nil)
	_, _, sessions := testSupport.TestRepos(st)

	in1 := registerInput("alice@example.com")
	in1.Username = "alice_one"
	first := testSupport.RegisterSession(t, authSvc, notify, in1)

	in2 := registerInput("bob@example.com")
	in2.Username = "bob_two"
	second := testSupport.RegisterSession(t, authSvc, notify, in2)

	log := logger.New(logger.Options{Service: "user-test", Env: "test", Level: "error"})
	svc := userservice.NewService(
		users,
		groupmock.NewPgRepositoryMock(st),
		groupmembermock.NewPgRepositoryMock(st),
		nil,
		sessions,
		nil,
		nil,
		nil,
		[]byte(testSupport.TestJWTSecret),
		log,
	)

	// Invalid username format
	if _, err := svc.UpdateUsername(ctx, first.AccessToken, "ab", in1.Password); codeOf(t, err) != string(exception.MsgInvalidInput) {
		t.Fatalf("expected invalid_input, got %v", err)
	}

	// Wrong password
	if _, err := svc.UpdateUsername(ctx, first.AccessToken, "alice.new", "WrongPass123!"); codeOf(t, err) != string(exception.MsgInvalidCredentials) {
		t.Fatalf("expected invalid_credentials, got %v", err)
	}

	// Taken username
	if _, err := svc.UpdateUsername(ctx, first.AccessToken, second.User.Username, in1.Password); codeOf(t, err) != string(exception.MsgUsernameTaken) {
		t.Fatalf("expected username_taken, got %v", err)
	}

	// Success (normalizes case)
	updated, err := svc.UpdateUsername(ctx, first.AccessToken, "Alice.New_1", in1.Password)
	if err != nil {
		t.Fatalf("update username: %v", err)
	}
	if updated.Username != "alice.new_1" {
		t.Fatalf("expected alice.new_1, got %q", updated.Username)
	}
}

func TestEmailChangeFlow(t *testing.T) {
	ctx := context.Background()
	st := usermock.NewState()
	authSvc, notify, users, _, _, _ := testSupport.NewAuthService(t, st, nil)
	_, _, sessions := testSupport.TestRepos(st)

	in1 := registerInput("change-me@example.com")
	in1.Username = "change_me"
	first := testSupport.RegisterSession(t, authSvc, notify, in1)

	in2 := registerInput("taken@example.com")
	in2.Username = "taken_user"
	_ = testSupport.RegisterSession(t, authSvc, notify, in2)

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	log := logger.New(logger.Options{Service: "user-test", Env: "test", Level: "error"})
	svc := userservice.NewService(
		users,
		groupmock.NewPgRepositoryMock(st),
		groupmembermock.NewPgRepositoryMock(st),
		nil,
		sessions,
		authrepo.NewRedisEmailAuthStore(rdb),
		userrepo.NewRedisEmailChangeStore(rdb),
		notify,
		[]byte(testSupport.TestJWTSecret),
		log,
	)

	// 0. Unavailable when stores are nil
	nilSvc := userservice.NewService(users, groupmock.NewPgRepositoryMock(st), groupmembermock.NewPgRepositoryMock(st), nil, sessions, nil, nil, nil, []byte(testSupport.TestJWTSecret), log)
	if _, err := nilSvc.RequestEmailChange(ctx, first.AccessToken, "en"); codeOf(t, err) != string(exception.MsgUnavailable) {
		t.Fatalf("expected unavailable, got %v", err)
	}
	if _, err := nilSvc.VerifyEmailChange(ctx, first.AccessToken, "ch", "123456"); codeOf(t, err) != string(exception.MsgUnavailable) {
		t.Fatalf("expected unavailable, got %v", err)
	}
	if _, err := nilSvc.CompleteEmailChange(ctx, first.AccessToken, "tok", "a@example.com", in1.Password, "other"); codeOf(t, err) != string(exception.MsgUnavailable) {
		t.Fatalf("expected unavailable, got %v", err)
	}

	// 1. Request OTP
	ch, err := svc.RequestEmailChange(ctx, first.AccessToken, "en")
	if err != nil {
		t.Fatalf("request email change: %v", err)
	}
	if ch.ChallengeID == "" || ch.ResendCooldownSeconds <= 0 {
		t.Fatalf("unexpected challenge: %+v", ch)
	}

	// 2. Back-to-back request hits cooldown
	if _, err := svc.RequestEmailChange(ctx, first.AccessToken, "en"); codeOf(t, err) != string(exception.MsgResetRateLimited) {
		t.Fatalf("expected rate limited, got %v", err)
	}

	// 3. Empty/wrong OTP fails
	if _, err := svc.VerifyEmailChange(ctx, first.AccessToken, "", ""); codeOf(t, err) != string(exception.MsgResetCodeInvalid) {
		t.Fatalf("expected reset_code_invalid on empty, got %v", err)
	}
	if _, err := svc.VerifyEmailChange(ctx, first.AccessToken, "missing-ch", "123456"); codeOf(t, err) != string(exception.MsgResetCodeInvalid) {
		t.Fatalf("expected reset_code_invalid on missing challenge, got %v", err)
	}
	if _, err := svc.VerifyEmailChange(ctx, first.AccessToken, ch.ChallengeID, "000000"); codeOf(t, err) != string(exception.MsgResetCodeInvalid) {
		t.Fatalf("expected reset_code_invalid, got %v", err)
	}

	// 4. Correct OTP returns change token
	changeToken, err := svc.VerifyEmailChange(ctx, first.AccessToken, ch.ChallengeID, notify.LastCode)
	if err != nil || changeToken == "" {
		t.Fatalf("verify email change: %v token=%q", err, changeToken)
	}

	// 5. Empty token / invalid email / wrong password fail
	if _, err := svc.CompleteEmailChange(ctx, first.AccessToken, "", "fresh@example.com", in1.Password, "new_email"); codeOf(t, err) != string(exception.MsgResetCodeInvalid) {
		t.Fatalf("expected reset_code_invalid on empty token, got %v", err)
	}
	if _, err := svc.CompleteEmailChange(ctx, first.AccessToken, changeToken, "not-an-email", in1.Password, "new_email"); codeOf(t, err) != string(exception.MsgInvalidInput) {
		t.Fatalf("expected invalid_input on bad email, got %v", err)
	}
	if _, err := svc.CompleteEmailChange(ctx, first.AccessToken, changeToken, "fresh@example.com", "WrongPass123!", "new_email"); codeOf(t, err) != string(exception.MsgInvalidCredentials) {
		t.Fatalf("expected invalid_credentials, got %v", err)
	}

	// 6. Duplicate email fails with conflict
	if _, err := svc.CompleteEmailChange(ctx, first.AccessToken, changeToken, "taken@example.com", in1.Password, "new_email"); codeOf(t, err) != string(exception.MsgEmailTaken) {
		t.Fatalf("expected email_taken, got %v", err)
	}

	// 7. Re-verify? Wait: step 6 consumed token, so request & verify again after clearing cooldown
	mr.FastForward(2 * 60_000_000_000) // 2 minutes
	ch2, err := svc.RequestEmailChange(ctx, first.AccessToken, "en")
	if err != nil {
		t.Fatalf("second request email change: %v", err)
	}
	changeToken2, err := svc.VerifyEmailChange(ctx, first.AccessToken, ch2.ChallengeID, notify.LastCode)
	if err != nil {
		t.Fatalf("second verify email change: %v", err)
	}

	// 8. Complete succeeds
	updated, err := svc.CompleteEmailChange(ctx, first.AccessToken, changeToken2, "Fresh.Email@example.com", in1.Password, "new_email")
	if err != nil {
		t.Fatalf("complete email change: %v", err)
	}
	if updated.Email != "fresh.email@example.com" {
		t.Fatalf("expected fresh.email@example.com, got %q", updated.Email)
	}

	// 9. Reusing consumed token fails
	if _, err := svc.CompleteEmailChange(ctx, first.AccessToken, changeToken2, "another@example.com", in1.Password, "other"); codeOf(t, err) != string(exception.MsgResetCodeInvalid) {
		t.Fatalf("expected reset_code_invalid on reused token, got %v", err)
	}
}

func TestUpdatePassword(t *testing.T) {
	ctx := context.Background()
	st := usermock.NewState()
	authSvc, notify, users, _, _, _ := testSupport.NewAuthService(t, st, nil)
	_, _, sessions := testSupport.TestRepos(st)

	in := registerInput("pw-update@example.com")
	in.Username = "pw_update"
	sess := testSupport.RegisterSession(t, authSvc, notify, in)

	log := logger.New(logger.Options{Service: "user-test", Env: "test", Level: "error"})
	svc := userservice.NewService(
		users,
		groupmock.NewPgRepositoryMock(st),
		groupmembermock.NewPgRepositoryMock(st),
		nil,
		sessions,
		nil,
		nil,
		nil,
		[]byte(testSupport.TestJWTSecret),
		log,
	)

	// Short new password fails
	if err := svc.UpdatePassword(ctx, sess.AccessToken, in.Password, "short"); codeOf(t, err) != string(exception.MsgInvalidInput) {
		t.Fatalf("expected invalid_input on short password, got %v", err)
	}

	// Wrong current password fails
	if err := svc.UpdatePassword(ctx, sess.AccessToken, "WrongPass123!", "new-strong-password-123"); codeOf(t, err) != string(exception.MsgInvalidCredentials) {
		t.Fatalf("expected invalid_credentials, got %v", err)
	}

	// Valid update succeeds
	if err := svc.UpdatePassword(ctx, sess.AccessToken, in.Password, "new-strong-password-123"); err != nil {
		t.Fatalf("update password: %v", err)
	}

	// Old password no longer works, new password works
	if err := svc.UpdatePassword(ctx, sess.AccessToken, in.Password, "another-strong-password-123"); codeOf(t, err) != string(exception.MsgInvalidCredentials) {
		t.Fatalf("expected old password to be rejected, got %v", err)
	}
	if err := svc.UpdatePassword(ctx, sess.AccessToken, "new-strong-password-123", "another-strong-password-123"); err != nil {
		t.Fatalf("expected new password to be accepted, got %v", err)
	}

	// OAuth-only user (empty password_hash) can set up a password without current_password.
	oauthUser := models.User{
		ID:          "oauth-u1",
		Email:       "oauth-pw@example.com",
		Username:    "oauth_pw",
		DisplayName: "OAuth User",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	st.Users = append(st.Users, oauthUser)
	st.Sessions = append(st.Sessions, models.Session{
		ID:        "oauth-sid-1",
		UserID:    oauthUser.ID,
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	})
	oauthTok, _, err := jwtpkg.Issue([]byte(testSupport.TestJWTSecret), oauthUser.ID, "oauth-sid-1", time.Now().UTC(), time.Hour)
	if err != nil {
		t.Fatalf("sign oauth token: %v", err)
	}
	if err := svc.UpdatePassword(ctx, oauthTok, "", "first-oauth-password-123"); err != nil {
		t.Fatalf("expected oauth first-time password setup to succeed, got %v", err)
	}
	if err := svc.UpdatePassword(ctx, oauthTok, "", "second-oauth-password-123"); codeOf(t, err) != string(exception.MsgInvalidCredentials) {
		t.Fatalf("expected empty current_password to be rejected after setup, got %v", err)
	}
}

func TestMFATOTPAndSecurityKeyManagement(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, userSvc, _, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	in := registerInput("mfa-user@example.com")
	sess := testSupport.RegisterSession(t, authSvc, notify, in)

	st0, err := userSvc.GetMFAStatus(ctx, sess.AccessToken)
	if err != nil || st0.TOTPEnabled || len(st0.SecurityKeys) != 0 || st0.RecoveryCodesRemaining != 0 {
		t.Fatalf("unexpected initial mfa status: %+v %v", st0, err)
	}

	// Begin TOTP requires valid current password
	if _, err := userSvc.BeginTOTPSetup(ctx, sess.AccessToken, "wrong-password"); codeOf(t, err) != string(exception.MsgInvalidCredentials) {
		t.Fatalf("expected invalid_credentials, got %v", err)
	}

	setup, err := userSvc.BeginTOTPSetup(ctx, sess.AccessToken, in.Password)
	if err != nil || setup.SetupToken == "" || setup.Secret == "" {
		t.Fatalf("BeginTOTPSetup() = %+v, %v", setup, err)
	}

	code, err := totp.GenerateCode(setup.Secret, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	codes, err := userSvc.ConfirmTOTPSetup(ctx, sess.AccessToken, setup.SetupToken, code)
	if err != nil || len(codes) != constants.RecoveryCodeCount {
		t.Fatalf("ConfirmTOTPSetup() = %d codes, %v", len(codes), err)
	}

	st1, err := userSvc.GetMFAStatus(ctx, sess.AccessToken)
	if err != nil || !st1.TOTPEnabled || st1.RecoveryCodesRemaining != constants.RecoveryCodeCount {
		t.Fatalf("unexpected mfa status after TOTP: %+v %v", st1, err)
	}

	// Regenerate recovery codes
	regen, err := userSvc.RegenerateRecoveryCodes(ctx, sess.AccessToken, in.Password)
	if err != nil || len(regen) != constants.RecoveryCodeCount {
		t.Fatalf("RegenerateRecoveryCodes() = %d, %v", len(regen), err)
	}

	// Begin WebAuthn register returns valid options JSON; FinishWebAuthnRegister rejects invalid payload/challenge
	waBegin, err := userSvc.BeginWebAuthnRegister(ctx, sess.AccessToken, in.Password, "MacBook Touch ID")
	if err != nil || waBegin.ChallengeID == "" || !strings.Contains(waBegin.OptionsJSON, "publicKey") {
		t.Fatalf("BeginWebAuthnRegister() = %+v, %v", waBegin, err)
	}
	if _, _, err := userSvc.FinishWebAuthnRegister(ctx, sess.AccessToken, "missing-chal", `{}`); err == nil {
		t.Fatal("expected missing challenge to fail FinishWebAuthnRegister")
	}
	if _, _, err := userSvc.FinishWebAuthnRegister(ctx, sess.AccessToken, waBegin.ChallengeID, `{}`); err == nil {
		t.Fatal("expected invalid credential JSON to fail FinishWebAuthnRegister")
	}

	// Wrong-password & not-found guards
	if _, err := userSvc.BeginTOTPSetup(ctx, sess.AccessToken, "wrong-password"); err == nil {
		t.Fatal("expected wrong password to fail BeginTOTPSetup")
	}
	if err := userSvc.DisableTOTP(ctx, sess.AccessToken, "wrong-password"); err == nil {
		t.Fatal("expected wrong password to fail DisableTOTP")
	}
	if _, err := userSvc.RenameWebAuthnCredential(ctx, sess.AccessToken, "missing", "New"); err == nil {
		t.Fatal("expected missing credential to fail RenameWebAuthnCredential")
	}
	if err := userSvc.DeleteWebAuthnCredential(ctx, sess.AccessToken, "key-1", "wrong-password"); err == nil {
		t.Fatal("expected wrong password to fail DeleteWebAuthnCredential")
	}
	if _, err := userSvc.RegenerateRecoveryCodes(ctx, sess.AccessToken, "wrong-password"); err == nil {
		t.Fatal("expected wrong password to fail RegenerateRecoveryCodes")
	}
	if _, err := userSvc.ConfirmTOTPSetup(ctx, sess.AccessToken, "missing-token", "123456"); err == nil {
		t.Fatal("expected missing setup token to fail ConfirmTOTPSetup")
	}
	if _, err := userSvc.ConfirmTOTPSetup(ctx, sess.AccessToken, setup.SetupToken, "000000"); err == nil {
		t.Fatal("expected already-consumed setup token to fail ConfirmTOTPSetup")
	}
	if badSetup, err := userSvc.BeginTOTPSetup(ctx, sess.AccessToken, in.Password); err == nil {
		// TOTP already enabled -> should fail; or if allowed, wrong code fails
		_, _ = userSvc.ConfirmTOTPSetup(ctx, sess.AccessToken, badSetup.SetupToken, "000000")
	}
	if err := userSvc.DeleteWebAuthnCredential(ctx, sess.AccessToken, "nonexistent", in.Password); err == nil {
		t.Fatal("expected nonexistent credential to fail DeleteWebAuthnCredential")
	}
	if _, err := userSvc.UpdateUserProfile(ctx, sess.AccessToken, "Name", "", "", "not-a-hex", false); err == nil {
		t.Fatal("expected invalid banner color to fail UpdateUserProfile")
	}
	if err := userSvc.UpdatePassword(ctx, sess.AccessToken, "wrong-pass", "NewStrongPass123!"); err == nil {
		t.Fatal("expected wrong current password to fail UpdatePassword")
	}
	if err := userSvc.UpdatePassword(ctx, sess.AccessToken, in.Password, "short"); err == nil {
		t.Fatal("expected short new password to fail UpdatePassword")
	}
	if _, err := userSvc.CompleteEmailChange(ctx, sess.AccessToken, "bad-token", "new@example.com", in.Password, "other"); err == nil {
		t.Fatal("expected bad change token to fail CompleteEmailChange")
	}
	if _, err := userSvc.UpdateUsername(ctx, sess.AccessToken, "bad!name", in.Password); err == nil {
		t.Fatal("expected invalid username to fail UpdateUsername")
	}
	if _, err := userSvc.UpdateUsername(ctx, sess.AccessToken, "mfa.test.new", "wrong-pass"); err == nil {
		t.Fatal("expected wrong password to fail UpdateUsername")
	}
	if _, err := userSvc.VerifyEmailChange(ctx, sess.AccessToken, "missing-chal", "000000"); err == nil {
		t.Fatal("expected missing challenge to fail VerifyEmailChange")
	}
	// Fresh TOTP setup with wrong code first
	totpRetry, err := userSvc.BeginTOTPSetup(ctx, sess.AccessToken, in.Password)
	if err == nil {
		if _, err := userSvc.ConfirmTOTPSetup(ctx, sess.AccessToken, totpRetry.SetupToken, "000000"); err == nil {
			t.Fatal("expected wrong TOTP code to fail ConfirmTOTPSetup")
		}
	}
	if _, err := userSvc.UploadUserAvatar(ctx, sess.AccessToken, "text/plain", strings.NewReader("hi"), 2); err == nil {
		t.Fatal("expected invalid avatar content type to fail")
	}
	if _, err := userSvc.UploadUserAvatar(ctx, sess.AccessToken, "image/png", strings.NewReader(""), 0); err == nil {
		t.Fatal("expected empty avatar data to fail")
	}
	if _, err := userSvc.RequestEmailChange(ctx, "invalid-token", "en"); err == nil {
		t.Fatal("expected invalid token to fail RequestEmailChange")
	}
	if err := userSvc.CompleteInstanceSetup(ctx, sess.AccessToken, models.CompleteInstanceSetupInput{}); err == nil {
		t.Fatal("expected empty instance name to fail CompleteInstanceSetup")
	}
	if _, err := userSvc.GetMFAStatus(ctx, "invalid-token"); err == nil {
		t.Fatal("expected invalid token to fail GetMFAStatus")
	}
	_, _ = userSvc.SearchUsers(ctx, sess.AccessToken, "mfa", 10, "")
	_, _ = userSvc.SearchUsers(ctx, sess.AccessToken, "", 10, "")

	// Pre-populate a security key in mock state to test rename & delete
	st.Lock()
	st.WebAuthnCreds = append(st.WebAuthnCreds, models.WebAuthnCredential{
		ID:           "key-1",
		UserID:       sess.User.ID,
		Name:         "YubiKey",
		CredentialID: []byte("cred-id-1"),
		PublicKey:    []byte("pk-1"),
		CreatedAt:    time.Now().UTC(),
	})
	st.Unlock()

	if waBegin2, err := userSvc.BeginWebAuthnRegister(ctx, sess.AccessToken, in.Password, "Backup Key"); err != nil || waBegin2.ChallengeID == "" {
		t.Fatalf("BeginWebAuthnRegister with existing key: %v", err)
	}

	renamed, err := userSvc.RenameWebAuthnCredential(ctx, sess.AccessToken, "key-1", "YubiKey 5C")
	if err != nil || renamed.Name != "YubiKey 5C" {
		t.Fatalf("RenameWebAuthnCredential() = %+v, %v", renamed, err)
	}

	if err := userSvc.DeleteWebAuthnCredential(ctx, sess.AccessToken, "key-1", in.Password); err != nil {
		t.Fatalf("DeleteWebAuthnCredential() = %v", err)
	}

	// Disable TOTP cleans up recovery codes when no MFA remains
	if err := userSvc.DisableTOTP(ctx, sess.AccessToken, in.Password); err != nil {
		t.Fatalf("DisableTOTP() = %v", err)
	}
	st2, err := userSvc.GetMFAStatus(ctx, sess.AccessToken)
	if err != nil || st2.TOTPEnabled || st2.RecoveryCodesRemaining != 0 {
		t.Fatalf("expected clean MFA status after disable, got %+v %v", st2, err)
	}
	_, _ = userSvc.RegenerateRecoveryCodes(ctx, sess.AccessToken, in.Password)
	if err := userSvc.DisableTOTP(ctx, sess.AccessToken, in.Password); err == nil {
		t.Fatal("expected DisableTOTP to fail when already disabled")
	}
	if _, err := userSvc.UpdateUserProfile(ctx, sess.AccessToken, "MFA User Updated", "they/them", "Bio", "#112233", true); err != nil {
		t.Fatalf("UpdateUserProfile valid: %v", err)
	}
	if _, err := userSvc.UpdateUsername(ctx, sess.AccessToken, "mfa.updated", in.Password); err != nil {
		t.Fatalf("UpdateUsername valid: %v", err)
	}
	reqCh, err := userSvc.RequestEmailChange(ctx, sess.AccessToken, "en")
	if err != nil {
		t.Fatalf("RequestEmailChange: %v", err)
	}
	chgTok, err := userSvc.VerifyEmailChange(ctx, sess.AccessToken, reqCh.ChallengeID, notify.LastCode)
	if err != nil {
		t.Fatalf("VerifyEmailChange: %v", err)
	}
	if _, err := userSvc.CompleteEmailChange(ctx, sess.AccessToken, chgTok, "mfa-new@example.com", in.Password, "new_email"); err != nil {
		t.Fatalf("CompleteEmailChange: %v", err)
	}
}

func TestDisableAndDeleteAccount(t *testing.T) {
	ctx := context.Background()
	st := usermock.NewState()
	authSvc, notify, users, userSvc, groupSvc, _ := testSupport.NewAuthService(t, st, nil)

	// 1. User owning a group is blocked from disabling or deleting.
	ownerIn := registerInput("owner-life@example.com")
	ownerIn.Username = "owner_life"
	ownerSess := testSupport.RegisterSession(t, authSvc, notify, ownerIn)
	grp, err := groupSvc.CreateGroup(ctx, ownerSess.AccessToken, "Owned Group")
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}
	if err := userSvc.DisableAccount(ctx, ownerSess.AccessToken, ownerIn.Password); codeOf(t, err) != string(exception.MsgUserOwnsGroups) {
		t.Fatalf("expected user.owns_groups on disable, got %v", err)
	}
	if err := userSvc.DeleteAccount(ctx, ownerSess.AccessToken, ownerIn.Password); codeOf(t, err) != string(exception.MsgUserOwnsGroups) {
		t.Fatalf("expected user.owns_groups on delete, got %v", err)
	}

	// 2. After deleting the group, wrong password fails and valid DisableAccount succeeds.
	if err := groupSvc.DeleteGroup(ctx, ownerSess.AccessToken, grp.ID); err != nil {
		t.Fatalf("DeleteGroup: %v", err)
	}
	if err := userSvc.DisableAccount(ctx, ownerSess.AccessToken, "WrongPassword123!"); codeOf(t, err) != string(exception.MsgInvalidCredentials) {
		t.Fatalf("expected invalid_credentials on disable, got %v", err)
	}
	if err := userSvc.DisableAccount(ctx, ownerSess.AccessToken, ownerIn.Password); err != nil {
		t.Fatalf("DisableAccount: %v", err)
	}
	if _, err := authSvc.Refresh(ctx, ownerSess.RefreshToken, models.ClientMeta{}); codeOf(t, err) != string(exception.MsgUnauthorized) {
		t.Fatalf("expected revoked refresh after disable, got %v", err)
	}
	if _, err := authSvc.Login(ctx, models.LoginInput{Login: ownerIn.Email, Password: ownerIn.Password}); codeOf(t, err) != string(exception.MsgInvalidCredentials) {
		t.Fatalf("expected neutral invalid_credentials on login for disabled account, got %v", err)
	}

	// 3. DeleteAccount anonymizes PII and revokes sessions.
	delIn := registerInput("del-user@example.com")
	delIn.Username = "del_user"
	delSess := testSupport.RegisterSession(t, authSvc, notify, delIn)
	if err := userSvc.DeleteAccount(ctx, delSess.AccessToken, delIn.Password); err != nil {
		t.Fatalf("DeleteAccount: %v", err)
	}
	deletedUser, err := users.FindUserByID(ctx, delSess.User.ID)
	if err != nil {
		t.Fatalf("FindUserByID after delete: %v", err)
	}
	if deletedUser.DeletedAt == nil || deletedUser.Email != "" || deletedUser.DisplayName != "Deleted User" || !strings.HasPrefix(deletedUser.Username, "deleted_") {
		t.Fatalf("unexpected anonymized user: %+v", deletedUser)
	}
	if _, err := authSvc.Refresh(ctx, delSess.RefreshToken, models.ClientMeta{}); codeOf(t, err) != string(exception.MsgUnauthorized) {
		t.Fatalf("expected revoked refresh after delete, got %v", err)
	}

	// 4. OAuth-only user (no password_hash) can delete account without current_password.
	oauthUser := models.User{
		ID:          "oauth-del-1",
		Email:       "oauth-del@example.com",
		Username:    "oauth_del",
		DisplayName: "OAuth Del",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	st.Users = append(st.Users, oauthUser)
	st.Sessions = append(st.Sessions, models.Session{
		ID:        "oauth-del-sid",
		UserID:    oauthUser.ID,
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	})
	oauthTok, _, err := jwtpkg.Issue([]byte(testSupport.TestJWTSecret), oauthUser.ID, "oauth-del-sid", time.Now().UTC(), time.Hour)
	if err != nil {
		t.Fatalf("issue oauth token: %v", err)
	}
	if err := userSvc.DeleteAccount(ctx, oauthTok, ""); err != nil {
		t.Fatalf("DeleteAccount oauth user: %v", err)
	}
}

func TestUserSessionsListAndRevoke(t *testing.T) {
	ctx := context.Background()
	st := usermock.NewState()
	authSvc, notify, _, userSvc, _, _ := testSupport.NewAuthService(t, st, nil)

	in := registerInput("devices@example.com")
	in.Username = "devices_user"
	sess1 := testSupport.RegisterSession(t, authSvc, notify, in)

	// Log in a second device session.
	loginOut, err := authSvc.Login(ctx, models.LoginInput{
		Login:    in.Email,
		Password: in.Password,
		ClientMeta: models.ClientMeta{
			UserAgent:     "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X)",
			ClientIP:      "203.0.113.10",
			Platform:      "mobile",
			DeviceLabel:   "Safari · iOS",
			LocationLabel: "Jakarta",
		},
		Locale: "en",
	})
	if err != nil {
		t.Fatalf("second Login: %v", err)
	}
	var sess2 *models.SessionResult
	if loginOut.Session != nil {
		sess2 = loginOut.Session
	} else {
		code := notify.LastCode
		sess2, err = authSvc.VerifyLogin(ctx, loginOut.Challenge.ChallengeID, code)
		if err != nil {
			t.Fatalf("VerifyLogin: %v", err)
		}
	}

	listed, currentID, err := userSvc.ListSessions(ctx, sess1.AccessToken)
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(listed) != 2 || currentID == "" {
		t.Fatalf("expected 2 active sessions, got %d (current=%q)", len(listed), currentID)
	}

	var otherID string
	for _, s := range listed {
		if s.ID != currentID {
			otherID = s.ID
		}
	}
	if otherID == "" {
		t.Fatal("expected other session ID")
	}

	// Self-revocation, missing ID, and wrong password are rejected.
	if err := userSvc.RevokeSession(ctx, sess1.AccessToken, "", in.Password); codeOf(t, err) != string(exception.MsgInvalidInput) {
		t.Fatalf("expected invalid input for empty id, got %v", err)
	}
	if err := userSvc.RevokeSession(ctx, sess1.AccessToken, currentID, in.Password); codeOf(t, err) != string(exception.MsgInvalidInput) {
		t.Fatalf("expected invalid input for current session, got %v", err)
	}
	if err := userSvc.RevokeSession(ctx, sess1.AccessToken, otherID, "WrongPassword123!"); codeOf(t, err) != string(exception.MsgInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
	if err := userSvc.RevokeSession(ctx, sess1.AccessToken, "nonexistent", in.Password); codeOf(t, err) != string(exception.MsgInvalidInput) {
		t.Fatalf("expected not found/invalid input, got %v", err)
	}

	// Valid single revocation removes sess2 and revokes its refresh token.
	if err := userSvc.RevokeSession(ctx, sess1.AccessToken, otherID, in.Password); err != nil {
		t.Fatalf("RevokeSession: %v", err)
	}
	if _, err := authSvc.Refresh(ctx, sess2.RefreshToken, models.ClientMeta{}); codeOf(t, err) != string(exception.MsgUnauthorized) {
		t.Fatalf("expected revoked refresh for sess2, got %v", err)
	}
	listedAfterSingle, _, err := userSvc.ListSessions(ctx, sess1.AccessToken)
	if err != nil || len(listedAfterSingle) != 1 || listedAfterSingle[0].ID != currentID {
		t.Fatalf("expected only current session left, got %+v %v", listedAfterSingle, err)
	}

	// Add a third session and revoke all others.
	_ = st.InsertSessionRecord(models.Session{
		ID:         "extra-sess-3",
		UserID:     sess1.User.ID,
		ExpiresAt:  time.Now().UTC().Add(24 * time.Hour),
		LastUsedAt: time.Now().UTC(),
		CreatedAt:  time.Now().UTC(),
	})
	if err := userSvc.RevokeOtherSessions(ctx, sess1.AccessToken, "WrongPassword123!"); codeOf(t, err) != string(exception.MsgInvalidCredentials) {
		t.Fatalf("expected invalid credentials on RevokeOtherSessions, got %v", err)
	}
	if err := userSvc.RevokeOtherSessions(ctx, sess1.AccessToken, in.Password); err != nil {
		t.Fatalf("RevokeOtherSessions: %v", err)
	}
	listedFinal, _, err := userSvc.ListSessions(ctx, sess1.AccessToken)
	if err != nil || len(listedFinal) != 1 || listedFinal[0].ID != currentID {
		t.Fatalf("expected only current session after RevokeOtherSessions, got %+v %v", listedFinal, err)
	}
}

func TestUserActivityRecording(t *testing.T) {
	ctx := context.Background()
	st := usermock.NewState()
	actState := activitymock.NewState()
	actRepo := activitymock.NewPgRepositoryMock(actState)
	authSvc, notify, users, _, _, _ := testSupport.NewAuthService(t, st, nil)
	_, _, sessions := testSupport.TestRepos(st)

	log := logger.New(logger.Options{Service: "user-test", Env: "test", Level: "error"})
	actSvc := activityservice.NewService(actRepo, sessions, []byte(testSupport.TestJWTSecret), log)

	sess := testSupport.RegisterSession(t, authSvc, notify, registerInput("activity@example.com"))
	svc := userservice.NewService(
		users,
		groupmock.NewPgRepositoryMock(st),
		groupmembermock.NewPgRepositoryMock(st),
		nil,
		sessions,
		nil,
		nil,
		nil,
		[]byte(testSupport.TestJWTSecret),
		log,
	).WithActivityLog(actSvc)

	if _, err := svc.UpdateUserProfile(ctx, sess.AccessToken, "Ada L", "she/her", "Builder", "#123456", false); err != nil {
		t.Fatalf("UpdateUserProfile: %v", err)
	}
	if _, err := svc.UpdateUserStatusNote(ctx, sess.AccessToken, "Focus", "🎯", ""); err != nil {
		t.Fatalf("UpdateUserStatusNote set: %v", err)
	}
	if _, err := svc.UpdateUserStatusNote(ctx, sess.AccessToken, "", "", ""); err != nil {
		t.Fatalf("UpdateUserStatusNote clear: %v", err)
	}
	if _, err := svc.UpdateUserPreferences(ctx, sess.AccessToken, "id", "24h"); err != nil {
		t.Fatalf("UpdateUserPreferences: %v", err)
	}
	if _, err := svc.UpdateUsername(ctx, sess.AccessToken, "ada_updated", testPassword); err != nil {
		t.Fatalf("UpdateUsername: %v", err)
	}
	if err := svc.UpdatePassword(ctx, sess.AccessToken, testPassword, "new-correct-horse-battery"); err != nil {
		t.Fatalf("UpdatePassword: %v", err)
	}

	page, err := actSvc.ListMyActivityLogs(ctx, sess.AccessToken, "", "", 20, "")
	if err != nil {
		t.Fatalf("ListMyActivityLogs: %v", err)
	}
	gotActions := make(map[string]bool)
	for _, entry := range page.Logs {
		gotActions[entry.Action] = true
	}
	for _, want := range []string{
		constants.ActivityUserProfileUpdated,
		constants.ActivityUserStatusNoteUpdated,
		constants.ActivityUserStatusNoteCleared,
		constants.ActivityUserPreferencesUpdated,
		constants.ActivityUserUsernameUpdated,
		constants.ActivityUserPasswordUpdated,
	} {
		if !gotActions[want] {
			t.Fatalf("expected recorded action %q, got %v", want, gotActions)
		}
	}
}

func TestUserRelationships(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, _, svc, _, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	alice := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "alice-rel@example.com", Username: "alice.rel", Password: testPassword, DisplayName: "Alice", BirthDate: "2000-01-01",
	})
	bob := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email: "bob-rel@example.com", Username: "bob.rel", Password: testPassword, DisplayName: "Bob", BirthDate: "2000-01-01",
	})

	if _, err := svc.SetRelationship(ctx, alice.AccessToken, alice.User.ID, "ignore"); err == nil {
		t.Fatal("expected self relationship to fail")
	}
	if _, err := svc.SetRelationship(ctx, alice.AccessToken, bob.User.ID, "invalid"); err == nil {
		t.Fatal("expected invalid kind to fail")
	}

	kind, err := svc.SetRelationship(ctx, alice.AccessToken, bob.User.ID, "ignore")
	if err != nil || kind != "ignore" {
		t.Fatalf("SetRelationship ignore: %q %v", kind, err)
	}
	rels, err := svc.ListRelationships(ctx, alice.AccessToken)
	if err != nil || len(rels) != 1 || rels[0].TargetUserID != bob.User.ID || rels[0].Kind != "ignore" {
		t.Fatalf("ListRelationships: %+v %v", rels, err)
	}

	kind, err = svc.SetRelationship(ctx, alice.AccessToken, bob.User.ID, "block")
	if err != nil || kind != "block" {
		t.Fatalf("SetRelationship block: %q %v", kind, err)
	}
	rels, err = svc.ListRelationships(ctx, alice.AccessToken)
	if err != nil || len(rels) != 1 || rels[0].Kind != "block" {
		t.Fatalf("ListRelationships block: %+v %v", rels, err)
	}

	kind, err = svc.SetRelationship(ctx, alice.AccessToken, bob.User.ID, "none")
	if err != nil || kind != "none" {
		t.Fatalf("SetRelationship none: %q %v", kind, err)
	}
	rels, err = svc.ListRelationships(ctx, alice.AccessToken)
	if err != nil || len(rels) != 0 {
		t.Fatalf("ListRelationships after clear: %+v %v", rels, err)
	}
}
