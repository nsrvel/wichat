package http

import (
	"net/http"
	"time"

	"github.com/wichat/wichat/backend/api-gateway/internal/middleware"
)

// Register mounts user & instance-setup routes on mux.
func Register(mux *http.ServeMux, h Handler) {
	setupLimit := middleware.NewLimiter(5, time.Hour)
	checkLimit := middleware.NewLimiter(60, time.Minute)
	searchLimit := middleware.NewLimiter(120, time.Hour)
	profileLimit := middleware.NewLimiter(30, time.Minute)

	mux.HandleFunc("GET /api/v1/auth/setup-status", h.SetupStatus)
	mux.Handle("POST /api/v1/auth/instance-setup", setupLimit.Wrap(http.HandlerFunc(h.InstanceSetup)))
	mux.Handle("GET /api/v1/users/check-username", checkLimit.Wrap(http.HandlerFunc(h.CheckUsername)))
	mux.Handle("GET /api/v1/users/search", searchLimit.Wrap(http.HandlerFunc(h.SearchUsers)))
	mux.Handle("POST /api/v1/users/me/avatar", profileLimit.Wrap(http.HandlerFunc(h.UploadAvatar)))
	mux.Handle("PATCH /api/v1/users/me/profile", profileLimit.Wrap(http.HandlerFunc(h.UpdateProfile)))
	mux.Handle("PATCH /api/v1/users/me/status-note", profileLimit.Wrap(http.HandlerFunc(h.UpdateStatusNote)))
	mux.Handle("PATCH /api/v1/users/me/preferences", profileLimit.Wrap(http.HandlerFunc(h.UpdatePreferences)))
	mux.Handle("PATCH /api/v1/users/me/username", profileLimit.Wrap(http.HandlerFunc(h.UpdateUsername)))
	mux.Handle("POST /api/v1/users/me/email/request", profileLimit.Wrap(http.HandlerFunc(h.RequestEmailChange)))
	mux.Handle("POST /api/v1/users/me/email/verify", profileLimit.Wrap(http.HandlerFunc(h.VerifyEmailChange)))
	mux.Handle("PATCH /api/v1/users/me/email", profileLimit.Wrap(http.HandlerFunc(h.CompleteEmailChange)))
	mux.Handle("PATCH /api/v1/users/me/password", profileLimit.Wrap(http.HandlerFunc(h.UpdatePassword)))
	mux.Handle("GET /api/v1/users/me/mfa", profileLimit.Wrap(http.HandlerFunc(h.GetMFAStatus)))
	mux.Handle("POST /api/v1/users/me/mfa/totp/begin", profileLimit.Wrap(http.HandlerFunc(h.BeginTOTPSetup)))
	mux.Handle("POST /api/v1/users/me/mfa/totp/confirm", profileLimit.Wrap(http.HandlerFunc(h.ConfirmTOTPSetup)))
	mux.Handle("DELETE /api/v1/users/me/mfa/totp", profileLimit.Wrap(http.HandlerFunc(h.DisableTOTP)))
	mux.Handle("POST /api/v1/users/me/mfa/security-keys/begin", profileLimit.Wrap(http.HandlerFunc(h.BeginWebAuthnRegister)))
	mux.Handle("POST /api/v1/users/me/mfa/security-keys/finish", profileLimit.Wrap(http.HandlerFunc(h.FinishWebAuthnRegister)))
	mux.Handle("PATCH /api/v1/users/me/mfa/security-keys/{id}", profileLimit.Wrap(http.HandlerFunc(h.RenameWebAuthnCredential)))
	mux.Handle("DELETE /api/v1/users/me/mfa/security-keys/{id}", profileLimit.Wrap(http.HandlerFunc(h.DeleteWebAuthnCredential)))
	mux.Handle("POST /api/v1/users/me/mfa/recovery-codes/regenerate", profileLimit.Wrap(http.HandlerFunc(h.RegenerateRecoveryCodes)))
	mux.Handle("POST /api/v1/users/me/disable", profileLimit.Wrap(http.HandlerFunc(h.DisableAccount)))
	mux.Handle("POST /api/v1/users/me/delete", profileLimit.Wrap(http.HandlerFunc(h.DeleteAccount)))
	mux.Handle("GET /api/v1/users/me/devices", profileLimit.Wrap(http.HandlerFunc(h.ListDevices)))
	mux.Handle("DELETE /api/v1/users/me/devices/{id}", profileLimit.Wrap(http.HandlerFunc(h.RevokeDevice)))
	mux.Handle("POST /api/v1/users/me/devices/revoke-others", profileLimit.Wrap(http.HandlerFunc(h.RevokeOtherDevices)))
	mux.Handle("GET /api/v1/users/me/activity-logs", profileLimit.Wrap(http.HandlerFunc(h.ListActivityLogs)))
	mux.Handle("GET /api/v1/users/me/presence", profileLimit.Wrap(http.HandlerFunc(h.GetMyPresence)))
	mux.Handle("PUT /api/v1/users/me/presence", profileLimit.Wrap(http.HandlerFunc(h.UpdateMyPresence)))
}
