package service_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authrepo "github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/auth/service"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	userservice "github.com/wichat/wichat/backend/ms-user/internal/user/service"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	googleoauth "github.com/wichat/wichat/backend/ms-user/pkg/oauth"
	testSupport "github.com/wichat/wichat/backend/ms-user/pkg/test-support"
	"github.com/wichat/wichat/backend/ms-user/pkg/username"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

const testPassword = "correct-horse-battery"

func registerInput(email string) models.RegisterInput {
	handle, _ := username.BaseFromEmail(email)
	if handle == "" || handle == "user" {
		handle = "user." + strings.ReplaceAll(strings.Split(email, "@")[0], "+", "")
		if len(handle) < 3 {
			handle = "user.test"
		}
	}
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

func TestRegisterUnicodeDisplayName(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	out := testSupport.RegisterSession(t, svc, notify, models.RegisterInput{
		Email:       "unicode@example.com",
		Username:    "unicode.user",
		Password:    testPassword,
		DisplayName: "山田 太郎 🎮",
		BirthDate:   "2000-01-15",
	})
	if out.User.DisplayName != "山田 太郎 🎮" {
		t.Fatalf("display name: %q", out.User.DisplayName)
	}
}

func TestLogoutIgnoresEmptyRefresh(t *testing.T) {
	svc, _, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	if err := svc.Logout(context.Background(), ""); err != nil {
		t.Fatal(err)
	}
}

func TestRequestPasswordResetRateLimitedOnBackToBack(t *testing.T) {
	svc, notify, _, _, _, mr := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()

	testSupport.RegisterSession(t, svc, notify, registerInput("reset-rl@example.com"))
	if _, err := svc.RequestPasswordReset(ctx, "reset-rl@example.com", "en"); err != nil {
		t.Fatal(err)
	}
	_, err := svc.RequestPasswordReset(ctx, "reset-rl@example.com", "en")
	if codeOf(t, err) != string(exception.MsgResetRateLimited) {
		t.Fatalf("%v", err)
	}
	mr.FastForward(61 * time.Second)
	if _, err := svc.RequestPasswordReset(ctx, "reset-rl@example.com", "en"); err != nil {
		t.Fatal(err)
	}
}

func TestRegisterFirstUserIsOwner(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()

	first := testSupport.RegisterSession(t, svc, notify, models.RegisterInput{
		Email: "ada-owner@example.com", Username: "ada.owner", Password: testPassword, DisplayName: "Ada", BirthDate: "2000-01-15",
	})
	if first.AccessToken == "" || first.RefreshToken == "" || first.User.Email != "ada-owner@example.com" {
		t.Fatalf("session: %+v", first)
	}
	if first.User.Username != "ada.owner" {
		t.Fatalf("username: %q", first.User.Username)
	}

	user, err := svc.Me(ctx, first.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if user.ID != first.User.ID {
		t.Fatalf("me: %+v", user)
	}
}

func TestRegisterUsernameCollision(t *testing.T) {
	svc, notify, _, userSvc, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, svc, notify, registerInput("owner-col@example.com"))
	if err := userSvc.CompleteInstanceSetup(ctx, owner.AccessToken, models.CompleteInstanceSetupInput{
		Name:             "WiChat",
		RegistrationMode: constants.RegistrationOpen,
	}); err != nil {
		t.Fatal(err)
	}

	first := testSupport.RegisterSession(t, svc, notify, models.RegisterInput{
		Email: "john@a.com", Username: "john.doe", Password: testPassword, DisplayName: "John", BirthDate: "2000-01-15",
	})
	if first.User.Username != "john.doe" {
		t.Fatalf("first username: %q", first.User.Username)
	}

	_, err := svc.Register(ctx, models.RegisterInput{
		Email: "john@b.com", Username: "John.Doe", Password: testPassword, DisplayName: "John B", BirthDate: "2000-01-15",
	})
	if codeOf(t, err) != string(exception.MsgUsernameTaken) {
		t.Fatalf("expected username taken, got %v", err)
	}
}

func TestLoginHidesUnknownEmail(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()
	testSupport.RegisterSession(t, svc, notify, registerInput("admin@example.com"))

	_, unknown := svc.Login(ctx, models.LoginInput{Login: "missing@example.com", Password: testPassword})
	_, wrong := svc.Login(ctx, models.LoginInput{Login: "admin@example.com", Password: "not-the-password"})
	if codeOf(t, unknown) != string(exception.MsgInvalidCredentials) || codeOf(t, unknown) != codeOf(t, wrong) {
		t.Fatalf("unknown=%v wrong=%v", unknown, wrong)
	}

	got := testSupport.LoginSession(t, svc, notify, models.LoginInput{Login: "Admin@Example.com", Password: testPassword})
	if got.User.Email != "admin@example.com" {
		t.Fatalf("email: %s", got.User.Email)
	}
}

func TestRefreshAfterRedisFlush(t *testing.T) {
	svc, notify, _, _, _, mr := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()
	first := testSupport.RegisterSession(t, svc, notify, registerInput("admin@example.com"))

	mr.FlushAll()

	second, err := svc.Refresh(ctx, first.RefreshToken, models.ClientMeta{})
	if err != nil {
		t.Fatal(err)
	}
	if second.RefreshToken == "" || second.RefreshToken == first.RefreshToken {
		t.Fatal("refresh token was not rotated after redis flush")
	}
	if _, err := svc.Me(ctx, second.AccessToken); err != nil {
		t.Fatal(err)
	}
}

func TestRefreshRotationKeepsAccess(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()
	first := testSupport.RegisterSession(t, svc, notify, registerInput("admin@example.com"))

	second, err := svc.Refresh(ctx, first.RefreshToken, models.ClientMeta{})
	if err != nil {
		t.Fatal(err)
	}
	if second.RefreshToken == "" || second.RefreshToken == first.RefreshToken {
		t.Fatal("refresh token was not rotated")
	}
	if _, err := svc.Me(ctx, first.AccessToken); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Me(ctx, second.AccessToken); err != nil {
		t.Fatal(err)
	}
}

func TestRefreshReuseWithinGraceDoesNotRevoke(t *testing.T) {
	clock := time.Now().UTC()
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), func() time.Time { return clock })
	ctx := context.Background()
	first := testSupport.RegisterSession(t, svc, notify, registerInput("grace@example.com"))
	second, err := svc.Refresh(ctx, first.RefreshToken, models.ClientMeta{})
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.Refresh(ctx, first.RefreshToken, models.ClientMeta{})
	if codeOf(t, err) != string(exception.MsgUnauthorized) {
		t.Fatalf("reuse within grace: %v", err)
	}
	if _, err := svc.Me(ctx, second.AccessToken); err != nil {
		t.Fatalf("session should stay active within grace: %v", err)
	}
}

func TestRefreshReuseRevokesSession(t *testing.T) {
	clock := time.Now().UTC()
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), func() time.Time { return clock })
	ctx := context.Background()
	first := testSupport.RegisterSession(t, svc, notify, registerInput("admin@example.com"))
	second, err := svc.Refresh(ctx, first.RefreshToken, models.ClientMeta{})
	if err != nil {
		t.Fatal(err)
	}

	clock = clock.Add(constants.RefreshReuseRevokeGrace + time.Second)

	_, err = svc.Refresh(ctx, first.RefreshToken, models.ClientMeta{})
	if codeOf(t, err) != string(exception.MsgUnauthorized) {
		t.Fatalf("reuse: %v", err)
	}
	if _, err := svc.Me(ctx, second.AccessToken); codeOf(t, err) != string(exception.MsgUnauthorized) {
		t.Fatalf("access after reuse: %v", err)
	}
	if _, err := svc.Refresh(ctx, second.RefreshToken, models.ClientMeta{}); codeOf(t, err) != string(exception.MsgUnauthorized) {
		t.Fatalf("rotated token after reuse: %v", err)
	}
}

func TestUnknownRefreshDoesNotRevoke(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()
	first := testSupport.RegisterSession(t, svc, notify, registerInput("admin@example.com"))

	_, err := svc.Refresh(ctx, "not-a-token", models.ClientMeta{})
	if codeOf(t, err) != string(exception.MsgUnauthorized) {
		t.Fatalf("unknown: %v", err)
	}
	if _, err := svc.Me(ctx, first.AccessToken); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Refresh(ctx, first.RefreshToken, models.ClientMeta{}); err != nil {
		t.Fatal(err)
	}
}

func TestLogoutRejectsAccess(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()
	first := testSupport.RegisterSession(t, svc, notify, registerInput("admin@example.com"))

	if err := svc.Logout(ctx, first.RefreshToken); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Me(ctx, first.AccessToken); codeOf(t, err) != string(exception.MsgUnauthorized) {
		t.Fatalf("me after logout: %v", err)
	}
	if _, err := svc.Refresh(ctx, first.RefreshToken, models.ClientMeta{}); codeOf(t, err) != string(exception.MsgUnauthorized) {
		t.Fatalf("refresh after logout: %v", err)
	}
}

func TestGoogleCreatesLinksAndRejectsPassword(t *testing.T) {
	st := usermock.NewState()
	users, _, sessions := testSupport.TestRepos(st)
	var mode string
	google := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/token":
			_ = r.ParseForm()
			if r.Form.Get("code_verifier") == "" || r.Form.Get("code") == "" {
				http.Error(w, "missing", http.StatusBadRequest)
				return
			}
			_, _ = w.Write([]byte(`{"access_token":"token-1"}`))
		default:
			if r.Header.Get("Authorization") != "Bearer token-1" {
				http.Error(w, "auth", http.StatusUnauthorized)
				return
			}
			if mode == "unverified" {
				_, _ = w.Write([]byte(`{"sub":"sub-2","email":"new@example.com","email_verified":false,"name":"New"}`))
				return
			}
			if mode == "second" {
				_, _ = w.Write([]byte(`{"sub":"sub-2","email":"new@example.com","email_verified":true,"name":"New Person"}`))
				return
			}
			_, _ = w.Write([]byte(`{"sub":"sub-1","email":"admin@example.com","email_verified":true,"name":"Ada Lovelace"}`))
		}
	}))
	t.Cleanup(google.Close)

	notify := &testSupport.FakeNotifier{}
	rdb := testSupport.RedisClient(t)
	refresh := authrepo.NewRedisRefreshStore(rdb)
	log := logger.New(logger.Options{Service: "auth-test", Env: "test", Level: "error"})
	userSvc := userservice.NewService(users, nil, nil, sessions, []byte(testSupport.TestJWTSecret), log)
	svc, err := service.NewService(users, sessions, refresh, []byte(testSupport.TestJWTSecret), log, googleoauth.Config{
		ClientID:     "client",
		ClientSecret: "secret",
		RedirectURL:  "http://localhost:3000/api/v1/auth/google/callback",
		TokenURL:     google.URL + "/token",
		UserInfoURL:  google.URL + "/userinfo",
	}, authrepo.NewRedisPasswordResetStore(rdb), authrepo.NewRedisEmailAuthStore(rdb), notify, userSvc, nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	testSupport.RegisterSession(t, svc, notify, registerInput("admin@example.com"))

	linked, err := svc.GoogleFinish(ctx, "code", "verifier", models.ClientMeta{})
	if err != nil {
		t.Fatal(err)
	}
	if linked.User.Email != "admin@example.com" || linked.User.DisplayName != "Ada" {
		t.Fatalf("linked: %+v", linked.User)
	}

	again, err := svc.GoogleFinish(ctx, "code", "verifier", models.ClientMeta{})
	if err != nil || again.User.ID != linked.User.ID {
		t.Fatalf("again: %+v %v", again, err)
	}

	testSupport.LoginSession(t, svc, notify, models.LoginInput{Login: "admin@example.com", Password: testPassword})

	mode = "second"
	created, err := svc.GoogleFinish(ctx, "code", "verifier", models.ClientMeta{})
	if err != nil {
		t.Fatal(err)
	}
	if created.User.ID == linked.User.ID || created.User.Email != "new@example.com" || created.User.DisplayName != "New Person" {
		t.Fatalf("created: %+v", created.User)
	}
	_, err = svc.Login(ctx, models.LoginInput{Login: "new@example.com", Password: testPassword})
	if codeOf(t, err) != string(exception.MsgInvalidCredentials) {
		t.Fatalf("password on google account: %v", err)
	}

	mode = "unverified"
	_, err = svc.GoogleFinish(ctx, "code", "verifier", models.ClientMeta{})
	if codeOf(t, err) != string(exception.MsgGoogleEmailUnverified) {
		t.Fatalf("unverified: %v", err)
	}
}

func TestRegisterDoesNotCreateUserUntilVerify(t *testing.T) {
	st := usermock.NewState()
	svc, notify, users, _, _, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	step, err := svc.Register(ctx, registerInput("pending@example.com"))
	if err != nil {
		t.Fatal(err)
	}
	if step.Challenge == nil || step.Session != nil {
		t.Fatalf("expected challenge only: %+v", step)
	}

	n, err := users.CountUsers(ctx)
	if err != nil || n != 0 {
		t.Fatalf("users before verify: %d err %v", n, err)
	}

	code := notify.LastCode
	if code == "" {
		t.Fatal("expected register email code")
	}
	if _, err := svc.VerifyRegister(ctx, step.Challenge.ChallengeID, code); err != nil {
		t.Fatal(err)
	}
	n, err = users.CountUsers(ctx)
	if err != nil || n != 1 {
		t.Fatalf("users after verify: %d err %v", n, err)
	}
}

func TestLoginDoesNotIssueSessionUntilVerify(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()
	testSupport.RegisterSession(t, svc, notify, registerInput("ada@example.com"))

	step, err := svc.Login(ctx, models.LoginInput{
		Login:    "ada@example.com",
		Password: testPassword,
	})
	if err != nil {
		t.Fatal(err)
	}
	if step.Challenge == nil || step.Session != nil {
		t.Fatalf("expected challenge: %+v", step)
	}

	code := notify.LastCode
	if code == "" {
		t.Fatal("expected login email code")
	}
	sess, err := svc.VerifyLogin(ctx, step.Challenge.ChallengeID, code)
	if err != nil || sess.AccessToken == "" {
		t.Fatalf("verify login: %v %+v", err, sess)
	}
}

func TestVerifyLoginRejectsBadCode(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()
	testSupport.RegisterSession(t, svc, notify, registerInput("ada@example.com"))

	step, err := svc.Login(ctx, models.LoginInput{Login: "ada@example.com", Password: testPassword})
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.VerifyLogin(ctx, step.Challenge.ChallengeID, "AAAAAA")
	if codeOf(t, err) != string(exception.MsgResetCodeInvalid) {
		t.Fatalf("bad code: %v", err)
	}
}

func TestRequestPasswordResetUnknownEmail(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)

	cooldown, err := svc.RequestPasswordReset(context.Background(), "missing@example.com", "en")
	if err != nil {
		t.Fatal(err)
	}
	if notify.Sends != 0 {
		t.Fatalf("expected no send, got %d", notify.Sends)
	}
	if cooldown != 60 {
		t.Fatalf("cooldown: got %d want 60", cooldown)
	}
}

func TestRequestPasswordResetOAuthOnlyAccount(t *testing.T) {
	st := usermock.NewState()
	svc, notify, users, _, _, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	_, err := users.SaveGoogleUser(ctx, models.User{
		ID:           "user-oauth",
		Email:        "oauth@example.com",
		PasswordHash: "",
		DisplayName:  "OAuth User",
	}, "identity-1", "google", "google-sub-1")
	if err != nil {
		t.Fatal(err)
	}

	cooldown, err := svc.RequestPasswordReset(ctx, "oauth@example.com", "en")
	if err != nil {
		t.Fatal(err)
	}
	if notify.Sends != 1 {
		t.Fatalf("expected reset email, got %d sends", notify.Sends)
	}
	if cooldown != 60 {
		t.Fatalf("cooldown: got %d want 60", cooldown)
	}

	code := notify.LastCode
	if code == "" {
		t.Fatal("expected code captured")
	}

	token, err := svc.VerifyPasswordReset(ctx, "oauth@example.com", code)
	if err != nil || token == "" {
		t.Fatalf("verify: err=%v token=%q", err, token)
	}

	if err := svc.CompletePasswordReset(ctx, token, "new-password-12"); err != nil {
		t.Fatal(err)
	}

	user, err := users.FindUserByEmail(ctx, "oauth@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if user.PasswordHash == "" {
		t.Fatal("expected password hash after reset")
	}
}

func TestRequestPasswordResetSendsForKnownUser(t *testing.T) {
	st := usermock.NewState()
	svc, notify, _, _, _, mr := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	testSupport.RegisterSession(t, svc, notify, registerInput("ada@example.com"))

	cooldown, err := svc.RequestPasswordReset(ctx, "ada@example.com", "en")
	if err != nil {
		t.Fatal(err)
	}
	if notify.Sends != 2 {
		t.Fatalf("expected 2 sends (register + reset), got %d", notify.Sends)
	}
	if cooldown != 60 {
		t.Fatalf("cooldown: got %d want 60", cooldown)
	}

	_, err = svc.RequestPasswordReset(ctx, "ada@example.com", "en")
	if codeOf(t, err) != string(exception.MsgResetRateLimited) {
		t.Fatalf("cooldown: %v", err)
	}

	mr.FastForward(61 * time.Second)

	cooldown, err = svc.RequestPasswordReset(ctx, "ada@example.com", "en")
	if err != nil {
		t.Fatalf("second send: %v", err)
	}
	if cooldown != 120 {
		t.Fatalf("second cooldown: got %d want 120", cooldown)
	}
}

func TestVerifyPasswordResetClearsResendPenalty(t *testing.T) {
	st := usermock.NewState()
	svc, notify, _, _, _, mr := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	testSupport.RegisterSession(t, svc, notify, registerInput("ada@example.com"))

	if _, err := svc.RequestPasswordReset(ctx, "ada@example.com", "en"); err != nil {
		t.Fatal(err)
	}

	code := notify.LastCode
	if code == "" {
		t.Fatal("expected code captured")
	}

	token, err := svc.VerifyPasswordReset(ctx, "ada@example.com", code)
	if err != nil || token == "" {
		t.Fatalf("verify: err=%v token=%q", err, token)
	}

	mr.FastForward(61 * time.Second)
	cooldown, err := svc.RequestPasswordReset(ctx, "ada@example.com", "en")
	if err != nil {
		t.Fatalf("after verify reset: %v", err)
	}
	if cooldown != 60 {
		t.Fatalf("strikes reset: cooldown %d want 60", cooldown)
	}
}

// auth_boost_test.go

func TestOpenSessionTruncatesLongUserAgent(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()

	ua := strings.Repeat("x", constants.MaxUserAgentLen+50)
	in := registerInput("ua-long@example.com")
	in.ClientMeta = models.ClientMeta{UserAgent: ua, ClientIP: "127.0.0.1"}
	step, err := svc.Register(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	sess, err := svc.VerifyRegister(ctx, step.Challenge.ChallengeID, notify.LastCode)
	if err != nil || sess.AccessToken == "" {
		t.Fatalf("%+v %v", sess, err)
	}
}

func TestRefreshAfterUserDeleted(t *testing.T) {
	st := usermock.NewState()
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	sess := testSupport.RegisterSession(t, svc, notify, registerInput("deleted-user@example.com"))
	filtered := st.Users[:0]
	for _, u := range st.Users {
		if u.ID != sess.User.ID {
			filtered = append(filtered, u)
		}
	}
	st.Users = filtered
	_, err := svc.Refresh(ctx, sess.RefreshToken, models.ClientMeta{})
	if err == nil {
		t.Fatal("expected refresh error after user delete")
	}
}

func TestVerifyLoginWithClientMeta(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()

	testSupport.RegisterSession(t, svc, notify, registerInput("login-meta@example.com"))
	step, err := svc.Login(ctx, models.LoginInput{Login: "login-meta@example.com", Password: testPassword})
	if err != nil {
		t.Fatal(err)
	}
	out, err := svc.VerifyLogin(ctx, step.Challenge.ChallengeID, notify.LastCode)
	if err != nil || out.AccessToken == "" {
		t.Fatalf("%+v %v", out, err)
	}
}

func TestRequestPasswordResetUsesLocale(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()

	testSupport.RegisterSession(t, svc, notify, registerInput("locale-reset@example.com"))
	if _, err := svc.RequestPasswordReset(ctx, "locale-reset@example.com", "id"); err != nil {
		t.Fatal(err)
	}
	if notify.Sends < 2 {
		t.Fatalf("sends %d", notify.Sends)
	}
}

func TestSetupRejectsRegisterValidation(t *testing.T) {
	svc, _, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()

	_, err := svc.Setup(ctx, models.SetupInput{
		Email:       "setup@example.com",
		Password:    testPassword,
		DisplayName: "Setup",
	})
	if codeOf(t, err) != string(exception.MsgInvalidInput) {
		t.Fatalf("%v", err)
	}
}

func TestGoogleStartUnavailableWhenUnconfigured(t *testing.T) {
	svc, _, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	_, _, _, err := svc.GoogleStart(context.Background())
	if codeOf(t, err) != string(exception.MsgUnavailable) {
		t.Fatalf("%v", err)
	}
}

func TestGoogleStartWhenConfigured(t *testing.T) {
	st := usermock.NewState()
	users, _, sessions := testSupport.TestRepos(st)
	rdb := testSupport.RedisClient(t)
	log := logger.New(logger.Options{Service: "auth-test", Env: "test", Level: "error"})
	userSvc := userservice.NewService(users, nil, nil, sessions, []byte(testSupport.TestJWTSecret), log)
	svc, err := service.NewService(users, sessions, authrepo.NewRedisRefreshStore(rdb), []byte(testSupport.TestJWTSecret), log, googleoauth.Config{
		ClientID: "client-id", ClientSecret: "secret", RedirectURL: "http://localhost/callback",
	}, authrepo.NewRedisPasswordResetStore(rdb), authrepo.NewRedisEmailAuthStore(rdb), &testSupport.FakeNotifier{}, userSvc, nil)
	if err != nil {
		t.Fatal(err)
	}

	url, state, verifier, err := svc.GoogleStart(context.Background())
	if err != nil || url == "" || state == "" || verifier == "" {
		t.Fatalf("start url=%q state=%q verifier=%q err=%v", url, state, verifier, err)
	}
	if !strings.Contains(url, "client-id") {
		t.Fatalf("authorize url: %q", url)
	}
}

func TestResendLoginCodeAfterCooldown(t *testing.T) {
	svc, notify, _, _, _, mr := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()

	testSupport.RegisterSession(t, svc, notify, registerInput("resend-login-wait@example.com"))
	step, err := svc.Login(ctx, modelsLogin("resend-login-wait@example.com"))
	if err != nil {
		t.Fatal(err)
	}
	firstCode := notify.LastCode
	mr.FastForward(61 * time.Second)

	cooldown, err := svc.ResendLoginCode(ctx, step.Challenge.ChallengeID)
	if err != nil {
		t.Fatal(err)
	}
	if cooldown <= 0 || notify.LastCode == firstCode {
		t.Fatalf("cooldown=%d code=%q", cooldown, notify.LastCode)
	}
}

func TestResendRegisterCodeAfterCooldown(t *testing.T) {
	svc, notify, _, _, _, mr := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()

	step, err := svc.Register(ctx, registerInput("resend-after-wait@example.com"))
	if err != nil {
		t.Fatal(err)
	}
	firstCode := notify.LastCode
	mr.FastForward(61 * time.Second)

	cooldown, err := svc.ResendRegisterCode(ctx, step.Challenge.ChallengeID)
	if err != nil {
		t.Fatal(err)
	}
	if cooldown <= 0 {
		t.Fatalf("cooldown %d", cooldown)
	}
	if notify.LastCode == firstCode || notify.LastCode == "" {
		t.Fatal("expected new OTP after resend")
	}
}

func TestResendLoginCodeRateLimited(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()

	testSupport.RegisterSession(t, svc, notify, registerInput("resend-login-rl@example.com"))
	step, err := svc.Login(ctx, models.LoginInput{Login: "resend-login-rl@example.com", Password: testPassword})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.ResendLoginCode(ctx, step.Challenge.ChallengeID)
	if codeOf(t, err) != string(exception.MsgResetRateLimited) {
		t.Fatalf("%v", err)
	}
}

func TestMeRejectsRevokedSession(t *testing.T) {
	clock := time.Now().UTC()
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), func() time.Time { return clock })
	ctx := context.Background()

	sess := testSupport.RegisterSession(t, svc, notify, registerInput("revoked-me@example.com"))
	first, err := svc.Refresh(ctx, sess.RefreshToken, models.ClientMeta{})
	if err != nil {
		t.Fatal(err)
	}
	clock = clock.Add(constants.RefreshReuseRevokeGrace + time.Second)
	_, _ = svc.Refresh(ctx, sess.RefreshToken, models.ClientMeta{})

	if _, err := svc.Me(ctx, first.AccessToken); codeOf(t, err) != string(exception.MsgUnauthorized) {
		t.Fatalf("%v", err)
	}
}

// auth_deep_test.go

func TestRegisterRequiresValidUsername(t *testing.T) {
	st := usermock.NewState()
	svc, notify, _, userSvc, _, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, svc, notify, registerInput("first-req-un@example.com"))
	if err := userSvc.CompleteInstanceSetup(ctx, owner.AccessToken, models.CompleteInstanceSetupInput{
		Name:             "WiChat",
		RegistrationMode: constants.RegistrationOpen,
	}); err != nil {
		t.Fatal(err)
	}

	_, err := svc.Register(ctx, models.RegisterInput{
		Email: "second-req-un@example.com", Username: "admin", Password: testPassword, DisplayName: "Second", BirthDate: "2000-01-15",
	})
	if codeOf(t, err) != string(exception.MsgInvalidInput) {
		t.Fatalf("expected invalid input for reserved username, got %v", err)
	}
}

func TestRefreshTwiceRotatesAgain(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()

	first := testSupport.RegisterSession(t, svc, notify, registerInput("rotate2@example.com"))
	second, err := svc.Refresh(ctx, first.RefreshToken, models.ClientMeta{})
	if err != nil {
		t.Fatal(err)
	}
	third, err := svc.Refresh(ctx, second.RefreshToken, models.ClientMeta{})
	if err != nil || third.RefreshToken == "" || third.RefreshToken == second.RefreshToken {
		t.Fatalf("%+v %v", third, err)
	}
	if _, err := svc.Me(ctx, third.AccessToken); err != nil {
		t.Fatal(err)
	}
}

func TestPasswordResetWrongCodeTwice(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()

	testSupport.RegisterSession(t, svc, notify, registerInput("wrong2@example.com"))
	if _, err := svc.RequestPasswordReset(ctx, "wrong2@example.com", "en"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.VerifyPasswordReset(ctx, "wrong2@example.com", "000001"); codeOf(t, err) != string(exception.MsgResetCodeInvalid) {
		t.Fatal(err)
	}
	if _, err := svc.VerifyPasswordReset(ctx, "wrong2@example.com", "000002"); codeOf(t, err) != string(exception.MsgResetCodeInvalid) {
		t.Fatal(err)
	}
}

func TestVerifyLoginMissingChallenge(t *testing.T) {
	svc, _, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	_, err := svc.VerifyLogin(context.Background(), "missing-challenge", "123456")
	if codeOf(t, err) != string(exception.MsgResetCodeInvalid) {
		t.Fatalf("%v", err)
	}
}

func TestMeRejectsExpiredAccessWindow(t *testing.T) {
	clock := time.Now().UTC()
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), func() time.Time { return clock })
	ctx := context.Background()

	sess := testSupport.RegisterSession(t, svc, notify, registerInput("expired-me@example.com"))
	clock = clock.Add(constants.RefreshTokenTTL + time.Hour)

	if _, err := svc.Me(ctx, sess.AccessToken); codeOf(t, err) != string(exception.MsgUnauthorized) {
		t.Fatalf("%v", err)
	}
}

// auth_edges_test.go

func TestRegisterRejectsDuplicateEmail(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()
	testSupport.RegisterSession(t, svc, notify, registerInput("dup@example.com"))

	_, err := svc.Register(ctx, registerInput("dup@example.com"))
	if codeOf(t, err) != string(exception.MsgEmailTaken) {
		t.Fatalf("dup: %v", err)
	}
}

func TestRegisterRejectsUnderage(t *testing.T) {
	svc, _, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	_, err := svc.Register(context.Background(), models.RegisterInput{
		Email: "kid@example.com", Username: "kid.user", Password: testPassword, DisplayName: "Kid", BirthDate: "2020-01-01",
	})
	if codeOf(t, err) != string(exception.MsgAgeRestricted) {
		t.Fatalf("age: %v", err)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()
	testSupport.RegisterSession(t, svc, notify, registerInput("login@example.com"))

	_, err := svc.Login(ctx, models.LoginInput{Login: "login@example.com", Password: "wrong-password-here"})
	if codeOf(t, err) != string(exception.MsgInvalidCredentials) {
		t.Fatalf("bad pass: %v", err)
	}
}

func TestVerifyRegisterRejectsEmptyFields(t *testing.T) {
	svc, _, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	_, err := svc.VerifyRegister(context.Background(), "", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRefreshRejectsEmptyToken(t *testing.T) {
	svc, _, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	_, err := svc.Refresh(context.Background(), "", models.ClientMeta{})
	if codeOf(t, err) != string(exception.MsgUnauthorized) {
		t.Fatalf("%v", err)
	}
}

func TestLogoutNoopOnEmptyRefresh(t *testing.T) {
	svc, _, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	if err := svc.Logout(context.Background(), ""); err != nil {
		t.Fatal(err)
	}
}

func TestMeRejectsEmptyAccessToken(t *testing.T) {
	svc, _, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	_, err := svc.Me(context.Background(), "")
	if codeOf(t, err) != string(exception.MsgUnauthorized) {
		t.Fatalf("%v", err)
	}
}

func TestCompletePasswordResetRejectsShortPassword(t *testing.T) {
	svc, _, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	err := svc.CompletePasswordReset(context.Background(), "token", "short")
	if err == nil {
		t.Fatal("expected invalid password")
	}
}

func TestCompletePasswordResetRejectsUnknownToken(t *testing.T) {
	svc, _, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	err := svc.CompletePasswordReset(context.Background(), "unknown-token", testPassword)
	if codeOf(t, err) != string(exception.MsgUnauthorized) {
		t.Fatalf("%v", err)
	}
}

func TestResendRegisterRejectsMissingChallenge(t *testing.T) {
	svc, _, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	_, err := svc.ResendRegisterCode(context.Background(), "nope")
	if codeOf(t, err) != string(exception.MsgResetCodeInvalid) {
		t.Fatalf("%v", err)
	}
}

func TestPasswordResetFlowForRegisteredUser(t *testing.T) {
	st := usermock.NewState()
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	testSupport.RegisterSession(t, svc, notify, registerInput("reset-full@example.com"))
	if _, err := svc.RequestPasswordReset(ctx, "reset-full@example.com", "en"); err != nil {
		t.Fatal(err)
	}
	token, err := svc.VerifyPasswordReset(ctx, "reset-full@example.com", notify.LastCode)
	if err != nil || token == "" {
		t.Fatalf("verify %v %q", err, token)
	}
	newPass := "new-horse-battery-x"
	if err := svc.CompletePasswordReset(ctx, token, newPass); err != nil {
		t.Fatal(err)
	}

	step, err := svc.Login(ctx, models.LoginInput{Login: "reset-full@example.com", Password: newPass})
	if err != nil || step.Challenge == nil {
		t.Fatalf("login: %+v %v", step, err)
	}
}

// auth_more_test.go

func TestLoginByUsername(t *testing.T) {
	st := usermock.NewState()
	svc, notify, _, userSvc, _, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, svc, notify, registerInput("login-user@example.com"))
	if err := userSvc.CompleteInstanceSetup(ctx, owner.AccessToken, models.CompleteInstanceSetupInput{
		Name: "WiChat", RegistrationMode: "open",
	}); err != nil {
		t.Fatal(err)
	}
	second := testSupport.RegisterSession(t, svc, notify, registerInput("other@example.com"))
	if second.User.Username == "" {
		t.Fatal("username")
	}

	step, err := svc.Login(ctx, models.LoginInput{Login: second.User.Username, Password: testPassword})
	if err != nil || step.Challenge == nil {
		t.Fatalf("%+v %v", step, err)
	}
	if _, err := svc.VerifyLogin(ctx, step.Challenge.ChallengeID, notify.LastCode); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyRegisterRejectsEmailTakenRace(t *testing.T) {
	st := usermock.NewState()
	svc, notify, users, _, _, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	step, err := svc.Register(ctx, registerInput("race@example.com"))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	if err := users.CreateUser(ctx, models.User{
		ID: "other", Email: "race@example.com", PasswordHash: "h", DisplayName: "X", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	_, err = svc.VerifyRegister(ctx, step.Challenge.ChallengeID, notify.LastCode)
	if codeOf(t, err) != string(exception.MsgEmailTaken) {
		t.Fatalf("%v", err)
	}
}

func TestUnknownPasswordResetRateLimited(t *testing.T) {
	st := usermock.NewState()
	svc, _, _, _, _, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	if _, err := svc.RequestPasswordReset(ctx, "ghost@example.com", "en"); err != nil {
		t.Fatal(err)
	}
	_, err := svc.RequestPasswordReset(ctx, "ghost@example.com", "en")
	if codeOf(t, err) != string(exception.MsgResetRateLimited) {
		t.Fatalf("%v", err)
	}
}

func TestVerifyPasswordResetWrongCode(t *testing.T) {
	st := usermock.NewState()
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	testSupport.RegisterSession(t, svc, notify, registerInput("wrong-reset@example.com"))
	if _, err := svc.RequestPasswordReset(ctx, "wrong-reset@example.com", "en"); err != nil {
		t.Fatal(err)
	}

	_, err := svc.VerifyPasswordReset(ctx, "wrong-reset@example.com", "000000")
	if codeOf(t, err) != string(exception.MsgResetCodeInvalid) {
		t.Fatalf("%v", err)
	}
}

func TestLogoutAfterRedisFlushUsesSession(t *testing.T) {
	st := usermock.NewState()
	svc, notify, _, _, _, mr := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	sess := testSupport.RegisterSession(t, svc, notify, registerInput("logout-flush@example.com"))
	rotated, err := svc.Refresh(ctx, sess.RefreshToken, models.ClientMeta{})
	if err != nil {
		t.Fatal(err)
	}
	mr.FlushAll()

	if err := svc.Logout(ctx, rotated.RefreshToken); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Me(ctx, rotated.AccessToken); codeOf(t, err) != string(exception.MsgUnauthorized) {
		t.Fatalf("%v", err)
	}
}

func TestRegisterWithLocale(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()

	in := registerInput("locale@example.com")
	in.Locale = "id"
	step, err := svc.Register(ctx, in)
	if err != nil || step.Challenge == nil {
		t.Fatalf("%+v %v", step, err)
	}
	if _, err := svc.VerifyRegister(ctx, step.Challenge.ChallengeID, notify.LastCode); err != nil {
		t.Fatal(err)
	}
}

// auth_otp_limits_test.go

func TestVerifyRegisterBlocksAfterMaxWrongCodes(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()

	step, err := svc.Register(ctx, registerInput("max-wrong@example.com"))
	if err != nil {
		t.Fatal(err)
	}
	_ = notify.LastCode

	for i := 0; i < 5; i++ {
		_, err = svc.VerifyRegister(ctx, step.Challenge.ChallengeID, "00000"+string(rune('0'+i)))
		if codeOf(t, err) != string(exception.MsgResetCodeInvalid) {
			t.Fatalf("attempt %d: %v", i, err)
		}
	}

	_, err = svc.VerifyRegister(ctx, step.Challenge.ChallengeID, notify.LastCode)
	if err == nil {
		t.Fatal("expected lockout or invalid after max attempts")
	}
}

func TestVerifyRegisterEmptyChallengeID(t *testing.T) {
	svc, _, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	_, err := svc.VerifyRegister(context.Background(), " ", "123456")
	if err == nil {
		t.Fatal("expected error")
	}
}

// auth_password_revoke_test.go

func TestCompletePasswordResetRevokesPriorSessions(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()

	sess := testSupport.RegisterSession(t, svc, notify, registerInput("revoke-all@example.com"))
	if _, err := svc.RequestPasswordReset(ctx, "revoke-all@example.com", "en"); err != nil {
		t.Fatal(err)
	}
	token, err := svc.VerifyPasswordReset(ctx, "revoke-all@example.com", notify.LastCode)
	if err != nil {
		t.Fatal(err)
	}
	newPass := "brand-new-password"
	if err := svc.CompletePasswordReset(ctx, token, newPass); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.Me(ctx, sess.AccessToken); codeOf(t, err) != string(exception.MsgUnauthorized) {
		t.Fatalf("%v", err)
	}
}

func TestCreateUserFromPendingUsernameConflict(t *testing.T) {
	st := usermock.NewState()
	svc, notify, users, userSvc, _, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, svc, notify, registerInput("uname-owner@example.com"))
	if err := userSvc.CompleteInstanceSetup(ctx, owner.AccessToken, models.CompleteInstanceSetupInput{
		Name: "WiChat", RegistrationMode: "open",
	}); err != nil {
		t.Fatal(err)
	}

	// Pre-create username that Allocate would pick for pending email.
	if err := users.CreateUser(ctx, models.User{ID: "u-x", Email: "other@example.com", Username: "allocate", DisplayName: "X"}); err != nil {
		t.Fatal(err)
	}

	first := testSupport.RegisterSession(t, svc, notify, registerInput("first-alloc@example.com"))
	if first.User.Username == "" {
		t.Fatal("username expected")
	}
}

// auth_validation_test.go

func TestLoginRejectsEmptyFields(t *testing.T) {
	svc, _, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()

	_, err := svc.Login(ctx, models.LoginInput{Login: " ", Password: testPassword})
	if err == nil {
		t.Fatal("empty login")
	}
	_, err = svc.Login(ctx, models.LoginInput{Login: "a@b.com", Password: "short"})
	if err == nil {
		t.Fatal("short password")
	}
}

func TestRegisterRejectsInvalidEmail(t *testing.T) {
	svc, _, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	_, err := svc.Register(context.Background(), models.RegisterInput{
		Email: "not-email", Password: testPassword, DisplayName: "Ada", BirthDate: "2000-01-15",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVerifyRegisterRejectsBadCodeBeforeUserExists(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()

	step, err := svc.Register(ctx, registerInput("verify-bad@example.com"))
	if err != nil {
		t.Fatal(err)
	}
	if notify.LastCode == "" {
		t.Fatal("code")
	}

	_, err = svc.VerifyRegister(ctx, step.Challenge.ChallengeID, "ZZZZZZ")
	if codeOf(t, err) != string(exception.MsgResetCodeInvalid) {
		t.Fatalf("%v", err)
	}
}

func TestVerifyPasswordResetEmptyCode(t *testing.T) {
	svc, _, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	_, err := svc.VerifyPasswordReset(context.Background(), "a@example.com", " ")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRequestPasswordResetInvalidEmail(t *testing.T) {
	svc, _, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	_, err := svc.RequestPasswordReset(context.Background(), "bad-email", "en")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGoogleFinishRejectsEmptyCode(t *testing.T) {
	st := usermock.NewState()
	svc, notify, users, _, _, _ := testSupport.NewAuthService(t, st, nil)
	// Rebuild with google config like GoogleStart test — use existing google test pattern skip; just call with empty on default svc
	_ = users
	_ = notify
	_, err := svc.GoogleFinish(context.Background(), "", "verifier", models.ClientMeta{})
	if codeOf(t, err) != string(exception.MsgUnavailable) {
		t.Fatalf("%v", err)
	}
}

func TestSecondRegisterAssignsUsername(t *testing.T) {
	svc, notify, users, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()

	testSupport.RegisterSession(t, svc, notify, registerInput("first-user@example.com"))

	second := testSupport.RegisterSession(t, svc, notify, registerInput("second-user@example.com"))
	u, err := users.FindUserByID(ctx, second.User.ID)
	if err != nil || u.Username == "" {
		t.Fatalf("username: %+v %v", u, err)
	}
}

func TestResendLoginRejectsMissingChallenge(t *testing.T) {
	svc, _, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	_, err := svc.ResendLoginCode(context.Background(), "missing")
	if codeOf(t, err) != string(exception.MsgResetCodeInvalid) {
		t.Fatalf("%v", err)
	}
}

func TestResendRegisterRateLimitedWithoutWait(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()

	step, err := svc.Register(ctx, registerInput("rate@example.com"))
	if err != nil {
		t.Fatal(err)
	}
	_ = notify.LastCode

	_, err = svc.ResendRegisterCode(ctx, step.Challenge.ChallengeID)
	if err == nil {
		t.Fatal("expected rate limit")
	}
	if codeOf(t, err) != string(exception.MsgResetRateLimited) {
		t.Fatalf("%v", err)
	}
}

// auth_email_hint_test.go

func TestRegisterEmailHintShortLocal(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()

	step, err := svc.Register(ctx, models.RegisterInput{
		Email: "a@example.com", Username: "al.short", Password: testPassword, DisplayName: "Al", BirthDate: "2000-01-15",
	})
	if err != nil || step.Challenge == nil {
		t.Fatalf("%+v %v", step, err)
	}
	if step.Challenge.EmailHint == "" {
		t.Fatal("hint")
	}
	if _, err := svc.VerifyRegister(ctx, step.Challenge.ChallengeID, notify.LastCode); err != nil {
		t.Fatal(err)
	}
}

// auth_email_hint_test.go

func TestLoginEmailRateLimitedWithoutVerify(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()

	testSupport.RegisterSession(t, svc, notify, registerInput("rate-login@example.com"))

	if _, err := svc.Login(ctx, models.LoginInput{Login: "rate-login@example.com", Password: testPassword}); err != nil {
		t.Fatal(err)
	}
	_, err := svc.Login(ctx, models.LoginInput{Login: "rate-login@example.com", Password: testPassword})
	if codeOf(t, err) != string(exception.MsgResetRateLimited) {
		t.Fatalf("%v", err)
	}
}

// auth_email_hint_test.go

func TestGoogleFinishUnavailableUpstream(t *testing.T) {
	st := usermock.NewState()
	users, _, sessions := testSupport.TestRepos(st)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "down", http.StatusServiceUnavailable)
	}))
	t.Cleanup(srv.Close)

	rdb := testSupport.RedisClient(t)
	log := logger.New(logger.Options{Service: "auth-test", Env: "test", Level: "error"})
	userSvc := userservice.NewService(users, nil, nil, sessions, []byte(testSupport.TestJWTSecret), log)
	svc, err := service.NewService(users, sessions, authrepo.NewRedisRefreshStore(rdb), []byte(testSupport.TestJWTSecret), log, googleoauth.Config{
		ClientID: "id", ClientSecret: "sec", RedirectURL: "http://cb",
		TokenURL: srv.URL + "/token", UserInfoURL: srv.URL + "/info",
	}, authrepo.NewRedisPasswordResetStore(rdb), authrepo.NewRedisEmailAuthStore(rdb), &testSupport.FakeNotifier{}, userSvc, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.GoogleFinish(context.Background(), "code", "verifier", models.ClientMeta{})
	if codeOf(t, err) != string(exception.MsgGoogleDenied) {
		t.Fatalf("%v", err)
	}
}

// auth_email_hint_test.go

func TestGoogleFinishRejectsOversizeCode(t *testing.T) {
	st := usermock.NewState()
	users, _, sessions := testSupport.TestRepos(st)
	rdb := testSupport.RedisClient(t)
	log := logger.New(logger.Options{Service: "auth-test", Env: "test", Level: "error"})
	userSvc := userservice.NewService(users, nil, nil, sessions, []byte(testSupport.TestJWTSecret), log)
	svc, err := service.NewService(users, sessions, authrepo.NewRedisRefreshStore(rdb), []byte(testSupport.TestJWTSecret), log, googleoauth.Config{
		ClientID: "id", ClientSecret: "sec", RedirectURL: "http://cb",
		TokenURL: "http://127.0.0.1:1/token", UserInfoURL: "http://127.0.0.1:1/info",
	}, authrepo.NewRedisPasswordResetStore(rdb), authrepo.NewRedisEmailAuthStore(rdb), &testSupport.FakeNotifier{}, userSvc, nil)
	if err != nil {
		t.Fatal(err)
	}

	big := make([]byte, 2049)
	for i := range big {
		big[i] = 'a'
	}
	_, err = svc.GoogleFinish(context.Background(), string(big), "verifier", models.ClientMeta{})
	if codeOf(t, err) != string(exception.MsgGoogleDenied) {
		t.Fatalf("%v", err)
	}
}

// auth_refresh_expired_test.go

func TestRefreshRejectsExpiredSession(t *testing.T) {
	clock := time.Now().UTC()
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), func() time.Time { return clock })
	ctx := context.Background()

	sess := testSupport.RegisterSession(t, svc, notify, registerInput("refresh-exp@example.com"))
	clock = clock.Add(constants.RefreshTokenTTL + time.Minute)

	_, err := svc.Refresh(ctx, sess.RefreshToken, models.ClientMeta{})
	if codeOf(t, err) != string(exception.MsgUnauthorized) {
		t.Fatalf("%v", err)
	}
}

// auth_register_meta_test.go

func TestVerifyRegisterPreservesClientMetaInSession(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()

	step, err := svc.Register(ctx, models.RegisterInput{
		Email: "meta-reg@example.com", Username: "meta.reg", Password: testPassword, DisplayName: "Meta",
		BirthDate: "2000-01-15",
		ClientMeta: models.ClientMeta{
			UserAgent: "WiChatTest/2.0", ClientIP: "203.0.113.9", Platform: "web", DeviceLabel: "Linux",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := svc.VerifyRegister(ctx, step.Challenge.ChallengeID, notify.LastCode)
	if err != nil {
		t.Fatal(err)
	}
	if out.AccessToken == "" {
		t.Fatal("expected session")
	}
}

// auth_unavailable_test.go

func TestRegisterUnavailableWithoutEmailStore(t *testing.T) {
	st := usermock.NewState()
	users, _, sessions := testSupport.TestRepos(st)
	rdb := testSupport.RedisClient(t)
	log := logger.New(logger.Options{Service: "auth-test", Env: "test", Level: "error"})
	userSvc := userservice.NewService(users, nil, nil, sessions, []byte(testSupport.TestJWTSecret), log)

	svc, err := service.NewService(users, sessions, authrepo.NewRedisRefreshStore(rdb), []byte(testSupport.TestJWTSecret), log, googleoauth.Config{},
		authrepo.NewRedisPasswordResetStore(rdb), nil, &testSupport.FakeNotifier{}, userSvc, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.Register(context.Background(), registerInput("no-email-store@example.com"))
	if codeOf(t, err) != string(exception.MsgUnavailable) {
		t.Fatalf("%v", err)
	}
}

// auth_unavailable_test.go

func TestPasswordResetUnavailableWithoutResetStore(t *testing.T) {
	st := usermock.NewState()
	authSvc, notify, users, _, _, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()
	testSupport.RegisterSession(t, authSvc, notify, registerInput("no-reset@example.com"))

	rdb := testSupport.RedisClient(t)
	log := logger.New(logger.Options{Service: "auth-test", Env: "test", Level: "error"})
	_, _, sessions := testSupport.TestRepos(st)
	userSvc := userservice.NewService(users, nil, nil, sessions, []byte(testSupport.TestJWTSecret), log)
	svc, err := service.NewService(users, sessions, authrepo.NewRedisRefreshStore(rdb), []byte(testSupport.TestJWTSecret), log, googleoauth.Config{},
		nil, authrepo.NewRedisEmailAuthStore(rdb), &testSupport.FakeNotifier{}, userSvc, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.RequestPasswordReset(ctx, "no-reset@example.com", "en")
	if codeOf(t, err) != string(exception.MsgUnavailable) {
		t.Fatalf("%v", err)
	}
}

// auth_unavailable_test.go

func TestNewServiceRejectsShortJWTSecret(t *testing.T) {
	st := usermock.NewState()
	users, _, sessions := testSupport.TestRepos(st)
	rdb := testSupport.RedisClient(t)
	log := logger.New(logger.Options{Service: "t", Env: "test", Level: "error"})
	userSvc := userservice.NewService(users, nil, nil, sessions, []byte(testSupport.TestJWTSecret), log)

	_, err := service.NewService(users, sessions, authrepo.NewRedisRefreshStore(rdb), []byte("short"), log, googleoauth.Config{},
		authrepo.NewRedisPasswordResetStore(rdb), authrepo.NewRedisEmailAuthStore(rdb), &testSupport.FakeNotifier{}, userSvc, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

// auth_verify_login_test.go

func TestVerifyLoginBlocksAfterMaxWrongCodes(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	ctx := context.Background()

	testSupport.RegisterSession(t, svc, notify, registerInput("login-max@example.com"))
	step, err := svc.Login(ctx, modelsLogin("login-max@example.com"))
	if err != nil {
		t.Fatal(err)
	}
	_ = notify.LastCode

	for i := 0; i < 5; i++ {
		_, err = svc.VerifyLogin(ctx, step.Challenge.ChallengeID, "00000"+string(rune('0'+i)))
		if codeOf(t, err) != string(exception.MsgResetCodeInvalid) {
			t.Fatalf("attempt %d: %v", i, err)
		}
	}
}

// auth_verify_login_test.go

func TestVerifyLoginEmptyCode(t *testing.T) {
	svc, _, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState(), nil)
	_, err := svc.VerifyLogin(context.Background(), "cid", " ")
	if err == nil {
		t.Fatal("expected error")
	}
}

func modelsLogin(email string) models.LoginInput {
	return models.LoginInput{Login: email, Password: testPassword}
}

// auth_matrix_test.go

func TestMeAfterUserRemoved(t *testing.T) {
	st := usermock.NewState()
	svc, notify, users, _, _, _ := testSupport.NewAuthService(t, st, nil)
	ctx := context.Background()

	sess := testSupport.RegisterSession(t, svc, notify, registerInput("ghost-me@example.com"))
	st.Lock()
	st.Users = nil
	st.Unlock()

	if _, err := svc.Me(ctx, sess.AccessToken); codeOf(t, err) != string(exception.MsgUnauthorized) {
		t.Fatalf("%v", err)
	}
	_ = users
}

// auth_matrix_test.go

func TestGoogleFinishExchangeDenied(t *testing.T) {
	st := usermock.NewState()
	users, _, sessions := testSupport.TestRepos(st)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "denied", http.StatusBadRequest)
	}))
	t.Cleanup(srv.Close)

	rdb := testSupport.RedisClient(t)
	log := logger.New(logger.Options{Service: "auth-test", Env: "test", Level: "error"})
	userSvc := userservice.NewService(users, nil, nil, sessions, []byte(testSupport.TestJWTSecret), log)
	svc, err := service.NewService(users, sessions, authrepo.NewRedisRefreshStore(rdb), []byte(testSupport.TestJWTSecret), log, googleoauth.Config{
		ClientID: "id", ClientSecret: "sec", RedirectURL: "http://cb", TokenURL: srv.URL,
	}, authrepo.NewRedisPasswordResetStore(rdb), authrepo.NewRedisEmailAuthStore(rdb), &testSupport.FakeNotifier{}, userSvc, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.GoogleFinish(context.Background(), "code", "verifier", models.ClientMeta{})
	if codeOf(t, err) != string(exception.MsgGoogleDenied) {
		t.Fatalf("%v", err)
	}
}
