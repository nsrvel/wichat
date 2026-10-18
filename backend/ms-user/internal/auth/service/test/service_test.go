package service_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/auth/service"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	googleoauth "github.com/wichat/wichat/backend/ms-user/pkg/google"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

const testPassword = "correct-horse-battery"

func TestSetupOnce(t *testing.T) {
	svc := newSvc(t)
	ctx := context.Background()

	needs, err := svc.SetupStatus(ctx)
	if err != nil || !needs {
		t.Fatalf("needs setup: %v %v", needs, err)
	}

	first, err := svc.Setup(ctx, setupInput())
	if err != nil {
		t.Fatal(err)
	}
	if first.AccessToken == "" || first.RefreshToken == "" || first.User.Email != "admin@example.com" {
		t.Fatalf("session: %+v", first)
	}

	_, err = svc.Setup(ctx, setupInput())
	if codeOf(t, err) != string(exception.MsgSetupClosed) {
		t.Fatalf("second setup: %v", err)
	}

	needs, err = svc.SetupStatus(ctx)
	if err != nil || needs {
		t.Fatalf("setup closed status: %v %v", needs, err)
	}
}

func TestLoginHidesUnknownEmail(t *testing.T) {
	svc := newSvc(t)
	ctx := context.Background()
	if _, err := svc.Setup(ctx, setupInput()); err != nil {
		t.Fatal(err)
	}

	_, unknown := svc.Login(ctx, models.LoginInput{Email: "missing@example.com", Password: testPassword})
	_, wrong := svc.Login(ctx, models.LoginInput{Email: "admin@example.com", Password: "not-the-password"})
	if codeOf(t, unknown) != string(exception.MsgInvalidCredentials) || codeOf(t, unknown) != codeOf(t, wrong) {
		t.Fatalf("unknown=%v wrong=%v", unknown, wrong)
	}

	got, err := svc.Login(ctx, models.LoginInput{Email: "Admin@Example.com", Password: testPassword})
	if err != nil {
		t.Fatal(err)
	}
	if got.User.Email != "admin@example.com" {
		t.Fatalf("email: %s", got.User.Email)
	}
}

func TestRefreshRotationKeepsAccess(t *testing.T) {
	svc := newSvc(t)
	ctx := context.Background()
	first, err := svc.Setup(ctx, setupInput())
	if err != nil {
		t.Fatal(err)
	}

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
	svc := newSvc(t)
	ctx := context.Background()
	first, err := svc.Setup(ctx, setupInput())
	if err != nil {
		t.Fatal(err)
	}
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
	svc := newSvc(t)
	ctx := context.Background()
	first, err := svc.Setup(ctx, setupInput())
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.Refresh(ctx, "not-a-token", models.ClientMeta{})
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
	svc := newSvc(t)
	ctx := context.Background()
	first, err := svc.Setup(ctx, setupInput())
	if err != nil {
		t.Fatal(err)
	}

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
	repo := &fakeRepo{}
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

	svc, err := service.NewService(repo, []byte("test-secret-must-be-32-characters!!"), logger.New(logger.Options{
		Service: "auth-test",
		Env:     "test",
		Level:   "error",
	}), googleoauth.Config{
		ClientID:     "client",
		ClientSecret: "secret",
		RedirectURL:  "http://localhost:3000/api/v1/auth/google/callback",
		TokenURL:     google.URL + "/token",
		UserInfoURL:  google.URL + "/userinfo",
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if _, err := svc.Setup(ctx, setupInput()); err != nil {
		t.Fatal(err)
	}

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

	_, err = svc.Login(ctx, models.LoginInput{Email: "admin@example.com", Password: testPassword})
	if err != nil {
		t.Fatal(err)
	}

	mode = "second"
	created, err := svc.GoogleFinish(ctx, "code", "verifier", models.ClientMeta{})
	if err != nil {
		t.Fatal(err)
	}
	if created.User.ID == linked.User.ID || created.User.Email != "new@example.com" || created.User.DisplayName != "New Person" {
		t.Fatalf("created: %+v", created.User)
	}
	_, err = svc.Login(ctx, models.LoginInput{Email: "new@example.com", Password: testPassword})
	if codeOf(t, err) != string(exception.MsgInvalidCredentials) {
		t.Fatalf("password on google account: %v", err)
	}

	mode = "unverified"
	_, err = svc.GoogleFinish(ctx, "code", "verifier", models.ClientMeta{})
	if codeOf(t, err) != string(exception.MsgGoogleEmailUnverified) {
		t.Fatalf("unverified: %v", err)
	}
}

func newSvc(t *testing.T) service.Service {
	t.Helper()

	svc, err := service.NewService(&fakeRepo{}, []byte("test-secret-must-be-32-characters!!"), logger.New(logger.Options{
		Service: "auth-test",
		Env:     "test",
		Level:   "error",
	}), googleoauth.Config{})
	if err != nil {
		t.Fatal(err)
	}

	return svc
}

func setupInput() models.SetupInput {
	return models.SetupInput{
		Email:       "admin@example.com",
		Password:    testPassword,
		DisplayName: "Ada",
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

var _ repository.Repository = (*fakeRepo)(nil)
