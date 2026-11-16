package http

import (
	"net/http"
	"time"

	"github.com/wichat/wichat/backend/api-gateway/internal/middleware"
)

// Register mounts group-member routes on mux.
func Register(mux *http.ServeMux, h Handler) {
	groupLimit := middleware.NewLimiter(30, time.Hour)

	mux.Handle("POST /api/v1/groups/{id}/join", groupLimit.Wrap(http.HandlerFunc(h.JoinGroup)))
	mux.Handle("POST /api/v1/groups/{id}/leave", groupLimit.Wrap(http.HandlerFunc(h.LeaveGroup)))
	mux.HandleFunc("GET /api/v1/groups/{id}/members", h.ListGroupMembers)
	mux.Handle("POST /api/v1/groups/{id}/members/{userId}/kick", groupLimit.Wrap(http.HandlerFunc(h.KickGroupMember)))
	mux.Handle("PUT /api/v1/groups/{id}/members/{userId}/timeout", groupLimit.Wrap(http.HandlerFunc(h.TimeoutGroupMember)))
	mux.HandleFunc("GET /api/v1/groups/{id}/bans", h.ListGroupBans)
	mux.Handle("POST /api/v1/groups/{id}/bans", groupLimit.Wrap(http.HandlerFunc(h.BanGroupMembers)))
	mux.Handle("DELETE /api/v1/groups/{id}/bans/{userId}", groupLimit.Wrap(http.HandlerFunc(h.UnbanGroupMember)))
}
