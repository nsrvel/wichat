package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authdelivery "github.com/wichat/wichat/backend/api-gateway/internal/auth/delivery/http"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func testSession() *userv1.Session {
	return &userv1.Session{
		AccessToken:  "access-new",
		RefreshToken: refreshSecret,
		ExpiresIn:    900,
		User:         &userv1.User{Id: "u1", Email: "ada@example.com", DisplayName: "Ada"},
	}
}

func TestForgotPasswordFlow(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{}}), testLog(), false, "http://localhost:8080")

	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", strings.NewReader(`{"email":"ada@example.com","locale":"en"}`))
	rec1 := httptest.NewRecorder()
	handler.ForgotPasswordRequest(rec1, req1)
	if rec1.Code != http.StatusOK || !strings.Contains(rec1.Body.String(), `"ok":true`) {
		t.Fatalf("request: %d %s", rec1.Code, rec1.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password/verify", strings.NewReader(`{"email":"ada@example.com","code":"ABC123"}`))
	rec2 := httptest.NewRecorder()
	handler.ForgotPasswordVerify(rec2, req2)
	if rec2.Code != http.StatusOK || !strings.Contains(rec2.Body.String(), `"reset_token":"token-1"`) {
		t.Fatalf("verify: %d %s", rec2.Code, rec2.Body.String())
	}

	req3 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password/reset", strings.NewReader(`{"reset_token":"token-1","password":"new-password-12chars"}`))
	rec3 := httptest.NewRecorder()
	handler.ForgotPasswordReset(rec3, req3)
	if rec3.Code != http.StatusOK || !strings.Contains(rec3.Body.String(), `"ok":true`) {
		t.Fatalf("reset: %d %s", rec3.Code, rec3.Body.String())
	}
}

func TestForgotPasswordVerifyGrpcError(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{
		verifyPasswordReset: func(context.Context, *userv1.VerifyPasswordResetRequest) (*userv1.VerifyPasswordResetResponse, error) {
			return nil, status.Error(codes.InvalidArgument, string(exception.MsgResetCodeInvalid))
		},
	}}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password/verify", strings.NewReader(`{"email":"ada@example.com","code":"BAD"}`))
	rec := httptest.NewRecorder()
	handler.ForgotPasswordVerify(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestRefreshSetsSession(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{
		refresh: func(context.Context, *userv1.RefreshRequest) (*userv1.RefreshResponse, error) {
			return &userv1.RefreshResponse{Session: testSession()}, nil
		},
	}}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "wichat_refresh", Value: refreshSecret})
	rec := httptest.NewRecorder()
	handler.Refresh(rec, req)

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"access-new"`) {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestRefreshWithoutCookie(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{}), testLog(), false, "http://localhost:8080")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	rec := httptest.NewRecorder()
	handler.Refresh(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestLogoutClearsCookie(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{
		logout: func(context.Context, *userv1.LogoutRequest) (*userv1.LogoutResponse, error) {
			return &userv1.LogoutResponse{Ok: true}, nil
		},
	}}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "wichat_refresh", Value: refreshSecret})
	rec := httptest.NewRecorder()
	handler.Logout(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	res := rec.Result()
	t.Cleanup(func() { _ = res.Body.Close() })
	var cleared bool
	for _, c := range res.Cookies() {
		if c.Name == "wichat_refresh" && c.MaxAge < 0 {
			cleared = true
		}
	}
	if !cleared {
		t.Fatal("expected refresh cookie cleared")
	}
}

func TestLogoutWithoutCookieStillOK(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{}), testLog(), false, "http://localhost:8080")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	rec := httptest.NewRecorder()
	handler.Logout(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestMeReturnsUser(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{
		me: func(context.Context, *userv1.MeRequest) (*userv1.MeResponse, error) {
			return &userv1.MeResponse{User: &userv1.User{Id: "u1", Email: "ada@example.com"}}, nil
		},
	}}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer access-token")
	rec := httptest.NewRecorder()
	handler.Me(rec, req)

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"ada@example.com"`) {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestSetupDelegatesToRegister(t *testing.T) {
	called := false
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{
		register: func(context.Context, *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
			called = true
			return &userv1.RegisterResponse{Challenge: &userv1.AuthChallenge{ChallengeId: "reg-1"}}, nil
		},
	}}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/setup", strings.NewReader(`{"email":"a@b.com","username":"ada","password":"longpassword12","display_name":"Ada","birth_date":"2000-01-01"}`))
	rec := httptest.NewRecorder()
	handler.Setup(rec, req)

	if !called || rec.Code != http.StatusOK {
		t.Fatalf("called=%v status=%d", called, rec.Code)
	}
}

func TestResendRegisterCode(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{
		resendRegister: func(context.Context, *userv1.ResendRegisterCodeRequest) (*userv1.ResendRegisterCodeResponse, error) {
			return &userv1.ResendRegisterCodeResponse{ResendCooldownSeconds: 30}, nil
		},
	}}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register/resend", strings.NewReader(`{"challenge_id":"reg-1"}`))
	rec := httptest.NewRecorder()
	handler.ResendRegisterCode(rec, req)

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"resend_after_seconds":30`) {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestLoginUsesEmailFieldFallback(t *testing.T) {
	var login string
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{
		login: func(_ context.Context, in *userv1.LoginRequest) (*userv1.LoginResponse, error) {
			login = in.GetLogin()
			return &userv1.LoginResponse{Session: testSession()}, nil
		},
	}}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"ada@example.com","password":"secret-password12"}`))
	rec := httptest.NewRecorder()
	handler.Login(rec, req)

	if rec.Code != http.StatusOK || login != "ada@example.com" {
		t.Fatalf("login=%q status %d", login, rec.Code)
	}
}
