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
func (groupRouteStub) ListJoinableGroups(http.ResponseWriter, *http.Request) {}
func (groupRouteStub) CreateGroup(http.ResponseWriter, *http.Request)        {}
func (groupRouteStub) DeleteGroup(http.ResponseWriter, *http.Request)        {}
func (groupRouteStub) UploadGroupIcon(http.ResponseWriter, *http.Request)    {}
func (groupRouteStub) UpdateGroupProfile(http.ResponseWriter, *http.Request) {}
func (groupRouteStub) UploadGroupBanner(http.ResponseWriter, *http.Request)  {}

func TestRegisterMountsCreateGroup(t *testing.T) {
	mux := http.NewServeMux()
	grouphttp.Register(mux, groupRouteStub{})

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/groups", nil)
	_, pattern := mux.Handler(req)
	if pattern == "" {
		t.Fatal("create group route not registered")
	}
}
