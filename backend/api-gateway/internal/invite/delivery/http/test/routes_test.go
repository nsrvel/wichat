package http_test

import (
	"context"
	"net/http"
	"testing"

	invitehttp "github.com/wichat/wichat/backend/api-gateway/internal/invite/delivery/http"
)

type inviteRouteStub struct{}

func (inviteRouteStub) GetGroupInviteLink(http.ResponseWriter, *http.Request) {}
func (inviteRouteStub) SendGroupInvites(http.ResponseWriter, *http.Request)   {}
func (inviteRouteStub) PreviewGroupInvite(http.ResponseWriter, *http.Request) {}
func (inviteRouteStub) RedeemGroupInvite(http.ResponseWriter, *http.Request)  {}

func TestRegisterMountsInviteRoutes(t *testing.T) {
	mux := http.NewServeMux()
	invitehttp.Register(mux, inviteRouteStub{})

	cases := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/groups/abc/invite-link"},
		{http.MethodPost, "/api/v1/groups/abc/invites"},
		{http.MethodGet, "/api/v1/invites/preview"},
		{http.MethodPost, "/api/v1/invites/redeem"},
	}
	for _, tc := range cases {
		req, _ := http.NewRequestWithContext(context.Background(), tc.method, tc.path, nil)
		_, pattern := mux.Handler(req)
		if pattern == "" {
			t.Fatalf("route not registered: %s %s", tc.method, tc.path)
		}
	}
}
