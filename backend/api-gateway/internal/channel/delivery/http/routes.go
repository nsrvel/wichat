package http

import "net/http"

// Register mounts channel routes on mux.
func Register(mux *http.ServeMux, h Handler) {
	mux.HandleFunc("GET /api/v1/groups/{id}/channels", h.ListGroupChannels)
}
