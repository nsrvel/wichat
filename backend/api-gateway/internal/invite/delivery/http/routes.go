package http

import (
	"net/http"
	"time"

	"github.com/wichat/wichat/backend/api-gateway/internal/middleware"
)

// Register mounts invite routes on mux.
func Register(mux *http.ServeMux, h Handler) {
	inviteLimit := middleware.NewLimiter(60, time.Hour)

	mux.Handle("GET /api/v1/groups/{id}/invite-link", inviteLimit.Wrap(http.HandlerFunc(h.GetGroupInviteLink)))
	mux.Handle("POST /api/v1/groups/{id}/invites", inviteLimit.Wrap(http.HandlerFunc(h.SendGroupInvites)))
	mux.Handle("GET /api/v1/invites/preview", inviteLimit.Wrap(http.HandlerFunc(h.PreviewGroupInvite)))
	mux.Handle("POST /api/v1/invites/redeem", inviteLimit.Wrap(http.HandlerFunc(h.RedeemGroupInvite)))
}
