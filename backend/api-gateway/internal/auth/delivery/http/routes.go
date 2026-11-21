package http

import (
	"net/http"
	"time"

	"github.com/wichat/wichat/backend/api-gateway/internal/middleware"
)

// Register mounts auth routes on mux. development lifts the register POST cap for local E2E.
func Register(mux *http.ServeMux, h Handler, development bool) {
	registerCap := 5
	if development {
		// ponytail: dev serial register-flow e2e exceeds 5/hour from one IP; production stays 5.
		registerCap = 40
	}
	registerLimit := middleware.NewLimiter(registerCap, time.Hour)
	registerVerifyLimit := middleware.NewLimiter(40, time.Hour)
	registerResendLimit := middleware.NewLimiter(15, time.Hour)
	loginLimit := middleware.NewLimiter(10, time.Minute)
	loginVerifyLimit := middleware.NewLimiter(40, time.Hour)
	loginResendLimit := middleware.NewLimiter(15, time.Hour)
	refreshLimit := middleware.NewLimiter(30, time.Minute)
	googleLimit := middleware.NewLimiter(20, time.Minute)

	mux.Handle("POST /api/v1/auth/setup", registerLimit.Wrap(http.HandlerFunc(h.Setup)))
	mux.Handle("POST /api/v1/auth/register", registerLimit.Wrap(http.HandlerFunc(h.Register)))
	mux.Handle("POST /api/v1/auth/register/verify", registerVerifyLimit.Wrap(http.HandlerFunc(h.VerifyRegister)))
	mux.Handle("POST /api/v1/auth/register/resend", registerResendLimit.Wrap(http.HandlerFunc(h.ResendRegisterCode)))
	mux.Handle("POST /api/v1/auth/login", loginLimit.Wrap(http.HandlerFunc(h.Login)))
	mux.Handle("POST /api/v1/auth/login/verify", loginVerifyLimit.Wrap(http.HandlerFunc(h.VerifyLogin)))
	mux.Handle("POST /api/v1/auth/login/resend", loginResendLimit.Wrap(http.HandlerFunc(h.ResendLoginCode)))
	mux.Handle("POST /api/v1/auth/login/mfa/verify", loginVerifyLimit.Wrap(http.HandlerFunc(h.VerifyLoginMFA)))
	mux.Handle("POST /api/v1/auth/login/mfa/webauthn/begin", loginLimit.Wrap(http.HandlerFunc(h.BeginLoginWebAuthn)))
	googleStart := googleLimit.Wrap(http.HandlerFunc(h.GoogleStart))
	googleCallback := googleLimit.Wrap(http.HandlerFunc(h.GoogleCallback))
	mux.Handle("GET /api/v1/auth/google", googleStart)
	mux.Handle("GET /api/v1/auth/google/callback", googleCallback)
	mux.Handle("GET /api/auth/callback/google", googleCallback)
	mux.Handle("POST /api/v1/auth/refresh", refreshLimit.Wrap(http.HandlerFunc(h.Refresh)))
	mux.HandleFunc("POST /api/v1/auth/logout", h.Logout)
	mux.HandleFunc("GET /api/v1/auth/me", h.Me)
	// Separate buckets: one shared 5/hour cap blocked the reset step after a normal send→verify flow.
	forgotRequestLimit := middleware.NewLimiter(15, time.Hour)
	forgotVerifyLimit := middleware.NewLimiter(40, time.Hour)
	forgotResetLimit := middleware.NewLimiter(20, time.Hour)
	mux.Handle("POST /api/v1/auth/forgot-password", forgotRequestLimit.Wrap(http.HandlerFunc(h.ForgotPasswordRequest)))
	mux.Handle("POST /api/v1/auth/forgot-password/verify", forgotVerifyLimit.Wrap(http.HandlerFunc(h.ForgotPasswordVerify)))
	mux.Handle("POST /api/v1/auth/forgot-password/reset", forgotResetLimit.Wrap(http.HandlerFunc(h.ForgotPasswordReset)))
}
