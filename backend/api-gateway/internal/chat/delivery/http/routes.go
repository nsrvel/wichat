package http

import (
	"net/http"
	"time"

	"github.com/wichat/wichat/backend/api-gateway/internal/middleware"
)

// Register mounts chat message routes on mux.
func Register(mux *http.ServeMux, h Handler) {
	messageLimit := middleware.NewLimiter(120, time.Hour)

	mux.HandleFunc("GET /api/v1/groups/{groupId}/channels/{channelId}/messages", h.ListChannelMessages)
	mux.Handle("POST /api/v1/groups/{groupId}/channels/{channelId}/messages", messageLimit.Wrap(http.HandlerFunc(h.CreateChannelMessage)))
}
