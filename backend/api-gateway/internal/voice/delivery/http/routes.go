package http

import "net/http"

// Register mounts voice/video and LiveKit webhook routes on mux.
func Register(mux *http.ServeMux, h Handler) {
	mux.HandleFunc("POST /api/v1/channels/{channelId}/voice/token", h.IssueVoiceToken)
	mux.HandleFunc("POST /api/v1/groups/{groupId}/channels/{channelId}/voice/token", h.IssueVoiceToken)
	mux.HandleFunc("GET /api/v1/groups/{groupId}/voice/occupancy", h.ListGroupOccupancy)
	mux.HandleFunc("POST /api/v1/channels/{channelId}/voice/moderate", h.ModerateParticipant)
	mux.HandleFunc("POST /api/v1/channels/{channelId}/voice/keepalive", h.AcknowledgeIdleWarning)
	mux.HandleFunc("POST /api/v1/channels/{channelId}/voice/recording/start", h.StartRecording)
	mux.HandleFunc("POST /api/v1/channels/{channelId}/voice/recording/stop", h.StopRecording)
	mux.HandleFunc("GET /api/v1/channels/{channelId}/voice/recordings", h.ListChannelRecordings)
	mux.HandleFunc("GET /api/v1/voice/quota", h.GetQuotaUsage)
	mux.HandleFunc("POST /api/v1/webhooks/livekit", h.HandleWebhook)
}
