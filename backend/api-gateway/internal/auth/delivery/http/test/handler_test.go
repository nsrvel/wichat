package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authdelivery "github.com/wichat/wichat/backend/api-gateway/internal/auth/delivery/http"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
	msUser "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user"
	"github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/auth"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const refreshSecret = "super-secret-refresh"

const turnstileTestSecret = "test-turnstile-secret"

func TestLoginReturnsChallengeWithoutRefreshCookie(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{
		login: func(context.Context, *userv1.LoginRequest) (*userv1.LoginResponse, error) {
			return &userv1.LoginResponse{Challenge: &userv1.AuthChallenge{
				ChallengeId:           "ch-1",
				EmailHint:             "a***@example.com",
				ResendCooldownSeconds: 60,
				OtpRequired:           true,
			}}, nil
		},
	}}), testLog(), false, "http://localhost:8080")

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
			ChallengeID string `json:"challenge_id"`
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.ChallengeID != "ch-1" || body.Data.AccessToken != "" {
		t.Fatalf("challenge: %+v", body.Data)
	}
	if strings.Contains(rec.Body.String(), "email_hint") {
		t.Fatalf("email_hint leaked: %s", rec.Body.String())
	}

	res := rec.Result()
	t.Cleanup(func() { _ = res.Body.Close() })
	for _, item := range res.Cookies() {
		if item.Name == "wichat_refresh" {
			t.Fatal("unexpected refresh cookie on challenge response")
		}
	}
}

func TestVerifyLoginSetsRefreshCookie(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{
		verifyLogin: func(context.Context, *userv1.VerifyLoginRequest) (*userv1.VerifyLoginResponse, error) {
			return &userv1.VerifyLoginResponse{Session: &userv1.Session{
				AccessToken:  "access-token",
				ExpiresIn:    900,
				RefreshToken: refreshSecret,
				User:         &userv1.User{Id: "u1", Email: "ada@example.com", DisplayName: "Ada"},
			}}, nil
		},
	}}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login/verify", strings.NewReader(`{"challenge_id":"ch-1","code":"ABC123"}`))
	rec := httptest.NewRecorder()
	handler.VerifyLogin(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}

	res := rec.Result()
	t.Cleanup(func() { _ = res.Body.Close() })
	var cookie *http.Cookie
	for _, item := range res.Cookies() {
		if item.Name == "wichat_refresh" {
			cookie = item
		}
	}
	if cookie == nil || cookie.Value != refreshSecret {
		t.Fatalf("cookie: %+v", cookie)
	}
}

type msUserFakes struct {
	Auth auth.Client
}

func testClients(f msUserFakes) *microservices.Clients {
	if f.Auth == nil {
		f.Auth = fakeAuth{}
	}

	return &microservices.Clients{User: msUser.Service{GRPC: msUser.GRPC{
		Auth: f.Auth,
	}}}
}

type fakeAuth struct {
	login                 func(context.Context, *userv1.LoginRequest) (*userv1.LoginResponse, error)
	verifyLogin           func(context.Context, *userv1.VerifyLoginRequest) (*userv1.VerifyLoginResponse, error)
	verifyLoginMFA        func(context.Context, *userv1.VerifyLoginMFARequest) (*userv1.VerifyLoginMFAResponse, error)
	beginLoginWebAuthn    func(context.Context, *userv1.BeginLoginWebAuthnRequest) (*userv1.BeginLoginWebAuthnResponse, error)
	googleFinish          func(context.Context, *userv1.GoogleFinishRequest) (*userv1.GoogleFinishResponse, error)
	resendLogin           func(context.Context, *userv1.ResendLoginCodeRequest) (*userv1.ResendLoginCodeResponse, error)
	register              func(context.Context, *userv1.RegisterRequest) (*userv1.RegisterResponse, error)
	verifyRegister        func(context.Context, *userv1.VerifyRegisterRequest) (*userv1.VerifyRegisterResponse, error)
	googleStart           func(context.Context, *userv1.GoogleStartRequest) (*userv1.GoogleStartResponse, error)
	refresh               func(context.Context, *userv1.RefreshRequest) (*userv1.RefreshResponse, error)
	logout                func(context.Context, *userv1.LogoutRequest) (*userv1.LogoutResponse, error)
	me                    func(context.Context, *userv1.MeRequest) (*userv1.MeResponse, error)
	resendRegister        func(context.Context, *userv1.ResendRegisterCodeRequest) (*userv1.ResendRegisterCodeResponse, error)
	verifyPasswordReset   func(context.Context, *userv1.VerifyPasswordResetRequest) (*userv1.VerifyPasswordResetResponse, error)
	requestPasswordReset  func(context.Context, *userv1.RequestPasswordResetRequest) (*userv1.RequestPasswordResetResponse, error)
	completePasswordReset func(context.Context, *userv1.CompletePasswordResetRequest) (*userv1.CompletePasswordResetResponse, error)
}

func (f fakeAuth) Setup(context.Context, *userv1.SetupRequest) (*userv1.SetupResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) Register(ctx context.Context, in *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	if f.register != nil {
		return f.register(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) Login(ctx context.Context, in *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	return f.login(ctx, in)
}

func (f fakeAuth) VerifyRegister(ctx context.Context, in *userv1.VerifyRegisterRequest) (*userv1.VerifyRegisterResponse, error) {
	if f.verifyRegister != nil {
		return f.verifyRegister(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) VerifyLogin(ctx context.Context, in *userv1.VerifyLoginRequest) (*userv1.VerifyLoginResponse, error) {
	if f.verifyLogin != nil {
		return f.verifyLogin(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) VerifyLoginMFA(ctx context.Context, in *userv1.VerifyLoginMFARequest) (*userv1.VerifyLoginMFAResponse, error) {
	if f.verifyLoginMFA != nil {
		return f.verifyLoginMFA(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) BeginLoginWebAuthn(ctx context.Context, in *userv1.BeginLoginWebAuthnRequest) (*userv1.BeginLoginWebAuthnResponse, error) {
	if f.beginLoginWebAuthn != nil {
		return f.beginLoginWebAuthn(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) ResendRegisterCode(ctx context.Context, in *userv1.ResendRegisterCodeRequest) (*userv1.ResendRegisterCodeResponse, error) {
	if f.resendRegister != nil {
		return f.resendRegister(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) ResendLoginCode(ctx context.Context, in *userv1.ResendLoginCodeRequest) (*userv1.ResendLoginCodeResponse, error) {
	if f.resendLogin != nil {
		return f.resendLogin(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) Refresh(ctx context.Context, in *userv1.RefreshRequest) (*userv1.RefreshResponse, error) {
	if f.refresh != nil {
		return f.refresh(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) Logout(ctx context.Context, in *userv1.LogoutRequest) (*userv1.LogoutResponse, error) {
	if f.logout != nil {
		return f.logout(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) Me(ctx context.Context, in *userv1.MeRequest) (*userv1.MeResponse, error) {
	if f.me != nil {
		return f.me(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) GoogleStart(ctx context.Context, in *userv1.GoogleStartRequest) (*userv1.GoogleStartResponse, error) {
	if f.googleStart != nil {
		return f.googleStart(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) RequestPasswordReset(ctx context.Context, in *userv1.RequestPasswordResetRequest) (*userv1.RequestPasswordResetResponse, error) {
	if f.requestPasswordReset != nil {
		return f.requestPasswordReset(ctx, in)
	}
	return &userv1.RequestPasswordResetResponse{Ok: true, ResendCooldownSeconds: 60}, nil
}

func (f fakeAuth) VerifyPasswordReset(ctx context.Context, in *userv1.VerifyPasswordResetRequest) (*userv1.VerifyPasswordResetResponse, error) {
	if f.verifyPasswordReset != nil {
		return f.verifyPasswordReset(ctx, in)
	}
	return &userv1.VerifyPasswordResetResponse{ResetToken: "token-1"}, nil
}

func (f fakeAuth) CompletePasswordReset(ctx context.Context, in *userv1.CompletePasswordResetRequest) (*userv1.CompletePasswordResetResponse, error) {
	if f.completePasswordReset != nil {
		return f.completePasswordReset(ctx, in)
	}
	return &userv1.CompletePasswordResetResponse{Ok: true}, nil
}

func (f fakeAuth) GoogleFinish(ctx context.Context, in *userv1.GoogleFinishRequest) (*userv1.GoogleFinishResponse, error) {
	if f.googleFinish == nil {
		return nil, status.Error(codes.Unimplemented, "error.internal")
	}

	return f.googleFinish(ctx, in)
}

func TestGoogleCallbackRedirectsWithRefreshCookie(t *testing.T) {
	var verifier string
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{
		googleFinish: func(_ context.Context, in *userv1.GoogleFinishRequest) (*userv1.GoogleFinishResponse, error) {
			verifier = in.GetCodeVerifier()
			return &userv1.GoogleFinishResponse{Session: &userv1.Session{
				AccessToken:  "access-token",
				ExpiresIn:    900,
				RefreshToken: refreshSecret,
				User:         &userv1.User{Id: "u1", Email: "ada@example.com", DisplayName: "Ada"},
			}}, nil
		},
	}}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/google/callback?code=abc&state=state-1", nil)
	req.AddCookie(&http.Cookie{Name: "wichat_oauth", Value: "state-1.verifier-1"})
	rec := httptest.NewRecorder()
	handler.GoogleCallback(rec, req)

	wantLocation := "http://localhost:8080/"
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != wantLocation {
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

func TestGoogleCallbackRedirectsToOAuthReturnCookie(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{
		googleFinish: func(_ context.Context, _ *userv1.GoogleFinishRequest) (*userv1.GoogleFinishResponse, error) {
			return &userv1.GoogleFinishResponse{Session: &userv1.Session{
				AccessToken:  "access-token",
				ExpiresIn:    900,
				RefreshToken: refreshSecret,
				User:         &userv1.User{Id: "u1", Email: "ada@example.com", DisplayName: "Ada"},
			}}, nil
		},
	}}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/google/callback?code=abc&state=state-1", nil)
	req.AddCookie(&http.Cookie{Name: "wichat_oauth", Value: "state-1.verifier-1"})
	req.AddCookie(&http.Cookie{Name: "wichat_oauth_return", Value: "/invite/LK1FjUZy", Path: "/api"})
	rec := httptest.NewRecorder()
	handler.GoogleCallback(rec, req)

	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "http://localhost:8080/invite/LK1FjUZy" {
		t.Fatalf("redirect: %d %s", rec.Code, rec.Header().Get("Location"))
	}
}

func TestGoogleCallbackAliasPath(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{
		googleFinish: func(_ context.Context, in *userv1.GoogleFinishRequest) (*userv1.GoogleFinishResponse, error) {
			return &userv1.GoogleFinishResponse{Session: &userv1.Session{
				AccessToken:  "access-token",
				ExpiresIn:    900,
				RefreshToken: refreshSecret,
				User:         &userv1.User{Id: "u1", Email: "ada@example.com", DisplayName: "Ada"},
			}}, nil
		},
	}}), testLog(), false, "http://localhost:8080")

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
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{
		googleFinish: func(context.Context, *userv1.GoogleFinishRequest) (*userv1.GoogleFinishResponse, error) {
			called = true
			return nil, nil
		},
	}}), testLog(), false, "http://localhost:8080")

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

func TestLoginInvalidCredentials(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{
		login: func(context.Context, *userv1.LoginRequest) (*userv1.LoginResponse, error) {
			return nil, status.Error(codes.Unauthenticated, string(exception.MsgInvalidCredentials))
		},
	}}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"login":"ada","password":"wrong"}`))
	rec := httptest.NewRecorder()
	handler.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "invalid_credentials") {
		t.Fatalf("body %s", rec.Body.String())
	}
}

func TestResendLoginCodeReturnsCooldown(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{
		resendLogin: func(context.Context, *userv1.ResendLoginCodeRequest) (*userv1.ResendLoginCodeResponse, error) {
			return &userv1.ResendLoginCodeResponse{ResendCooldownSeconds: 45}, nil
		},
	}}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login/resend", strings.NewReader(`{"challenge_id":"ch-1"}`))
	rec := httptest.NewRecorder()
	handler.ResendLoginCode(rec, req)

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"resend_after_seconds":45`) {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestAuthMFAHandlersForwardFields(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{
		beginLoginWebAuthn: func(_ context.Context, in *userv1.BeginLoginWebAuthnRequest) (*userv1.BeginLoginWebAuthnResponse, error) {
			return &userv1.BeginLoginWebAuthnResponse{WebauthnOptionsJson: `{"publicKey":{"challenge":"abc"}}`}, nil
		},
		verifyLoginMFA: func(_ context.Context, in *userv1.VerifyLoginMFARequest) (*userv1.VerifyLoginMFAResponse, error) {
			return &userv1.VerifyLoginMFAResponse{
				Session: &userv1.Session{
					AccessToken:  "access-mfa",
					ExpiresIn:    900,
					RefreshToken: refreshSecret,
					User:         &userv1.User{Id: "u1", Email: "ada@example.com", DisplayName: "Ada"},
				},
			}, nil
		},
	}}), testLog(), false, "http://localhost:8080")

	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login/mfa/webauthn/begin", strings.NewReader(`{"challenge_id":"mfa-1"}`))
	rec1 := httptest.NewRecorder()
	handler.BeginLoginWebAuthn(rec1, req1)
	if rec1.Code != http.StatusOK || !strings.Contains(rec1.Body.String(), `"webauthn_options"`) {
		t.Fatalf("begin webauthn: %d %s", rec1.Code, rec1.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login/mfa/verify", strings.NewReader(`{"challenge_id":"mfa-1","method":"totp","code":"123456"}`))
	rec2 := httptest.NewRecorder()
	handler.VerifyLoginMFA(rec2, req2)
	if rec2.Code != http.StatusOK || !strings.Contains(rec2.Body.String(), `"access-mfa"`) {
		t.Fatalf("verify mfa: %d %s", rec2.Code, rec2.Body.String())
	}
}

// merged from handler_errors_test.go
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

// merged from handler_flow_test.go
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

func TestRegisterReturnsRegisterUnavailable(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{
		register: func(context.Context, *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
			return nil, status.Error(codes.InvalidArgument, string(exception.MsgRegisterUnavailable))
		},
	}}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"email":"a@b.com","username":"ada","password":"longpassword12","display_name":"Ada","birth_date":"2000-01-01"}`))
	rec := httptest.NewRecorder()
	handler.Register(rec, req)

	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), string(exception.MsgRegisterUnavailable)) {
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

// merged from handler_forgot_refresh_test.go
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

// merged from handler_turnstile_test.go
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

func TestGoogleStartInviteOnlyBlocksSignUpReturn(t *testing.T) {
	called := false
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{
		googleStart: func(context.Context, *userv1.GoogleStartRequest) (*userv1.GoogleStartResponse, error) {
			called = true
			return nil, nil
		},
	}}), testLog(), false, "http://localhost:8080").WithInviteOnly(true)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/google?next=/sign-up", nil)
	rec := httptest.NewRecorder()
	handler.GoogleStart(rec, req)

	if called {
		t.Fatal("expected gateway to block before gRPC")
	}
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "http://localhost:8080/forbidden" {
		t.Fatalf("status %d location %q", rec.Code, rec.Header().Get("Location"))
	}
}

func TestGoogleStartInviteOnlyAllowsInviteReturn(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{
		googleStart: func(context.Context, *userv1.GoogleStartRequest) (*userv1.GoogleStartResponse, error) {
			return &userv1.GoogleStartResponse{
				AuthorizationUrl: "https://accounts.google.com/o/oauth2/auth",
				State:            "state-1",
				CodeVerifier:     "verifier-1",
			}, nil
		},
	}}), testLog(), false, "http://localhost:8080").WithInviteOnly(true)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/google?next=/invite/abcdefgh", nil)
	rec := httptest.NewRecorder()
	handler.GoogleStart(rec, req)

	if rec.Code != http.StatusFound || !strings.Contains(rec.Header().Get("Location"), "accounts.google.com") {
		t.Fatalf("status %d location %q", rec.Code, rec.Header().Get("Location"))
	}
}

func TestGoogleCallbackRegistrationClosedRedirectsForbidden(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{Auth: fakeAuth{
		googleFinish: func(context.Context, *userv1.GoogleFinishRequest) (*userv1.GoogleFinishResponse, error) {
			return nil, status.Error(codes.PermissionDenied, string(exception.MsgRegistrationClosed))
		},
	}}), testLog(), false, "http://localhost:8080").WithInviteOnly(true)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/google/callback?code=abc&state=state-1", nil)
	req.AddCookie(&http.Cookie{Name: "wichat_oauth", Value: "state-1.verifier-1"})
	rec := httptest.NewRecorder()
	handler.GoogleCallback(rec, req)

	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "http://localhost:8080/forbidden" {
		t.Fatalf("status %d location %q", rec.Code, rec.Header().Get("Location"))
	}
}

func TestVerifyRegisterRejectsMissingTurnstileWhenEnabled(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{}), testLog(), false, "http://localhost:8080", turnstileTestSecret)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register/verify", strings.NewReader(`{"challenge_id":"ch-1","code":"ABC123"}`))
	rec := httptest.NewRecorder()
	handler.VerifyRegister(rec, req)

	if rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), "captcha_failed") {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestResendRegisterRejectsMissingTurnstileWhenEnabled(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(msUserFakes{}), testLog(), false, "http://localhost:8080", turnstileTestSecret)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register/resend", strings.NewReader(`{"challenge_id":"ch-1"}`))
	rec := httptest.NewRecorder()
	handler.ResendRegisterCode(rec, req)

	if rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), "captcha_failed") {
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
