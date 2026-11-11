package http

import (
	"net/http"
	"time"

	"github.com/wichat/wichat/backend/api-gateway/internal/middleware"
)

// Register mounts auth routes on mux.
func Register(mux *http.ServeMux, h Handler) {
	registerLimit := middleware.NewLimiter(5, time.Hour)
	registerVerifyLimit := middleware.NewLimiter(40, time.Hour)
	registerResendLimit := middleware.NewLimiter(15, time.Hour)
	groupLimit := middleware.NewLimiter(30, time.Hour)
	loginLimit := middleware.NewLimiter(10, time.Minute)
	loginVerifyLimit := middleware.NewLimiter(40, time.Hour)
	loginResendLimit := middleware.NewLimiter(15, time.Hour)
	refreshLimit := middleware.NewLimiter(30, time.Minute)
	googleLimit := middleware.NewLimiter(20, time.Minute)

	mux.HandleFunc("GET /api/v1/auth/setup-status", h.SetupStatus)
	mux.Handle("POST /api/v1/auth/instance-setup", registerLimit.Wrap(http.HandlerFunc(h.InstanceSetup)))
	mux.Handle("POST /api/v1/auth/setup", registerLimit.Wrap(http.HandlerFunc(h.Setup)))
	mux.Handle("POST /api/v1/auth/register", registerLimit.Wrap(http.HandlerFunc(h.Register)))
	mux.Handle("POST /api/v1/auth/register/verify", registerVerifyLimit.Wrap(http.HandlerFunc(h.VerifyRegister)))
	mux.Handle("POST /api/v1/auth/register/resend", registerResendLimit.Wrap(http.HandlerFunc(h.ResendRegisterCode)))
	mux.Handle("POST /api/v1/auth/login", loginLimit.Wrap(http.HandlerFunc(h.Login)))
	mux.Handle("POST /api/v1/auth/login/verify", loginVerifyLimit.Wrap(http.HandlerFunc(h.VerifyLogin)))
	mux.Handle("POST /api/v1/auth/login/resend", loginResendLimit.Wrap(http.HandlerFunc(h.ResendLoginCode)))
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

	mux.HandleFunc("GET /api/v1/groups/mine", h.ListMyGroups)
	mux.Handle("PUT /api/v1/groups/mine/order", groupLimit.Wrap(http.HandlerFunc(h.ReorderMyGroups)))
	mux.HandleFunc("GET /api/v1/groups/joinable", h.ListJoinableGroups)
	mux.Handle("POST /api/v1/groups", groupLimit.Wrap(http.HandlerFunc(h.CreateGroup)))
	mux.Handle("POST /api/v1/groups/{id}/join", groupLimit.Wrap(http.HandlerFunc(h.JoinGroup)))
	mux.Handle("POST /api/v1/groups/{id}/leave", groupLimit.Wrap(http.HandlerFunc(h.LeaveGroup)))
	mux.Handle("POST /api/v1/groups/{id}/delete", groupLimit.Wrap(http.HandlerFunc(h.DeleteGroup)))
	mux.Handle("DELETE /api/v1/groups/{id}", groupLimit.Wrap(http.HandlerFunc(h.DeleteGroup)))
	mux.HandleFunc("GET /api/v1/groups/{id}/channels", h.ListGroupChannels)
	mux.Handle("POST /api/v1/groups/{id}/icon", groupLimit.Wrap(http.HandlerFunc(h.UploadGroupIcon)))
	mux.Handle("PATCH /api/v1/groups/{id}/profile", groupLimit.Wrap(http.HandlerFunc(h.UpdateGroupProfile)))
	mux.Handle("POST /api/v1/groups/{id}/banner", groupLimit.Wrap(http.HandlerFunc(h.UploadGroupBanner)))
	inviteLimit := middleware.NewLimiter(60, time.Hour)
	searchLimit := middleware.NewLimiter(120, time.Hour)
	mux.HandleFunc("GET /api/v1/groups/{id}/members", h.ListGroupMembers)
	mux.Handle("GET /api/v1/users/search", searchLimit.Wrap(http.HandlerFunc(h.SearchUsers)))
	mux.Handle("GET /api/v1/groups/{id}/invite-link", inviteLimit.Wrap(http.HandlerFunc(h.GetGroupInviteLink)))
	mux.Handle("POST /api/v1/groups/{id}/invites", inviteLimit.Wrap(http.HandlerFunc(h.SendGroupInvites)))
	mux.Handle("GET /api/v1/invites/preview", inviteLimit.Wrap(http.HandlerFunc(h.PreviewGroupInvite)))
	mux.Handle("POST /api/v1/invites/redeem", inviteLimit.Wrap(http.HandlerFunc(h.RedeemGroupInvite)))
}
