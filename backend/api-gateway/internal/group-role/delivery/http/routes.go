package http

import (
	"net/http"
	"time"

	"github.com/wichat/wichat/backend/api-gateway/internal/middleware"
)

// Register mounts group-role routes on mux.
func Register(mux *http.ServeMux, h Handler) {
	roleLimit := middleware.NewLimiter(60, time.Minute)

	mux.HandleFunc("GET /api/v1/groups/{id}/roles", h.ListGroupRoles)
	mux.Handle("POST /api/v1/groups/{id}/roles", roleLimit.Wrap(http.HandlerFunc(h.CreateGroupRole)))
	mux.Handle("PATCH /api/v1/groups/{id}/roles/settings", roleLimit.Wrap(http.HandlerFunc(h.UpdateGroupHideRoles)))
	mux.Handle("PATCH /api/v1/groups/{id}/roles/{roleId}", roleLimit.Wrap(http.HandlerFunc(h.UpdateGroupRole)))
	mux.Handle("POST /api/v1/groups/{id}/roles/{roleId}/duplicate", roleLimit.Wrap(http.HandlerFunc(h.DuplicateGroupRole)))
	mux.Handle("DELETE /api/v1/groups/{id}/roles/{roleId}", roleLimit.Wrap(http.HandlerFunc(h.DeleteGroupRole)))
	mux.HandleFunc("GET /api/v1/groups/{id}/roles/{roleId}/members", h.ListGroupRoleMembers)
	mux.Handle("POST /api/v1/groups/{id}/roles/{roleId}/members", roleLimit.Wrap(http.HandlerFunc(h.AddGroupRoleMembers)))
	mux.Handle("DELETE /api/v1/groups/{id}/roles/{roleId}/members/{userId}", roleLimit.Wrap(http.HandlerFunc(h.RemoveGroupRoleMember)))
}
