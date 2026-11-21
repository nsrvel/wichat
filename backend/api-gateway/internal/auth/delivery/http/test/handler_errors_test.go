package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authdelivery "github.com/wichat/wichat/backend/api-gateway/internal/auth/delivery/http"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLoginInvalidJSON(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{}), testLog(), false, "http://localhost:8080")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{`))
	rec := httptest.NewRecorder()
	handler.Login(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestGoogleStartGrpcErrorRedirects(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{
		googleStart: func(context.Context, *userv1.GoogleStartRequest) (*userv1.GoogleStartResponse, error) {
			return nil, status.Error(codes.Unavailable, "auth.unavailable")
		},
	}}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/google", nil)
	rec := httptest.NewRecorder()
	handler.GoogleStart(rec, req)

	if rec.Code != http.StatusFound || !strings.Contains(rec.Header().Get("Location"), "auth.unavailable") {
		t.Fatalf("redirect %d %s", rec.Code, rec.Header().Get("Location"))
	}
}

func TestGoogleCallbackUserDenied(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/google/callback?error=access_denied", nil)
	rec := httptest.NewRecorder()
	handler.GoogleCallback(rec, req)

	if rec.Code != http.StatusFound || !strings.Contains(rec.Header().Get("Location"), "google_denied") {
		t.Fatalf("redirect %d %s", rec.Code, rec.Header().Get("Location"))
	}
}

func TestGoogleStartEmptyAuthURLRedirectsUnavailable(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{
		googleStart: func(context.Context, *userv1.GoogleStartRequest) (*userv1.GoogleStartResponse, error) {
			return &userv1.GoogleStartResponse{}, nil
		},
	}}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/google", nil)
	rec := httptest.NewRecorder()
	handler.GoogleStart(rec, req)

	if rec.Code != http.StatusFound || !strings.Contains(rec.Header().Get("Location"), "error=") {
		t.Fatalf("redirect %d %s", rec.Code, rec.Header().Get("Location"))
	}
}

func TestRefreshGrpcError(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{
		refresh: func(context.Context, *userv1.RefreshRequest) (*userv1.RefreshResponse, error) {
			return nil, status.Error(codes.Unauthenticated, "auth.unauthorized")
		},
	}}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "wichat_refresh", Value: refreshSecret})
	rec := httptest.NewRecorder()
	handler.Refresh(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestVerifyLoginGrpcError(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{
		verifyLogin: func(context.Context, *userv1.VerifyLoginRequest) (*userv1.VerifyLoginResponse, error) {
			return nil, status.Error(codes.InvalidArgument, "auth.reset_code_invalid")
		},
	}}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login/verify", strings.NewReader(`{"challenge_id":"ch-1","code":"BAD"}`))
	rec := httptest.NewRecorder()
	handler.VerifyLogin(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestBeginLoginWebAuthnError(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{
		beginLoginWebAuthn: func(context.Context, *userv1.BeginLoginWebAuthnRequest) (*userv1.BeginLoginWebAuthnResponse, error) {
			return nil, status.Error(codes.NotFound, "error.not_found")
		},
	}}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login/mfa/webauthn/begin", strings.NewReader(`{"challenge_id":"mfa-1"}`))
	rec := httptest.NewRecorder()
	handler.BeginLoginWebAuthn(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}
