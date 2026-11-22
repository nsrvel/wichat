package http

import (
	"net/http"
	"time"

	"github.com/wichat/wichat/backend/api-gateway/internal/middleware"
)

// Register mounts chat message routes on mux.
func Register(mux *http.ServeMux, h Handler) {
	messageLimit := middleware.NewLimiter(120, time.Hour)
	messageReadLimit := middleware.NewLimiter(600, time.Hour)
	attachmentLimit := middleware.NewLimiter(60, time.Hour)

	mux.Handle("GET /api/v1/groups/{groupId}/channels/{channelId}/messages", messageReadLimit.Wrap(http.HandlerFunc(h.ListChannelMessages)))
	mux.Handle("POST /api/v1/groups/{groupId}/channels/{channelId}/messages", messageLimit.Wrap(http.HandlerFunc(h.CreateChannelMessage)))
	mux.Handle("POST /api/v1/groups/{groupId}/channels/{channelId}/attachments", attachmentLimit.Wrap(http.HandlerFunc(h.UploadChannelAttachment)))
	mux.HandleFunc("PATCH /api/v1/groups/{groupId}/channels/{channelId}/messages/{messageId}", h.UpdateChannelMessage)
	mux.HandleFunc("DELETE /api/v1/groups/{groupId}/channels/{channelId}/messages/{messageId}", h.DeleteChannelMessage)
	mux.HandleFunc("POST /api/v1/groups/{groupId}/channels/{channelId}/messages/{messageId}/reactions", h.ToggleMessageReaction)
	mux.Handle("GET /api/v1/groups/{groupId}/channels/{channelId}/pins", messageReadLimit.Wrap(http.HandlerFunc(h.ListPinnedChannelMessages)))
	mux.HandleFunc("PUT /api/v1/groups/{groupId}/channels/{channelId}/pins/{messageId}", h.PinChannelMessage)
	mux.HandleFunc("DELETE /api/v1/groups/{groupId}/channels/{channelId}/pins/{messageId}", h.UnpinChannelMessage)
}
