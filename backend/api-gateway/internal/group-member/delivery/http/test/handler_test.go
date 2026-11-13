package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	groupmemberdelivery "github.com/wichat/wichat/backend/api-gateway/internal/group-member/delivery/http"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
	msUser "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const testGroupPublicID = "507f1f77bcf86cd799439011"

type fakeGroupMember struct {
	joinGroup func(context.Context, *userv1.JoinGroupRequest) (*userv1.JoinGroupResponse, error)
}

func (f fakeGroupMember) JoinGroup(ctx context.Context, in *userv1.JoinGroupRequest) (*userv1.JoinGroupResponse, error) {
	if f.joinGroup != nil {
		return f.joinGroup(ctx, in)
	}

	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (fakeGroupMember) LeaveGroup(context.Context, *userv1.LeaveGroupRequest) (*userv1.LeaveGroupResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (fakeGroupMember) ListGroupMembers(context.Context, *userv1.ListGroupMembersRequest) (*userv1.ListGroupMembersResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (fakeGroupMember) KickGroupMember(context.Context, *userv1.KickGroupMemberRequest) (*userv1.KickGroupMemberResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (fakeGroupMember) BanGroupMembers(context.Context, *userv1.BanGroupMembersRequest) (*userv1.BanGroupMembersResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func TestJoinGroupWithBearer(t *testing.T) {
	var gotToken, gotID string
	clients := &microservices.Clients{User: msUser.Service{GRPC: msUser.GRPC{
		GroupMember: fakeGroupMember{
			joinGroup: func(_ context.Context, in *userv1.JoinGroupRequest) (*userv1.JoinGroupResponse, error) {
				gotToken = in.GetAccessToken()
				gotID = in.GetGroupId()
				return &userv1.JoinGroupResponse{Ok: true}, nil
			},
		},
	}}}
	handler := groupmemberdelivery.NewHandler(clients, logger.New(logger.Options{Service: "gateway-test", Env: "test", Level: "error"}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups/"+testGroupPublicID+"/join", nil)
	req.SetPathValue("id", testGroupPublicID)
	req.Header.Set("Authorization", "Bearer access-token")
	rec := httptest.NewRecorder()
	handler.JoinGroup(rec, req)

	if rec.Code != http.StatusOK || gotToken != "access-token" || gotID != testGroupPublicID {
		t.Fatalf("status %d token %q id %q body %s", rec.Code, gotToken, gotID, rec.Body.String())
	}
}
