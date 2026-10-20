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
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const refreshSecret = "super-secret-refresh"

func TestLoginReturnsChallengeWithoutRefreshCookie(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(fakeAuth{
		login: func(context.Context, *userv1.LoginRequest) (*userv1.LoginResponse, error) {
			return &userv1.LoginResponse{Challenge: &userv1.AuthChallenge{
				ChallengeId:           "ch-1",
				EmailHint:             "a***@example.com",
				ResendCooldownSeconds: 60,
				OtpRequired:           true,
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

	res := rec.Result()
	t.Cleanup(func() { _ = res.Body.Close() })
	for _, item := range res.Cookies() {
		if item.Name == "wichat_refresh" {
			t.Fatal("unexpected refresh cookie on challenge response")
		}
	}
}

func TestVerifyLoginSetsRefreshCookie(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(fakeAuth{
		verifyLogin: func(context.Context, *userv1.VerifyLoginRequest) (*userv1.VerifyLoginResponse, error) {
			return &userv1.VerifyLoginResponse{Session: &userv1.Session{
				AccessToken:  "access-token",
				ExpiresIn:    900,
				RefreshToken: refreshSecret,
				User:         &userv1.User{Id: "u1", Email: "ada@example.com", DisplayName: "Ada"},
			}}, nil
		},
	}), testLog(), false, "http://localhost:8080")

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

type fakeAuth struct {
	login        func(context.Context, *userv1.LoginRequest) (*userv1.LoginResponse, error)
	verifyLogin  func(context.Context, *userv1.VerifyLoginRequest) (*userv1.VerifyLoginResponse, error)
	googleFinish func(context.Context, *userv1.GoogleFinishRequest) (*userv1.GoogleFinishResponse, error)
	setupStatus  func(context.Context, *userv1.SetupStatusRequest) (*userv1.SetupStatusResponse, error)
	createGroup  func(context.Context, *userv1.CreateGroupRequest) (*userv1.CreateGroupResponse, error)
	joinGroup    func(context.Context, *userv1.JoinGroupRequest) (*userv1.JoinGroupResponse, error)
}

func testClients(authClient auth.Client) *microservices.Clients {
	return &microservices.Clients{User: msUser.Service{GRPC: msUser.GRPC{Auth: authClient}}}
}

func (f fakeAuth) SetupStatus(ctx context.Context, in *userv1.SetupStatusRequest) (*userv1.SetupStatusResponse, error) {
	if f.setupStatus != nil {
		return f.setupStatus(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) Setup(context.Context, *userv1.SetupRequest) (*userv1.SetupResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) Register(context.Context, *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) RegisterGuest(context.Context, *userv1.RegisterGuestRequest) (*userv1.RegisterGuestResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) Login(ctx context.Context, in *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	return f.login(ctx, in)
}

func (f fakeAuth) VerifyRegister(context.Context, *userv1.VerifyRegisterRequest) (*userv1.VerifyRegisterResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) VerifyLogin(ctx context.Context, in *userv1.VerifyLoginRequest) (*userv1.VerifyLoginResponse, error) {
	if f.verifyLogin != nil {
		return f.verifyLogin(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) ResendRegisterCode(context.Context, *userv1.ResendRegisterCodeRequest) (*userv1.ResendRegisterCodeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) ResendLoginCode(context.Context, *userv1.ResendLoginCodeRequest) (*userv1.ResendLoginCodeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) Refresh(context.Context, *userv1.RefreshRequest) (*userv1.RefreshResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) Logout(context.Context, *userv1.LogoutRequest) (*userv1.LogoutResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) Me(context.Context, *userv1.MeRequest) (*userv1.MeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) CompleteInstanceSetup(context.Context, *userv1.CompleteInstanceSetupRequest) (*userv1.CompleteInstanceSetupResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) GoogleStart(context.Context, *userv1.GoogleStartRequest) (*userv1.GoogleStartResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) RequestPasswordReset(context.Context, *userv1.RequestPasswordResetRequest) (*userv1.RequestPasswordResetResponse, error) {
	return &userv1.RequestPasswordResetResponse{Ok: true}, nil
}

func (f fakeAuth) VerifyPasswordReset(context.Context, *userv1.VerifyPasswordResetRequest) (*userv1.VerifyPasswordResetResponse, error) {
	return &userv1.VerifyPasswordResetResponse{ResetToken: "token-1"}, nil
}

func (f fakeAuth) CompletePasswordReset(context.Context, *userv1.CompletePasswordResetRequest) (*userv1.CompletePasswordResetResponse, error) {
	return &userv1.CompletePasswordResetResponse{Ok: true}, nil
}

func (f fakeAuth) GoogleFinish(ctx context.Context, in *userv1.GoogleFinishRequest) (*userv1.GoogleFinishResponse, error) {
	if f.googleFinish == nil {
		return nil, status.Error(codes.Unimplemented, "error.internal")
	}

	return f.googleFinish(ctx, in)
}

func (f fakeAuth) ListMyGroups(context.Context, *userv1.ListMyGroupsRequest) (*userv1.ListMyGroupsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) ListJoinableGroups(context.Context, *userv1.ListJoinableGroupsRequest) (*userv1.ListJoinableGroupsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) CreateGroup(ctx context.Context, in *userv1.CreateGroupRequest) (*userv1.CreateGroupResponse, error) {
	if f.createGroup != nil {
		return f.createGroup(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeAuth) JoinGroup(ctx context.Context, in *userv1.JoinGroupRequest) (*userv1.JoinGroupResponse, error) {
	if f.joinGroup != nil {
		return f.joinGroup(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func TestGoogleCallbackRedirectsWithRefreshCookie(t *testing.T) {
	var verifier string
	handler := authdelivery.NewHandler(testClients(fakeAuth{
		googleFinish: func(_ context.Context, in *userv1.GoogleFinishRequest) (*userv1.GoogleFinishResponse, error) {
			verifier = in.GetCodeVerifier()
			return &userv1.GoogleFinishResponse{Session: &userv1.Session{
				AccessToken:  "access-token",
				ExpiresIn:    900,
				RefreshToken: refreshSecret,
				User:         &userv1.User{Id: "u1", Email: "ada@example.com", DisplayName: "Ada"},
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
		googleFinish: func(_ context.Context, in *userv1.GoogleFinishRequest) (*userv1.GoogleFinishResponse, error) {
			return &userv1.GoogleFinishResponse{Session: &userv1.Session{
				AccessToken:  "access-token",
				ExpiresIn:    900,
				RefreshToken: refreshSecret,
				User:         &userv1.User{Id: "u1", Email: "ada@example.com", DisplayName: "Ada"},
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

func TestSetupStatusReturnsInstancePolicy(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(fakeAuth{
		setupStatus: func(context.Context, *userv1.SetupStatusRequest) (*userv1.SetupStatusResponse, error) {
			return &userv1.SetupStatusResponse{
				RegistrationMode: "invite_only",
				GuestEnabled:     true,
				UsernameEnabled:  false,
			}, nil
		},
	}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/setup-status", nil)
	rec := httptest.NewRecorder()
	handler.SetupStatus(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Data struct {
			RegistrationMode string `json:"registration_mode"`
			GuestEnabled     bool   `json:"guest_enabled"`
			UsernameEnabled  bool   `json:"username_enabled"`
			NeedsAccount     *bool  `json:"needs_account"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.RegistrationMode != "invite_only" || !body.Data.GuestEnabled || body.Data.UsernameEnabled {
		t.Fatalf("policy: %+v", body.Data)
	}
	if body.Data.NeedsAccount != nil {
		t.Fatalf("deprecated field leaked: %+v", body.Data)
	}
}

func TestCreateGroupRequiresBearer(t *testing.T) {
	handler := authdelivery.NewHandler(testClients(fakeAuth{}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", strings.NewReader(`{"name":"Acme"}`))
	rec := httptest.NewRecorder()
	handler.CreateGroup(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestCreateGroupWithBearer(t *testing.T) {
	var token string
	handler := authdelivery.NewHandler(testClients(fakeAuth{
		createGroup: func(_ context.Context, in *userv1.CreateGroupRequest) (*userv1.CreateGroupResponse, error) {
			token = in.GetAccessToken()
			return &userv1.CreateGroupResponse{Group: &userv1.GroupSummary{Id: "g1", Name: in.GetName()}}, nil
		},
	}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", strings.NewReader(`{"name":"Acme"}`))
	req.Header.Set("Authorization", "Bearer access-token")
	rec := httptest.NewRecorder()
	handler.CreateGroup(rec, req)

	if rec.Code != http.StatusCreated || token != "access-token" {
		t.Fatalf("status %d token %q body %s", rec.Code, token, rec.Body.String())
	}
	var body struct {
		Data struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.ID != "g1" || body.Data.Name != "Acme" {
		t.Fatalf("group: %+v", body.Data)
	}
}

func TestJoinGroupWithBearer(t *testing.T) {
	var gotToken, gotID string
	handler := authdelivery.NewHandler(testClients(fakeAuth{
		joinGroup: func(_ context.Context, in *userv1.JoinGroupRequest) (*userv1.JoinGroupResponse, error) {
			gotToken = in.GetAccessToken()
			gotID = in.GetGroupId()
			return &userv1.JoinGroupResponse{Ok: true}, nil
		},
	}), testLog(), false, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups/g1/join", nil)
	req.SetPathValue("id", "g1")
	req.Header.Set("Authorization", "Bearer access-token")
	rec := httptest.NewRecorder()
	handler.JoinGroup(rec, req)

	if rec.Code != http.StatusOK || gotToken != "access-token" || gotID != "g1" {
		t.Fatalf("status %d token %q id %q body %s", rec.Code, gotToken, gotID, rec.Body.String())
	}
}

func TestGoogleCallbackRejectsStateMismatch(t *testing.T) {
	called := false
	handler := authdelivery.NewHandler(testClients(fakeAuth{
		googleFinish: func(context.Context, *userv1.GoogleFinishRequest) (*userv1.GoogleFinishResponse, error) {
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
