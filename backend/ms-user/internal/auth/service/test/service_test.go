package service_test

import (
	"context"
	"net/http"
	"net/http/httptest"
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
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

const testPassword = "correct-horse-battery"

func registerInput(email string) models.RegisterInput {
	return models.RegisterInput{
		Email:       email,
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
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState())
	out := testSupport.RegisterSession(t, svc, notify, models.RegisterInput{
		Email:       "unicode@example.com",
		Password:    testPassword,
		DisplayName: "山田 太郎 🎮",
		BirthDate:   "2000-01-15",
	})
	if out.User.DisplayName != "山田 太郎 🎮" {
		t.Fatalf("display name: %q", out.User.DisplayName)
	}
}

func TestRegisterFirstUserIsOwner(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState())
	ctx := context.Background()

	first := testSupport.RegisterSession(t, svc, notify, registerInput("admin@example.com"))
	if first.AccessToken == "" || first.RefreshToken == "" || first.User.Email != "admin@example.com" {
		t.Fatalf("session: %+v", first)
	}
	if first.User.Username != "admin" {
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
	svc, notify, _, userSvc, _, _ := testSupport.NewAuthService(t, usermock.NewState())
	ctx := context.Background()

	owner := testSupport.RegisterSession(t, svc, notify, registerInput("admin@example.com"))
	if err := userSvc.CompleteInstanceSetup(ctx, owner.AccessToken, models.CompleteInstanceSetupInput{
		Name:             "WiChat",
		RegistrationMode: constants.RegistrationOpen,
		UsernameEnabled:  true,
	}); err != nil {
		t.Fatal(err)
	}

	first := testSupport.RegisterSession(t, svc, notify, registerInput("john@a.com"))
	if first.User.Username != "john" {
		t.Fatalf("first username: %q", first.User.Username)
	}

	second := testSupport.RegisterSession(t, svc, notify, registerInput("john@b.com"))
	if second.User.Username != "john2" {
		t.Fatalf("second username: %q", second.User.Username)
	}
}

func TestLoginHidesUnknownEmail(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState())
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

func TestRefreshRotationKeepsAccess(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState())
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

func TestRefreshReuseRevokesSession(t *testing.T) {
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState())
	ctx := context.Background()
	first := testSupport.RegisterSession(t, svc, notify, registerInput("admin@example.com"))
	second, err := svc.Refresh(ctx, first.RefreshToken, models.ClientMeta{})
	if err != nil {
		t.Fatal(err)
	}

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
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState())
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
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState())
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
	users, sessions := testSupport.TestRepos(st)
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
	log := logger.New(logger.Options{Service: "auth-test", Env: "test", Level: "error"})
	userSvc := userservice.NewService(users, sessions, []byte(testSupport.TestJWTSecret), log)
	svc, err := service.NewService(users, sessions, []byte(testSupport.TestJWTSecret), log, googleoauth.Config{
		ClientID:     "client",
		ClientSecret: "secret",
		RedirectURL:  "http://localhost:3000/api/v1/auth/google/callback",
		TokenURL:     google.URL + "/token",
		UserInfoURL:  google.URL + "/userinfo",
	}, authrepo.NewRedisPasswordResetStore(rdb), authrepo.NewRedisEmailAuthStore(rdb), notify, userSvc)
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
	svc, notify, users, _, _, _ := testSupport.NewAuthService(t, st)
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
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState())
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
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState())
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
	svc, notify, _, _, _, _ := testSupport.NewAuthService(t, usermock.NewState())

	if _, err := svc.RequestPasswordReset(context.Background(), "missing@example.com", "en"); err != nil {
		t.Fatal(err)
	}
	if notify.Sends != 0 {
		t.Fatalf("expected no send, got %d", notify.Sends)
	}
}

func TestRequestPasswordResetSendsForKnownUser(t *testing.T) {
	st := usermock.NewState()
	svc, notify, _, _, _, mr := testSupport.NewAuthService(t, st)
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
	svc, notify, _, _, _, mr := testSupport.NewAuthService(t, st)
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
