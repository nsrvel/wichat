package http_test

import (
	"context"
	"net/http"
	"testing"

	grouphttp "github.com/wichat/wichat/backend/api-gateway/internal/group/delivery/http"
)

type groupRouteStub struct{}

func (groupRouteStub) ListMyGroups(http.ResponseWriter, *http.Request)       {}
func (groupRouteStub) ReorderMyGroups(http.ResponseWriter, *http.Request)    {}
func (groupRouteStub) CreateGroup(http.ResponseWriter, *http.Request)        {}
func (groupRouteStub) DeleteGroup(http.ResponseWriter, *http.Request)        {}
func (groupRouteStub) UploadGroupIcon(http.ResponseWriter, *http.Request)    {}
func (groupRouteStub) UpdateGroupProfile(http.ResponseWriter, *http.Request) {}
func (groupRouteStub) UploadGroupBanner(http.ResponseWriter, *http.Request)  {}
func (groupRouteStub) ListGroupAuditLogs(http.ResponseWriter, *http.Request) {}

func TestRegisterMountsCreateGroup(t *testing.T) {
	mux := http.NewServeMux()
	grouphttp.Register(mux, groupRouteStub{})

	for _, tc := range []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/groups"},
		{http.MethodGet, "/api/v1/groups/507f1f77bcf86cd799439011/audit-logs"},
	} {
		req, _ := http.NewRequestWithContext(context.Background(), tc.method, tc.path, nil)
		_, pattern := mux.Handler(req)
		if pattern == "" {
			t.Fatalf("%s %s route not registered", tc.method, tc.path)
		}
	}
}
