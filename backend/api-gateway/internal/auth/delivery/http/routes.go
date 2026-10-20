package http

import (
	"net/http"
	"time"

	"github.com/wichat/wichat/backend/api-gateway/internal/middleware"
)

// Register mounts auth routes on mux.
func Register(mux *http.ServeMux, h Handler) {
	setupLimit := middleware.NewLimiter(5, time.Hour)
	groupLimit := middleware.NewLimiter(30, time.Hour)
	loginLimit := middleware.NewLimiter(10, time.Minute)
	refreshLimit := middleware.NewLimiter(30, time.Minute)
	googleLimit := middleware.NewLimiter(20, time.Minute)

	mux.HandleFunc("GET /api/v1/auth/setup-status", h.SetupStatus)
	mux.Handle("POST /api/v1/auth/instance-setup", setupLimit.Wrap(http.HandlerFunc(h.InstanceSetup)))
	mux.Handle("POST /api/v1/auth/setup", setupLimit.Wrap(http.HandlerFunc(h.Setup)))
	mux.Handle("POST /api/v1/auth/register", setupLimit.Wrap(http.HandlerFunc(h.Register)))
	mux.Handle("POST /api/v1/auth/register/verify", setupLimit.Wrap(http.HandlerFunc(h.VerifyRegister)))
	mux.Handle("POST /api/v1/auth/register/resend", setupLimit.Wrap(http.HandlerFunc(h.ResendRegisterCode)))
	mux.Handle("POST /api/v1/auth/register/guest", setupLimit.Wrap(http.HandlerFunc(h.RegisterGuest)))
	mux.Handle("POST /api/v1/auth/login", loginLimit.Wrap(http.HandlerFunc(h.Login)))
	mux.Handle("POST /api/v1/auth/login/verify", loginLimit.Wrap(http.HandlerFunc(h.VerifyLogin)))
	mux.Handle("POST /api/v1/auth/login/resend", loginLimit.Wrap(http.HandlerFunc(h.ResendLoginCode)))
	googleStart := googleLimit.Wrap(http.HandlerFunc(h.GoogleStart))
	googleCallback := googleLimit.Wrap(http.HandlerFunc(h.GoogleCallback))
	mux.Handle("GET /api/v1/auth/google", googleStart)
	mux.Handle("GET /api/v1/auth/google/callback", googleCallback)
	mux.Handle("GET /api/auth/callback/google", googleCallback)
	mux.Handle("POST /api/v1/auth/refresh", refreshLimit.Wrap(http.HandlerFunc(h.Refresh)))
	mux.HandleFunc("POST /api/v1/auth/logout", h.Logout)
	mux.HandleFunc("GET /api/v1/auth/me", h.Me)
	forgotLimit := middleware.NewLimiter(5, time.Hour)
	mux.Handle("POST /api/v1/auth/forgot-password", forgotLimit.Wrap(http.HandlerFunc(h.ForgotPasswordRequest)))
	mux.Handle("POST /api/v1/auth/forgot-password/verify", forgotLimit.Wrap(http.HandlerFunc(h.ForgotPasswordVerify)))
	mux.Handle("POST /api/v1/auth/forgot-password/reset", forgotLimit.Wrap(http.HandlerFunc(h.ForgotPasswordReset)))

	mux.HandleFunc("GET /api/v1/groups/mine", h.ListMyGroups)
	mux.HandleFunc("GET /api/v1/groups/joinable", h.ListJoinableGroups)
	mux.Handle("POST /api/v1/groups", groupLimit.Wrap(http.HandlerFunc(h.CreateGroup)))
	mux.Handle("POST /api/v1/groups/{id}/join", groupLimit.Wrap(http.HandlerFunc(h.JoinGroup)))
}
