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
}
