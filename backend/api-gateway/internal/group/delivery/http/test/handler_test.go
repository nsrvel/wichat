package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	groupdelivery "github.com/wichat/wichat/backend/api-gateway/internal/group/delivery/http"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
	msUser "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const testGroupPublicID = "507f1f77bcf86cd799439011"

type fakeGroup struct {
	createGroup func(context.Context, *userv1.CreateGroupRequest) (*userv1.CreateGroupResponse, error)
}

func (fakeGroup) ListMyGroups(context.Context, *userv1.ListMyGroupsRequest) (*userv1.ListMyGroupsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (fakeGroup) ListJoinableGroups(context.Context, *userv1.ListJoinableGroupsRequest) (*userv1.ListJoinableGroupsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeGroup) CreateGroup(ctx context.Context, in *userv1.CreateGroupRequest) (*userv1.CreateGroupResponse, error) {
	if f.createGroup != nil {
		return f.createGroup(ctx, in)
	}

	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (fakeGroup) DeleteGroup(context.Context, *userv1.DeleteGroupRequest) (*userv1.DeleteGroupResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (fakeGroup) UploadGroupIcon(context.Context, *userv1.UploadGroupIconRequest) (*userv1.UploadGroupIconResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (fakeGroup) UpdateGroupProfile(context.Context, *userv1.UpdateGroupProfileRequest) (*userv1.UpdateGroupProfileResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (fakeGroup) UploadGroupBanner(context.Context, *userv1.UploadGroupBannerRequest) (*userv1.UploadGroupBannerResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (fakeGroup) ReorderMyGroups(context.Context, *userv1.ReorderMyGroupsRequest) (*userv1.ReorderMyGroupsResponse, error) {
	return &userv1.ReorderMyGroupsResponse{Ok: true}, nil
}

func testGroupHandler(fg fakeGroup) groupdelivery.Handler {
	clients := &microservices.Clients{User: msUser.Service{GRPC: msUser.GRPC{Group: fg}}}
	return groupdelivery.NewHandler(clients, logger.New(logger.Options{Service: "gateway-test", Env: "test", Level: "error"}))
}

func TestCreateGroupRequiresBearer(t *testing.T) {
	handler := testGroupHandler(fakeGroup{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", strings.NewReader(`{"name":"Acme"}`))
	rec := httptest.NewRecorder()
	handler.CreateGroup(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestCreateGroupRejectsEmptyName(t *testing.T) {
	handler := testGroupHandler(fakeGroup{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", strings.NewReader(`{"name":"   "}`))
	req.Header.Set("Authorization", "Bearer access-token")
	rec := httptest.NewRecorder()
	handler.CreateGroup(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestCreateGroupWithBearer(t *testing.T) {
	var token string
	handler := testGroupHandler(fakeGroup{
		createGroup: func(_ context.Context, in *userv1.CreateGroupRequest) (*userv1.CreateGroupResponse, error) {
			token = in.GetAccessToken()
			return &userv1.CreateGroupResponse{
				Group: &userv1.GroupSummary{Id: testGroupPublicID, Name: in.GetName()},
			}, nil
		},
	})

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
	if body.Data.ID != testGroupPublicID || body.Data.Name != "Acme" {
		t.Fatalf("group: %+v", body.Data)
	}
}
