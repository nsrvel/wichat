package http_test

import (
	"context"
	"net/http"
	"testing"

	groupmemberhttp "github.com/wichat/wichat/backend/api-gateway/internal/group-member/delivery/http"
)

type groupMemberRouteStub struct{}

func (groupMemberRouteStub) JoinGroup(http.ResponseWriter, *http.Request)          {}
func (groupMemberRouteStub) LeaveGroup(http.ResponseWriter, *http.Request)         {}
func (groupMemberRouteStub) ListGroupMembers(http.ResponseWriter, *http.Request)   {}
func (groupMemberRouteStub) KickGroupMember(http.ResponseWriter, *http.Request)    {}
func (groupMemberRouteStub) BanGroupMembers(http.ResponseWriter, *http.Request)    {}
func (groupMemberRouteStub) ListGroupBans(http.ResponseWriter, *http.Request)      {}
func (groupMemberRouteStub) UnbanGroupMember(http.ResponseWriter, *http.Request)   {}
func (groupMemberRouteStub) TimeoutGroupMember(http.ResponseWriter, *http.Request) {}

func TestRegisterMountsGroupMemberRoutes(t *testing.T) {
	mux := http.NewServeMux()
	groupmemberhttp.Register(mux, groupMemberRouteStub{})

	cases := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/groups/abc/join"},
		{http.MethodPost, "/api/v1/groups/abc/leave"},
		{http.MethodGet, "/api/v1/groups/abc/members"},
		{http.MethodPost, "/api/v1/groups/abc/members/u1/kick"},
		{http.MethodPut, "/api/v1/groups/abc/members/u1/timeout"},
		{http.MethodGet, "/api/v1/groups/abc/bans"},
		{http.MethodPost, "/api/v1/groups/abc/bans"},
		{http.MethodDelete, "/api/v1/groups/abc/bans/u1"},
	}
	for _, tc := range cases {
		req, _ := http.NewRequestWithContext(context.Background(), tc.method, tc.path, nil)
		_, pattern := mux.Handler(req)
		if pattern == "" {
			t.Fatalf("route not registered: %s %s", tc.method, tc.path)
		}
	}
}
