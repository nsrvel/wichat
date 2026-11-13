package http_test

import (
	"context"
	"net/http"
	"testing"

	userhttp "github.com/wichat/wichat/backend/api-gateway/internal/user/delivery/http"
)

type userRouteStub struct{}

func (userRouteStub) SetupStatus(http.ResponseWriter, *http.Request)   {}
func (userRouteStub) InstanceSetup(http.ResponseWriter, *http.Request) {}
func (userRouteStub) CheckUsername(http.ResponseWriter, *http.Request) {}
func (userRouteStub) SearchUsers(http.ResponseWriter, *http.Request)   {}

func TestRegisterMountsSetupStatus(t *testing.T) {
	mux := http.NewServeMux()
	userhttp.Register(mux, userRouteStub{})

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/auth/setup-status", nil)
	_, pattern := mux.Handler(req)
	if pattern == "" {
		t.Fatal("setup-status route not registered")
	}
}
