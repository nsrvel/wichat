package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	invitedelivery "github.com/wichat/wichat/backend/api-gateway/internal/invite/delivery/http"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
	msUser "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const turnstileTestSecret = "test-turnstile-secret"

type fakeInviteGRPC struct {
	preview func(context.Context, *userv1.PreviewGroupInviteRequest) (*userv1.PreviewGroupInviteResponse, error)
	redeem  func(context.Context, *userv1.RedeemGroupInviteRequest) (*userv1.RedeemGroupInviteResponse, error)
	link    func(context.Context, *userv1.GetGroupInviteLinkRequest) (*userv1.GetGroupInviteLinkResponse, error)
	send    func(context.Context, *userv1.SendGroupInvitesRequest) (*userv1.SendGroupInvitesResponse, error)
}

func (f fakeInviteGRPC) GetGroupInviteLink(ctx context.Context, in *userv1.GetGroupInviteLinkRequest) (*userv1.GetGroupInviteLinkResponse, error) {
	if f.link != nil {
		return f.link(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeInviteGRPC) SendGroupInvites(ctx context.Context, in *userv1.SendGroupInvitesRequest) (*userv1.SendGroupInvitesResponse, error) {
	if f.send != nil {
		return f.send(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeInviteGRPC) PreviewGroupInvite(ctx context.Context, in *userv1.PreviewGroupInviteRequest) (*userv1.PreviewGroupInviteResponse, error) {
	if f.preview != nil {
		return f.preview(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeInviteGRPC) RedeemGroupInvite(ctx context.Context, in *userv1.RedeemGroupInviteRequest) (*userv1.RedeemGroupInviteResponse, error) {
	if f.redeem != nil {
		return f.redeem(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func testInviteHandler(fi fakeInviteGRPC, turnstileSecret ...string) invitedelivery.Handler {
	clients := &microservices.Clients{User: msUser.Service{GRPC: msUser.GRPC{Invite: fi}}}
	return invitedelivery.NewHandler(clients, logger.New(logger.Options{Service: "gateway-test", Env: "test", Level: "error"}), "http://localhost:3000", turnstileSecret...)
}

func TestPreviewInviteRequiresTokenQuery(t *testing.T) {
	handler := testInviteHandler(fakeInviteGRPC{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/invites/preview", nil)
	rec := httptest.NewRecorder()
	handler.PreviewGroupInvite(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestPreviewInviteReturnsSoftInvalid(t *testing.T) {
	handler := testInviteHandler(fakeInviteGRPC{
		preview: func(context.Context, *userv1.PreviewGroupInviteRequest) (*userv1.PreviewGroupInviteResponse, error) {
			return &userv1.PreviewGroupInviteResponse{Valid: false, Reason: "not_found"}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/invites/preview?invite=abc", nil)
	rec := httptest.NewRecorder()
	handler.PreviewGroupInvite(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Data struct {
			Valid  bool   `json:"valid"`
			Reason string `json:"reason"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.Valid || body.Data.Reason != "not_found" {
		t.Fatalf("%+v", body.Data)
	}
}

func TestRedeemInviteRequiresBearer(t *testing.T) {
	handler := testInviteHandler(fakeInviteGRPC{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/invites/redeem", strings.NewReader(`{"token":"tok"}`))
	rec := httptest.NewRecorder()
	handler.RedeemGroupInvite(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestRedeemInviteForwardsToken(t *testing.T) {
	const groupID = "507f1f77bcf86cd799439011"
	const token = "share-link-token"
	handler := testInviteHandler(fakeInviteGRPC{
		redeem: func(_ context.Context, in *userv1.RedeemGroupInviteRequest) (*userv1.RedeemGroupInviteResponse, error) {
			if in.GetAccessToken() != "access-1" || in.GetToken() != token {
				t.Fatalf("unexpected redeem in: %+v", in)
			}
			return &userv1.RedeemGroupInviteResponse{GroupId: groupID}, nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/invites/redeem", strings.NewReader(`{"token":"share-link-token"}`))
	req.Header.Set("Authorization", "Bearer access-1")
	rec := httptest.NewRecorder()
	handler.RedeemGroupInvite(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestGetGroupInviteLinkBuildsURL(t *testing.T) {
	handler := testInviteHandler(fakeInviteGRPC{
		link: func(_ context.Context, in *userv1.GetGroupInviteLinkRequest) (*userv1.GetGroupInviteLinkResponse, error) {
			if in.GetGroupId() != "507f1f77bcf86cd799439011" {
				t.Fatalf("group id %+v", in)
			}
			return &userv1.GetGroupInviteLinkResponse{
				Token:     "abc123",
				ExpiresAt: "2030-01-01T00:00:00Z",
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups/507f1f77bcf86cd799439011/invite-link", nil)
	req.SetPathValue("id", "507f1f77bcf86cd799439011")
	req.Header.Set("Authorization", "Bearer access-1")
	rec := httptest.NewRecorder()
	handler.GetGroupInviteLink(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Data struct {
			Token string `json:"token"`
			URL   string `json:"url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.Token != "abc123" || !strings.Contains(body.Data.URL, "/invite/abc123") {
		t.Fatalf("%+v", body.Data)
	}
}

func TestSendGroupInvitesRequiresBearer(t *testing.T) {
	handler := testInviteHandler(fakeInviteGRPC{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups/507f1f77bcf86cd799439011/invites", strings.NewReader(`{"user_ids":["u1"]}`))
	req.SetPathValue("id", "507f1f77bcf86cd799439011")
	rec := httptest.NewRecorder()
	handler.SendGroupInvites(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestPreviewInviteRejectsMissingTurnstileWhenEnabled(t *testing.T) {
	handler := testInviteHandler(fakeInviteGRPC{}, turnstileTestSecret)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/invites/preview?invite=abc", nil)
	rec := httptest.NewRecorder()
	handler.PreviewGroupInvite(rec, req)

	if rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), "captcha_failed") {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestRedeemInviteRejectsMissingTurnstileWhenEnabled(t *testing.T) {
	handler := testInviteHandler(fakeInviteGRPC{}, turnstileTestSecret)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/invites/redeem", strings.NewReader(`{"token":"tok"}`))
	req.Header.Set("Authorization", "Bearer access-1")
	rec := httptest.NewRecorder()
	handler.RedeemGroupInvite(rec, req)

	if rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), "captcha_failed") {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestSendGroupInvitesRejectsMissingTurnstileWhenEnabled(t *testing.T) {
	handler := testInviteHandler(fakeInviteGRPC{}, turnstileTestSecret)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups/507f1f77bcf86cd799439011/invites", strings.NewReader(`{"user_ids":["u1"]}`))
	req.SetPathValue("id", "507f1f77bcf86cd799439011")
	req.Header.Set("Authorization", "Bearer access-1")
	rec := httptest.NewRecorder()
	handler.SendGroupInvites(rec, req)

	if rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), "captcha_failed") {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestGetGroupInviteLinkRejectsMissingTurnstileWhenEnabled(t *testing.T) {
	handler := testInviteHandler(fakeInviteGRPC{}, turnstileTestSecret)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups/507f1f77bcf86cd799439011/invite-link", nil)
	req.SetPathValue("id", "507f1f77bcf86cd799439011")
	req.Header.Set("Authorization", "Bearer access-1")
	rec := httptest.NewRecorder()
	handler.GetGroupInviteLink(rec, req)

	if rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), "captcha_failed") {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}
