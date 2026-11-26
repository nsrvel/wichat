package http

import "net/http"

// Register mounts channel routes on mux.
func Register(mux *http.ServeMux, h Handler) {
	mux.HandleFunc("GET /api/v1/groups/{id}/channels", h.ListGroupChannels)
	mux.HandleFunc("POST /api/v1/groups/{id}/channels", h.CreateGroupChannel)
	mux.HandleFunc("PATCH /api/v1/groups/{id}/channels/{channelId}", h.UpdateGroupChannel)
	mux.HandleFunc("DELETE /api/v1/groups/{id}/channels/{channelId}", h.DeleteGroupChannel)
	mux.HandleFunc("GET /api/v1/groups/{id}/channels/{channelId}/permissions", h.ListChannelPermissions)
	mux.HandleFunc("PUT /api/v1/groups/{id}/channels/{channelId}/permissions", h.UpsertChannelPermission)
	mux.HandleFunc("DELETE /api/v1/groups/{id}/channels/{channelId}/permissions/{targetType}/{targetId}", h.DeleteChannelPermission)
	mux.HandleFunc("POST /api/v1/groups/{id}/channels/{channelId}/permissions/sync", h.SyncChannelPermissions)
	mux.HandleFunc("PUT /api/v1/groups/{id}/channels/order", h.ReorderGroupChannels)
	mux.HandleFunc("POST /api/v1/groups/{id}/dm", h.EnsureGroupDmChannel)
}
