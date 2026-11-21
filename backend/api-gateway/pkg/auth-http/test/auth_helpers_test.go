package authHttp_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDecodeJSONInvalidBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`not-json`))
	rec := httptest.NewRecorder()
	var dst struct{}
	err := authHttp.DecodeJSON(rec, req, &dst)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBearerVariants(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if authHttp.Bearer(req) != "" {
		t.Fatal("expected empty")
	}
	req.Header.Set("Authorization", "Bearer tok-1")
	if authHttp.Bearer(req) != "tok-1" {
		t.Fatalf("got %q", authHttp.Bearer(req))
	}
}

func TestWriteSessionRejectsIncomplete(t *testing.T) {
	rec := httptest.NewRecorder()
	authHttp.WriteSession(rec, &userv1.Session{AccessToken: "only-access"}, false)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestWritePasswordAuthSessionPath(t *testing.T) {
	rec := httptest.NewRecorder()
	authHttp.WritePasswordAuth(rec, &userv1.Session{
		AccessToken:  "access",
		RefreshToken: "refresh",
		ExpiresIn:    60,
		User:         &userv1.User{Id: "u1"},
	}, nil, false)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"access"`) {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestWriteChallengeRejectsEmptyID(t *testing.T) {
	rec := httptest.NewRecorder()
	authHttp.WriteChallenge(rec, &userv1.AuthChallenge{})
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestChallengeFromPBWithMFAAndWebAuthn(t *testing.T) {
	body := authHttp.ChallengeFromPB(&userv1.AuthChallenge{
		ChallengeId:           "mfa-1",
		MfaRequired:           true,
		MfaMethods:            []string{"totp"},
		WebauthnOptionsJson:   `{"publicKey":{}}`,
		ResendCooldownSeconds: 10,
	})
	if !body.MFARequired || body.ChallengeID != "mfa-1" || len(body.MFAMethods) != 1 {
		t.Fatalf("%+v", body)
	}
	if len(body.WebAuthnOptions) == 0 {
		t.Fatal("expected webauthn options")
	}
}

func TestSecurityKeyFromPBNil(t *testing.T) {
	if authHttp.SecurityKeyFromPB(nil).ID != "" {
		t.Fatal("expected empty")
	}
}

func TestGRPCMessageCode(t *testing.T) {
	st := status.Error(codes.Unauthenticated, string(exception.MsgGoogleDenied))
	if code := authHttp.GRPCMessageCode(st); code != string(exception.MsgGoogleDenied) {
		t.Fatalf("got %q", code)
	}
	if code := authHttp.GRPCMessageCode(errors.New("plain")); code != string(exception.MsgInternal) {
		t.Fatalf("got %q", code)
	}
}

func TestSafeReturnPath(t *testing.T) {
	if authHttp.SafeReturnPath("") != "/" {
		t.Fatal("empty")
	}
	if authHttp.SafeReturnPath("//evil") != "/" {
		t.Fatal("open redirect")
	}
	if authHttp.SafeReturnPath("/invite/abc") != "/invite/abc" {
		t.Fatal("valid path")
	}
}

func TestWriteOutboundWithLogger(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	log := logger.New(logger.Options{Service: "test", Env: "test", Level: "error"})
	authHttp.WriteOutbound(rec, req, log, exception.InvalidInput(exception.MsgInvalidInput, nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rec.Code)
	}
}
