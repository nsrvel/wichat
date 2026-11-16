package http

import "net/http"

// Register mounts channel routes on mux.
func Register(mux *http.ServeMux, h Handler) {
	mux.HandleFunc("GET /api/v1/groups/{id}/channels", h.ListGroupChannels)
	mux.HandleFunc("POST /api/v1/groups/{id}/channels", h.CreateGroupChannel)
	mux.HandleFunc("POST /api/v1/groups/{id}/categories", h.CreateGroupCategory)
	mux.HandleFunc("PUT /api/v1/groups/{id}/categories/order", h.ReorderChannelCategories)
	mux.HandleFunc("PUT /api/v1/groups/{id}/channels/order", h.ReorderGroupChannels)
}
