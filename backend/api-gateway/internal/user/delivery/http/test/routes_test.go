package http_test

import (
	"context"
	"net/http"
	"testing"

	userhttp "github.com/wichat/wichat/backend/api-gateway/internal/user/delivery/http"
)

type userRouteStub struct{}

func (userRouteStub) SetupStatus(http.ResponseWriter, *http.Request)              {}
func (userRouteStub) InstanceSetup(http.ResponseWriter, *http.Request)            {}
func (userRouteStub) CheckUsername(http.ResponseWriter, *http.Request)            {}
func (userRouteStub) SearchUsers(http.ResponseWriter, *http.Request)              {}
func (userRouteStub) UploadAvatar(http.ResponseWriter, *http.Request)             {}
func (userRouteStub) UpdateProfile(http.ResponseWriter, *http.Request)            {}
func (userRouteStub) UpdateUsername(http.ResponseWriter, *http.Request)           {}
func (userRouteStub) RequestEmailChange(http.ResponseWriter, *http.Request)       {}
func (userRouteStub) VerifyEmailChange(http.ResponseWriter, *http.Request)        {}
func (userRouteStub) CompleteEmailChange(http.ResponseWriter, *http.Request)      {}
func (userRouteStub) UpdatePassword(http.ResponseWriter, *http.Request)           {}
func (userRouteStub) GetMFAStatus(http.ResponseWriter, *http.Request)             {}
func (userRouteStub) BeginTOTPSetup(http.ResponseWriter, *http.Request)           {}
func (userRouteStub) ConfirmTOTPSetup(http.ResponseWriter, *http.Request)         {}
func (userRouteStub) DisableTOTP(http.ResponseWriter, *http.Request)              {}
func (userRouteStub) BeginWebAuthnRegister(http.ResponseWriter, *http.Request)    {}
func (userRouteStub) FinishWebAuthnRegister(http.ResponseWriter, *http.Request)   {}
func (userRouteStub) RenameWebAuthnCredential(http.ResponseWriter, *http.Request) {}
func (userRouteStub) DeleteWebAuthnCredential(http.ResponseWriter, *http.Request) {}
func (userRouteStub) RegenerateRecoveryCodes(http.ResponseWriter, *http.Request)  {}
func (userRouteStub) DisableAccount(http.ResponseWriter, *http.Request)           {}
func (userRouteStub) DeleteAccount(http.ResponseWriter, *http.Request)            {}
func (userRouteStub) ListDevices(http.ResponseWriter, *http.Request)              {}
func (userRouteStub) RevokeDevice(http.ResponseWriter, *http.Request)             {}
func (userRouteStub) RevokeOtherDevices(http.ResponseWriter, *http.Request)       {}

func TestRegisterMountsSetupStatus(t *testing.T) {
	mux := http.NewServeMux()
	userhttp.Register(mux, userRouteStub{})

	for _, tc := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/auth/setup-status"},
		{http.MethodGet, "/api/v1/users/me/mfa"},
		{http.MethodPost, "/api/v1/users/me/mfa/totp/begin"},
		{http.MethodPost, "/api/v1/users/me/mfa/totp/confirm"},
		{http.MethodDelete, "/api/v1/users/me/mfa/totp"},
		{http.MethodPost, "/api/v1/users/me/mfa/security-keys/begin"},
		{http.MethodPost, "/api/v1/users/me/mfa/security-keys/finish"},
		{http.MethodPatch, "/api/v1/users/me/mfa/security-keys/k-1"},
		{http.MethodDelete, "/api/v1/users/me/mfa/security-keys/k-1"},
		{http.MethodPost, "/api/v1/users/me/mfa/recovery-codes/regenerate"},
		{http.MethodPost, "/api/v1/users/me/disable"},
		{http.MethodPost, "/api/v1/users/me/delete"},
		{http.MethodGet, "/api/v1/users/me/devices"},
		{http.MethodDelete, "/api/v1/users/me/devices/sess-1"},
		{http.MethodPost, "/api/v1/users/me/devices/revoke-others"},
	} {
		req, _ := http.NewRequestWithContext(context.Background(), tc.method, tc.path, nil)
		_, pattern := mux.Handler(req)
		if pattern == "" {
			t.Fatalf("route not registered: %s %s", tc.method, tc.path)
		}
	}
}
