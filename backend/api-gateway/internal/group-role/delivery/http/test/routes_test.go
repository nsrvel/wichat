package http_test

import (
	"context"
	"net/http"
	"testing"

	grouprolehttp "github.com/wichat/wichat/backend/api-gateway/internal/group-role/delivery/http"
)

type groupRoleRouteStub struct{}

func (groupRoleRouteStub) ListGroupRoles(http.ResponseWriter, *http.Request)        {}
func (groupRoleRouteStub) CreateGroupRole(http.ResponseWriter, *http.Request)       {}
func (groupRoleRouteStub) UpdateGroupRole(http.ResponseWriter, *http.Request)       {}
func (groupRoleRouteStub) DuplicateGroupRole(http.ResponseWriter, *http.Request)    {}
func (groupRoleRouteStub) DeleteGroupRole(http.ResponseWriter, *http.Request)       {}
func (groupRoleRouteStub) ListGroupRoleMembers(http.ResponseWriter, *http.Request)  {}
func (groupRoleRouteStub) AddGroupRoleMembers(http.ResponseWriter, *http.Request)   {}
func (groupRoleRouteStub) RemoveGroupRoleMember(http.ResponseWriter, *http.Request) {}
func (groupRoleRouteStub) UpdateGroupHideRoles(http.ResponseWriter, *http.Request)  {}

func TestRegisterMountsGroupRoleRoutes(t *testing.T) {
	mux := http.NewServeMux()
	grouprolehttp.Register(mux, groupRoleRouteStub{})

	cases := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/groups/g1/roles"},
		{http.MethodPost, "/api/v1/groups/g1/roles"},
		{http.MethodPatch, "/api/v1/groups/g1/roles/settings"},
		{http.MethodPatch, "/api/v1/groups/g1/roles/r1"},
		{http.MethodPost, "/api/v1/groups/g1/roles/r1/duplicate"},
		{http.MethodDelete, "/api/v1/groups/g1/roles/r1"},
		{http.MethodGet, "/api/v1/groups/g1/roles/r1/members"},
		{http.MethodPost, "/api/v1/groups/g1/roles/r1/members"},
		{http.MethodDelete, "/api/v1/groups/g1/roles/r1/members/u1"},
	}

	for _, tc := range cases {
		req, _ := http.NewRequestWithContext(context.Background(), tc.method, tc.path, nil)
		_, pattern := mux.Handler(req)
		if pattern == "" {
			t.Fatalf("route %s %s not registered", tc.method, tc.path)
		}
	}
}
