package http_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authdelivery "github.com/wichat/wichat/backend/api-gateway/internal/auth/delivery/http"
)

const turnstileTestSecret = "test-turnstile-secret"

func TestLoginRejectsMissingTurnstileWhenEnabled(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{}), testLog(), false, "http://localhost:8080", turnstileTestSecret)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"login":"ada","password":"secret-password12"}`))
	rec := httptest.NewRecorder()
	handler.Login(rec, req)

	if rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), "captcha_failed") {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestVerifyLoginRejectsMissingTurnstileWhenEnabled(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{}), testLog(), false, "http://localhost:8080", turnstileTestSecret)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login/verify", strings.NewReader(`{"challenge_id":"ch-1","code":"ABC123"}`))
	rec := httptest.NewRecorder()
	handler.VerifyLogin(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestRegisterRejectsMissingTurnstileWhenEnabled(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{}), testLog(), false, "http://localhost:8080", turnstileTestSecret)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"email":"a@b.com","username":"ada","password":"longpassword12","display_name":"Ada","birth_date":"2000-01-01"}`))
	rec := httptest.NewRecorder()
	handler.Register(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestGoogleStartRejectsMissingTurnstileWhenEnabled(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{}), testLog(), false, "http://localhost:8080", turnstileTestSecret)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/google", nil)
	rec := httptest.NewRecorder()
	handler.GoogleStart(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestResendLoginRejectsMissingTurnstileWhenEnabled(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{}), testLog(), false, "http://localhost:8080", turnstileTestSecret)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login/resend", strings.NewReader(`{"challenge_id":"ch-1"}`))
	rec := httptest.NewRecorder()
	handler.ResendLoginCode(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}
