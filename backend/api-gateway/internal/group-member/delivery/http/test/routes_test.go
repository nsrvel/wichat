package http_test

import (
	"context"
	"net/http"
	"testing"

	groupmemberhttp "github.com/wichat/wichat/backend/api-gateway/internal/group-member/delivery/http"
)

type groupMemberRouteStub struct{}

func (groupMemberRouteStub) JoinGroup(http.ResponseWriter, *http.Request)        {}
func (groupMemberRouteStub) LeaveGroup(http.ResponseWriter, *http.Request)       {}
func (groupMemberRouteStub) ListGroupMembers(http.ResponseWriter, *http.Request) {}

func TestRegisterMountsJoinGroup(t *testing.T) {
	mux := http.NewServeMux()
	groupmemberhttp.Register(mux, groupMemberRouteStub{})

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/groups/abc/join", nil)
	_, pattern := mux.Handler(req)
	if pattern == "" {
		t.Fatal("join group route not registered")
	}
}
