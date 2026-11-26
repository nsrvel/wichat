package http

import (
	"net/http"
	"time"

	"github.com/wichat/wichat/backend/api-gateway/internal/middleware"
)

// Register mounts chat message routes on mux.
func Register(mux *http.ServeMux, h Handler) {
	messageLimit := middleware.NewLimiter(120, time.Hour)
	messageReadLimit := middleware.NewLimiter(2400, time.Hour)
	attachmentLimit := middleware.NewLimiter(60, time.Hour)

	mux.Handle("GET /api/v1/groups/{groupId}/search/messages", messageReadLimit.Wrap(http.HandlerFunc(h.SearchGroupMessages)))
	mux.Handle("GET /api/v1/groups/{groupId}/search/files", messageReadLimit.Wrap(http.HandlerFunc(h.SearchGroupFiles)))
	mux.Handle("GET /api/v1/groups/{groupId}/mentions", messageReadLimit.Wrap(http.HandlerFunc(h.ListGroupMentions)))
	mux.Handle("GET /api/v1/groups/{groupId}/mentions/unread-count", messageReadLimit.Wrap(http.HandlerFunc(h.CountGroupMentionsUnread)))
	mux.Handle("POST /api/v1/groups/{groupId}/mentions/read", messageLimit.Wrap(http.HandlerFunc(h.MarkGroupMentionsRead)))
	mux.Handle("GET /api/v1/groups/{groupId}/channels/{channelId}/messages", messageReadLimit.Wrap(http.HandlerFunc(h.ListChannelMessages)))
	mux.Handle("GET /api/v1/groups/{groupId}/channels/{channelId}/messages/{messageId}/seq", messageReadLimit.Wrap(http.HandlerFunc(h.ResolveMessageChannelSeq)))
	mux.Handle("GET /api/v1/messages/{messageId}/seq", messageReadLimit.Wrap(http.HandlerFunc(h.ResolveMessageChannelSeqGlobal)))
	mux.Handle("POST /api/v1/groups/{groupId}/channels/{channelId}/messages", messageLimit.Wrap(http.HandlerFunc(h.CreateChannelMessage)))
	mux.Handle("POST /api/v1/groups/{groupId}/channels/{channelId}/attachments", attachmentLimit.Wrap(http.HandlerFunc(h.UploadChannelAttachment)))
	mux.HandleFunc("PATCH /api/v1/groups/{groupId}/channels/{channelId}/messages/{messageId}", h.UpdateChannelMessage)
	mux.HandleFunc("DELETE /api/v1/groups/{groupId}/channels/{channelId}/messages/{messageId}", h.DeleteChannelMessage)
	mux.HandleFunc("POST /api/v1/groups/{groupId}/channels/{channelId}/messages/{messageId}/reactions", h.ToggleMessageReaction)
	mux.Handle("GET /api/v1/groups/{groupId}/channels/{channelId}/pins", messageReadLimit.Wrap(http.HandlerFunc(h.ListPinnedChannelMessages)))
	mux.HandleFunc("PUT /api/v1/groups/{groupId}/channels/{channelId}/pins/{messageId}", h.PinChannelMessage)
	mux.HandleFunc("DELETE /api/v1/groups/{groupId}/channels/{channelId}/pins/{messageId}", h.UnpinChannelMessage)
}
