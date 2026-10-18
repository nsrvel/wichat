package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authdelivery "github.com/wichat/wichat/backend/api-gateway/internal/auth/delivery/http"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
	msUser "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user"
	"github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/auth"
	authv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/auth/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const refreshSecret = "super-secret-refresh"

func TestLoginStripsRefreshFromJSON(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(fakeAuth{
		login: func(context.Context, *authv1.LoginRequest) (*authv1.LoginResponse, error) {
			return &authv1.LoginResponse{Session: &authv1.Session{
				AccessToken:  "access-token",
				ExpiresIn:    900,
				RefreshToken: refreshSecret,
				User:         &authv1.User{Id: "u1", Email: "ada@example.com", DisplayName: "Ada"},
			}}, nil
		},
	}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"ada@example.com","password":"correct-horse-battery"}`))
	rec := httptest.NewRecorder()
	handler.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), refreshSecret) {
		t.Fatalf("refresh token leaked: %s", rec.Body.String())
	}

	var body struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.AccessToken != "access-token" {
		t.Fatalf("access token: %+v", body)
	}

	res := rec.Result()
	t.Cleanup(func() { _ = res.Body.Close() })
	var cookie *http.Cookie
	for _, item := range res.Cookies() {
		if item.Name == "wichat_refresh" {
			cookie = item
		}
	}
	if cookie == nil {
		t.Fatal("missing refresh cookie")
	}
	if cookie.Value != refreshSecret || !cookie.HttpOnly || cookie.Path != "/api/v1/auth" {
		t.Fatalf("cookie: %+v", cookie)
	}
}

type fakeAuth struct {
	login        func(context.Context, *authv1.LoginRequest) (*authv1.LoginResponse, error)
	googleFinish func(context.Context, *authv1.GoogleFinishRequest) (*authv1.GoogleFinishResponse, error)
}

func testClients(authClient auth.Client) *microservices.Clients {
	return &microservices.Clients{User: msUser.Service{GRPC: msUser.GRPC{Auth: authClient}}}
}

func (f fakeAuth) SetupStatus(context.Context, *authv1.SetupStatusRequest) (*authv1.SetupStatusResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) Setup(context.Context, *authv1.SetupRequest) (*authv1.SetupResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) Login(ctx context.Context, in *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	return f.login(ctx, in)
}

func (f fakeAuth) Refresh(context.Context, *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) Logout(context.Context, *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) Me(context.Context, *authv1.MeRequest) (*authv1.MeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) GoogleStart(context.Context, *authv1.GoogleStartRequest) (*authv1.GoogleStartResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) GoogleFinish(ctx context.Context, in *authv1.GoogleFinishRequest) (*authv1.GoogleFinishResponse, error) {
	if f.googleFinish == nil {
		return nil, status.Error(codes.Unimplemented, "error.internal")
	}

	return f.googleFinish(ctx, in)
}

func TestGoogleCallbackRedirectsWithRefreshCookie(t *testing.T) {
	var verifier string
	handler := authdelivery.NewHandler(testClients(fakeAuth{
		googleFinish: func(_ context.Context, in *authv1.GoogleFinishRequest) (*authv1.GoogleFinishResponse, error) {
			verifier = in.GetCodeVerifier()
			return &authv1.GoogleFinishResponse{Session: &authv1.Session{
				AccessToken:  "access-token",
				ExpiresIn:    900,
				RefreshToken: refreshSecret,
				User:         &authv1.User{Id: "u1", Email: "ada@example.com", DisplayName: "Ada"},
			}}, nil
		},
	}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/google/callback?code=abc&state=state-1", nil)
	req.AddCookie(&http.Cookie{Name: "wichat_oauth", Value: "state-1.verifier-1"})
	rec := httptest.NewRecorder()
	handler.GoogleCallback(rec, req)

	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "http://localhost:8080/" {
		t.Fatalf("redirect: %d %s", rec.Code, rec.Header().Get("Location"))
	}
	if strings.Contains(rec.Header().Get("Location"), refreshSecret) || verifier != "verifier-1" {
		t.Fatalf("verifier %q location %s", verifier, rec.Header().Get("Location"))
	}

	res := rec.Result()
	t.Cleanup(func() { _ = res.Body.Close() })
	var cookie *http.Cookie
	for _, item := range res.Cookies() {
		if item.Name == "wichat_refresh" {
			cookie = item
		}
	}
	if cookie == nil || cookie.Value != refreshSecret || !cookie.HttpOnly {
		t.Fatalf("cookie: %+v", cookie)
	}
}

func TestGoogleCallbackAliasPath(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(fakeAuth{
		googleFinish: func(_ context.Context, in *authv1.GoogleFinishRequest) (*authv1.GoogleFinishResponse, error) {
			return &authv1.GoogleFinishResponse{Session: &authv1.Session{
				AccessToken:  "access-token",
				ExpiresIn:    900,
				RefreshToken: refreshSecret,
				User:         &authv1.User{Id: "u1", Email: "ada@example.com", DisplayName: "Ada"},
			}}, nil
		},
	}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodGet, "/api/auth/callback/google?code=abc&state=state-1", nil)
	req.AddCookie(&http.Cookie{Name: "wichat_oauth", Value: "state-1.verifier-1", Path: "/api"})
	rec := httptest.NewRecorder()
	handler.GoogleCallback(rec, req)

	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "http://localhost:8080/" {
		t.Fatalf("redirect: %d %s", rec.Code, rec.Header().Get("Location"))
	}
}

func TestGoogleCallbackRejectsStateMismatch(t *testing.T) {
	called := false
	handler := authdelivery.NewHandler(testClients(fakeAuth{
		googleFinish: func(context.Context, *authv1.GoogleFinishRequest) (*authv1.GoogleFinishResponse, error) {
			called = true
			return nil, nil
		},
	}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/google/callback?code=abc&state=other", nil)
	req.AddCookie(&http.Cookie{Name: "wichat_oauth", Value: "state-1.verifier-1"})
	rec := httptest.NewRecorder()
	handler.GoogleCallback(rec, req)

	if called || rec.Code != http.StatusFound || rec.Header().Get("Location") != "http://localhost:8080/sign-in?error=auth.google_denied" {
		t.Fatalf("called %v redirect %d %s", called, rec.Code, rec.Header().Get("Location"))
	}
}

func testLog() logger.Logger {
	return logger.New(logger.Options{Service: "gateway-test", Env: "test", Level: "error"})
}
