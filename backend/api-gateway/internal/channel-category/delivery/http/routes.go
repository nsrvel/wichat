package http

import "net/http"

// Register mounts channel category routes on mux.
func Register(mux *http.ServeMux, h Handler) {
	mux.HandleFunc("POST /api/v1/groups/{id}/categories", h.CreateGroupCategory)
	mux.HandleFunc("PATCH /api/v1/groups/{id}/categories/{categoryId}", h.UpdateGroupCategory)
	mux.HandleFunc("DELETE /api/v1/groups/{id}/categories/{categoryId}", h.DeleteGroupCategory)
	mux.HandleFunc("GET /api/v1/groups/{id}/categories/{categoryId}/permissions", h.ListCategoryPermissions)
	mux.HandleFunc("PUT /api/v1/groups/{id}/categories/{categoryId}/permissions", h.UpsertCategoryPermission)
	mux.HandleFunc("DELETE /api/v1/groups/{id}/categories/{categoryId}/permissions/{targetType}/{targetId}", h.DeleteCategoryPermission)
	mux.HandleFunc("PUT /api/v1/groups/{id}/categories/order", h.ReorderChannelCategories)
}
