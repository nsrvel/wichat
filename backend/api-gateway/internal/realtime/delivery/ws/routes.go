package ws

import "net/http"

// Register mounts realtime WebSocket routes on mux.
func Register(mux *http.ServeMux, h Handler) {
	mux.HandleFunc("GET /api/v1/realtime", h.Realtime)
}
