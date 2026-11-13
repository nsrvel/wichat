package http_test

import (
	"context"
	"net/http"
	"testing"

	authhttp "github.com/wichat/wichat/backend/api-gateway/internal/auth/delivery/http"
)

type routeStub struct{}

func (routeStub) Setup(http.ResponseWriter, *http.Request)                 {}
func (routeStub) Register(http.ResponseWriter, *http.Request)              {}
func (routeStub) VerifyRegister(http.ResponseWriter, *http.Request)        {}
func (routeStub) ResendRegisterCode(http.ResponseWriter, *http.Request)    {}
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

func TestRegisterMountsLogin(t *testing.T) {
	mux := http.NewServeMux()
	authhttp.Register(mux, routeStub{})

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/auth/login", nil)
	_, pattern := mux.Handler(req)
	if pattern == "" {
		t.Fatal("login route not registered")
	}
}
