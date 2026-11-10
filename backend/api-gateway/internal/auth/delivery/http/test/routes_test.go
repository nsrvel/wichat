package http_test

import (
	"context"
	"net/http"
	"testing"

	authhttp "github.com/wichat/wichat/backend/api-gateway/internal/auth/delivery/http"
)

type routeStub struct{}

func (routeStub) SetupStatus(http.ResponseWriter, *http.Request)           {}
func (routeStub) Setup(http.ResponseWriter, *http.Request)                 {}
func (routeStub) Register(http.ResponseWriter, *http.Request)              {}
func (routeStub) VerifyRegister(http.ResponseWriter, *http.Request)        {}
func (routeStub) ResendRegisterCode(http.ResponseWriter, *http.Request)    {}
func (routeStub) RegisterGuest(http.ResponseWriter, *http.Request)         {}
func (routeStub) Login(http.ResponseWriter, *http.Request)                 {}
func (routeStub) VerifyLogin(http.ResponseWriter, *http.Request)           {}
func (routeStub) ResendLoginCode(http.ResponseWriter, *http.Request)       {}
func (routeStub) Refresh(http.ResponseWriter, *http.Request)               {}
func (routeStub) Logout(http.ResponseWriter, *http.Request)                {}
func (routeStub) Me(http.ResponseWriter, *http.Request)                    {}
func (routeStub) GoogleStart(http.ResponseWriter, *http.Request)           {}
func (routeStub) GoogleCallback(http.ResponseWriter, *http.Request)        {}
func (routeStub) ForgotPasswordRequest(http.ResponseWriter, *http.Request) {}
func (routeStub) ForgotPasswordVerify(http.ResponseWriter, *http.Request)  {}
func (routeStub) ForgotPasswordReset(http.ResponseWriter, *http.Request)   {}
func (routeStub) InstanceSetup(http.ResponseWriter, *http.Request)         {}
func (routeStub) ListMyGroups(http.ResponseWriter, *http.Request)          {}
func (routeStub) ReorderMyGroups(http.ResponseWriter, *http.Request)       {}
func (routeStub) ListJoinableGroups(http.ResponseWriter, *http.Request)    {}
func (routeStub) CreateGroup(http.ResponseWriter, *http.Request)           {}
func (routeStub) JoinGroup(http.ResponseWriter, *http.Request)             {}
func (routeStub) LeaveGroup(http.ResponseWriter, *http.Request)            {}
func (routeStub) DeleteGroup(http.ResponseWriter, *http.Request)           {}
func (routeStub) ListGroupChannels(http.ResponseWriter, *http.Request)     {}
func (routeStub) UploadGroupIcon(http.ResponseWriter, *http.Request)       {}
func (routeStub) ListGroupMembers(http.ResponseWriter, *http.Request)      {}
func (routeStub) SearchUsers(http.ResponseWriter, *http.Request)           {}
func (routeStub) GetGroupInviteLink(http.ResponseWriter, *http.Request)    {}
func (routeStub) SendGroupInvites(http.ResponseWriter, *http.Request)      {}
func (routeStub) PreviewGroupInvite(http.ResponseWriter, *http.Request)    {}
func (routeStub) RedeemGroupInvite(http.ResponseWriter, *http.Request)     {}

func TestRegisterMountsSetupStatus(t *testing.T) {
	mux := http.NewServeMux()
	authhttp.Register(mux, routeStub{})

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/auth/setup-status", nil)
	_, pattern := mux.Handler(req)
	if pattern == "" {
		t.Fatal("setup-status route not registered")
	}
}
