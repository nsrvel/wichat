package authHttp_test

import (
	"encoding/json"
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

func TestOAuthSignupReturnBlocked(t *testing.T) {
	if !authHttp.OAuthSignupReturnBlocked("/sign-up") {
		t.Fatal("sign-up should block")
	}
	if authHttp.OAuthSignupReturnBlocked("/invite/abc") {
		t.Fatal("invite path should allow")
	}
	if authHttp.OAuthSignupReturnBlocked("/") {
		t.Fatal("home should allow login return")
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

func TestCategoryPermissionOverrideFromPB(t *testing.T) {
	nilBody := authHttp.CategoryPermissionOverrideFromPB(nil)
	if len(nilBody.AllowPermissions) != 0 {
		t.Fatal("expected empty allow slice")
	}
	body := authHttp.CategoryPermissionOverrideFromPB(&userv1.ChannelCategoryPermissionOverride{
		TargetType:       "role",
		TargetId:         "r1",
		AllowPermissions: []string{"read"},
		DenyPermissions:  []string{"write"},
	})
	if body.TargetID != "r1" || len(body.AllowPermissions) != 1 {
		t.Fatalf("%+v", body)
	}
}

func TestGroupAndChannelMappers(t *testing.T) {
	if authHttp.GroupFromPB(nil).ID != "" {
		t.Fatal("nil group")
	}
	g := authHttp.GroupFromPB(&userv1.GroupSummary{Id: "g1", Name: "WiChat", EffectivePermissions: nil})
	if g.ID != "g1" || len(g.EffectivePermissions) != 0 {
		t.Fatalf("%+v", g)
	}

	ch := authHttp.ChannelFromPB(&userv1.ChannelSummary{Id: "c1", Name: "general", Type: "text"})
	if ch.ID != "c1" {
		t.Fatalf("%+v", ch)
	}
	chNilPerms := authHttp.ChannelFromPB(&userv1.ChannelSummary{Id: "c2", Name: "dm", Type: "dm"})
	if chNilPerms.CallerPermissions == nil || len(chNilPerms.CallerPermissions) != 0 {
		t.Fatalf("nil proto permissions want empty slice, got %+v", chNilPerms.CallerPermissions)
	}
	if authHttp.ChannelFromPB(nil).ID != "" {
		t.Fatal("nil channel")
	}

	cat := authHttp.ChannelCategoryFromPB(&userv1.ChannelCategory{Id: "cat-1", Name: "Text"})
	if cat.ID != "cat-1" {
		t.Fatal(cat)
	}
	cats2 := authHttp.ChannelCategoriesFromPB([]*userv1.ChannelCategory{{Id: "cat-1", Name: "Text"}})
	if len(cats2) != 1 {
		t.Fatal(cats2)
	}

	chans := authHttp.ChannelsFromPB([]*userv1.ChannelSummary{{Id: "c1", Name: "general"}})
	if len(chans) != 1 {
		t.Fatal(chans)
	}

	groups := authHttp.GroupsFromPB([]*userv1.GroupSummary{{Id: "g1", Name: "G"}})
	if len(groups) != 1 {
		t.Fatal(groups)
	}
}

func TestGroupMembersAndSearchMappers(t *testing.T) {
	members := authHttp.GroupMembersFromPB([]*userv1.GroupMemberEntry{
		{UserId: "u1", DisplayName: "Ada", Roles: []*userv1.GroupMemberRoleSummary{{Id: "r1", Name: "Admin"}}},
		nil,
	})
	if len(members) != 2 || members[0].UserID != "u1" {
		t.Fatalf("%+v", members)
	}

	search := authHttp.UserSearchFromPB([]*userv1.UserSearchHit{{UserId: "u1", Username: "ada"}})
	if len(search) != 1 || search[0].UserID != "u1" {
		t.Fatalf("%+v", search)
	}
}

func TestGroupRoleMappers(t *testing.T) {
	role := authHttp.GroupRoleFromPB(&userv1.GroupRole{Id: "r1", Name: "Mod", Permissions: nil})
	if role.ID != "r1" || len(role.Permissions) != 0 {
		t.Fatalf("%+v", role)
	}
	roles2 := authHttp.GroupRolesFromPB([]*userv1.GroupRole{{Id: "r1"}})
	if len(roles2) != 1 {
		t.Fatal(roles2)
	}
}

func TestChallengeFromPBStripsEmailHint(t *testing.T) {
	body := authHttp.ChallengeFromPB(&userv1.AuthChallenge{
		ChallengeId:           "ch-1",
		EmailHint:             "a***@example.com",
		ResendCooldownSeconds: 30,
		OtpRequired:           true,
	})
	if body.EmailHint != "" {
		t.Fatalf("email_hint leaked: %q", body.EmailHint)
	}
	if body.ChallengeID != "ch-1" || body.ResendAfterSeconds != 30 {
		t.Fatalf("body: %+v", body)
	}
}

func TestWriteChallengeSetsOTPRequired(t *testing.T) {
	rec := httptest.NewRecorder()
	authHttp.WriteChallenge(rec, &userv1.AuthChallenge{ChallengeId: "ch-2", OtpRequired: false})

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var parsed struct {
		Data struct {
			OTPRequired bool   `json:"otp_required"`
			EmailHint   string `json:"email_hint"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
		t.Fatal(err)
	}
	if !parsed.Data.OTPRequired || parsed.Data.EmailHint != "" {
		t.Fatalf("data: %+v", parsed.Data)
	}
}
