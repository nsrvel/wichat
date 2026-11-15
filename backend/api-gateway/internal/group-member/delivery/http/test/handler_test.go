package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
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
	joinGroup       func(context.Context, *userv1.JoinGroupRequest) (*userv1.JoinGroupResponse, error)
	kickGroupMember func(context.Context, *userv1.KickGroupMemberRequest) (*userv1.KickGroupMemberResponse, error)
	banGroupMembers func(context.Context, *userv1.BanGroupMembersRequest) (*userv1.BanGroupMembersResponse, error)
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

func (f fakeGroupMember) KickGroupMember(ctx context.Context, in *userv1.KickGroupMemberRequest) (*userv1.KickGroupMemberResponse, error) {
	if f.kickGroupMember != nil {
		return f.kickGroupMember(ctx, in)
	}
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (f fakeGroupMember) BanGroupMembers(ctx context.Context, in *userv1.BanGroupMembersRequest) (*userv1.BanGroupMembersResponse, error) {
	if f.banGroupMembers != nil {
		return f.banGroupMembers(ctx, in)
	}
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

func TestKickGroupMemberWithBearer(t *testing.T) {
	var gotToken, gotGroupID, gotUserID string
	clients := &microservices.Clients{User: msUser.Service{GRPC: msUser.GRPC{
		GroupMember: fakeGroupMember{
			kickGroupMember: func(_ context.Context, in *userv1.KickGroupMemberRequest) (*userv1.KickGroupMemberResponse, error) {
				gotToken = in.GetAccessToken()
				gotGroupID = in.GetGroupId()
				gotUserID = in.GetUserId()
				return &userv1.KickGroupMemberResponse{Success: true}, nil
			},
		},
	}}}
	handler := groupmemberdelivery.NewHandler(clients, logger.New(logger.Options{Service: "gateway-test", Env: "test", Level: "error"}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups/"+testGroupPublicID+"/members/user-2/kick", nil)
	req.SetPathValue("id", testGroupPublicID)
	req.SetPathValue("userId", "user-2")
	req.Header.Set("Authorization", "Bearer access-token")
	rec := httptest.NewRecorder()
	handler.KickGroupMember(rec, req)

	if rec.Code != http.StatusOK || gotToken != "access-token" || gotGroupID != testGroupPublicID || gotUserID != "user-2" {
		t.Fatalf("status %d token %q group %q user %q body %s", rec.Code, gotToken, gotGroupID, gotUserID, rec.Body.String())
	}
}

func TestBanGroupMembersWithBearer(t *testing.T) {
	var gotReason string
	var gotCount int
	var gotDeleteSeconds int32
	clients := &microservices.Clients{User: msUser.Service{GRPC: msUser.GRPC{
		GroupMember: fakeGroupMember{
			banGroupMembers: func(_ context.Context, in *userv1.BanGroupMembersRequest) (*userv1.BanGroupMembersResponse, error) {
				gotReason = in.GetReason()
				gotCount = len(in.GetUserIds())
				gotDeleteSeconds = in.GetDeleteMessageSeconds()
				return &userv1.BanGroupMembersResponse{BannedCount: 2}, nil
			},
		},
	}}}
	handler := groupmemberdelivery.NewHandler(clients, logger.New(logger.Options{Service: "gateway-test", Env: "test", Level: "error"}))

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/groups/"+testGroupPublicID+"/bans",
		strings.NewReader(`{"user_ids":["u1","u2"],"reason":" spam ","delete_message_seconds":3600}`),
	)
	req.SetPathValue("id", testGroupPublicID)
	req.Header.Set("Authorization", "Bearer access-token")
	rec := httptest.NewRecorder()
	handler.BanGroupMembers(rec, req)

	if rec.Code != http.StatusOK || gotReason != "spam" || gotCount != 2 || gotDeleteSeconds != 3600 {
		t.Fatalf("status %d reason %q count %d del %d body %s", rec.Code, gotReason, gotCount, gotDeleteSeconds, rec.Body.String())
	}
}
