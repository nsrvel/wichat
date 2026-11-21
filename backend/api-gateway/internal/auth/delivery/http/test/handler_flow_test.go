package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authdelivery "github.com/wichat/wichat/backend/api-gateway/internal/auth/delivery/http"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

func TestLoginMFAReturnsChallenge(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{
		login: func(context.Context, *userv1.LoginRequest) (*userv1.LoginResponse, error) {
			return &userv1.LoginResponse{Challenge: &userv1.AuthChallenge{
				ChallengeId: "mfa-1",
				MfaRequired: true,
				MfaMethods:  []string{"totp"},
			}}, nil
		},
	}}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"login":"ada","password":"ok"}`))
	rec := httptest.NewRecorder()
	handler.Login(rec, req)

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"mfa_required":true`) {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestRegisterReturnsChallenge(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{
		register: func(context.Context, *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
			return &userv1.RegisterResponse{Challenge: &userv1.AuthChallenge{
				ChallengeId:           "reg-1",
				OtpRequired:           true,
				ResendCooldownSeconds: 60,
			}}, nil
		},
	}}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"email":"a@b.com","username":"ada","password":"longpassword12","display_name":"Ada","birth_date":"2000-01-01"}`))
	rec := httptest.NewRecorder()
	handler.Register(rec, req)

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"challenge_id":"reg-1"`) {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestVerifyRegisterSetsSession(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{
		verifyRegister: func(context.Context, *userv1.VerifyRegisterRequest) (*userv1.VerifyRegisterResponse, error) {
			return &userv1.VerifyRegisterResponse{Session: &userv1.Session{
				AccessToken:  "tok",
				RefreshToken: refreshSecret,
				ExpiresIn:    900,
				User:         &userv1.User{Id: "u1"},
			}}, nil
		},
	}}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register/verify", strings.NewReader(`{"challenge_id":"reg-1","code":"ABC123"}`))
	rec := httptest.NewRecorder()
	handler.VerifyRegister(rec, req)

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"access_token":"tok"`) {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestMeRejectsMissingBearer(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{}), testLog(), false, "http://localhost:8080")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	rec := httptest.NewRecorder()
	handler.Me(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestGoogleStartRedirects(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{
		googleStart: func(context.Context, *userv1.GoogleStartRequest) (*userv1.GoogleStartResponse, error) {
			return &userv1.GoogleStartResponse{
				AuthorizationUrl: "https://accounts.google.com/o/oauth2/v2/auth",
				State:            "st",
				CodeVerifier:     "cv",
			}, nil
		},
	}}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/google", nil)
	rec := httptest.NewRecorder()
	handler.GoogleStart(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status %d", rec.Code)
	}
}
