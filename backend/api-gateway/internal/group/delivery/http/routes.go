package http

import (
	"net/http"
	"time"

	"github.com/wichat/wichat/backend/api-gateway/internal/middleware"
)

// Register mounts group routes on mux.
func Register(mux *http.ServeMux, h Handler) {
	groupLimit := middleware.NewLimiter(30, time.Hour)

	mux.HandleFunc("GET /api/v1/groups/mine", h.ListMyGroups)
	mux.Handle("PUT /api/v1/groups/mine/order", groupLimit.Wrap(http.HandlerFunc(h.ReorderMyGroups)))
	mux.HandleFunc("GET /api/v1/groups/joinable", h.ListJoinableGroups)
	mux.Handle("POST /api/v1/groups", groupLimit.Wrap(http.HandlerFunc(h.CreateGroup)))
	mux.Handle("POST /api/v1/groups/{id}/delete", groupLimit.Wrap(http.HandlerFunc(h.DeleteGroup)))
	mux.Handle("DELETE /api/v1/groups/{id}", groupLimit.Wrap(http.HandlerFunc(h.DeleteGroup)))
	mux.Handle("POST /api/v1/groups/{id}/icon", groupLimit.Wrap(http.HandlerFunc(h.UploadGroupIcon)))
	mux.Handle("PATCH /api/v1/groups/{id}/profile", groupLimit.Wrap(http.HandlerFunc(h.UpdateGroupProfile)))
	mux.Handle("POST /api/v1/groups/{id}/banner", groupLimit.Wrap(http.HandlerFunc(h.UploadGroupBanner)))
}
