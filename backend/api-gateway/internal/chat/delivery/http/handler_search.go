package http

import (
	"net/http"
	"strconv"
	"strings"

	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	chatHttp "github.com/wichat/wichat/backend/api-gateway/pkg/chat-http"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	chatv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/chat/v1"
	"github.com/wichat/wichat/backend/wi-shared/response"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

func (h handler) SearchGroupMessages(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	groupID := strings.TrimSpace(r.PathValue("groupId"))
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if groupID == "" || query == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, nil))
		return
	}

	limit := int32(8)
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		n, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || n <= 0 {
			authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "limit"}))
			return
		}
		limit = int32(n)
	}

	out, err := h.ms.Chat.GRPC.Search.SearchMessages(r.Context(), &chatv1.SearchMessagesRequest{
		AccessToken: token,
		GroupId:     groupID,
		Query:       query,
		Limit:       limit,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(chatHttp.MessageSearchHitsBody{
		Hits: chatHttp.MessageSearchHitsFromPB(out.GetHits()),
	}))
}

func (h handler) SearchGroupFiles(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	groupID := strings.TrimSpace(r.PathValue("groupId"))
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if groupID == "" || query == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, nil))
		return
	}

	limit := int32(8)
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		n, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || n <= 0 {
			authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "limit"}))
			return
		}
		limit = int32(n)
	}

	imagesOnly := strings.EqualFold(r.URL.Query().Get("images"), "1") ||
		strings.EqualFold(r.URL.Query().Get("images"), "true")

	out, err := h.ms.Chat.GRPC.Search.SearchFiles(r.Context(), &chatv1.SearchFilesRequest{
		AccessToken: token,
		GroupId:     groupID,
		Query:       query,
		Limit:       limit,
		ImagesOnly:  imagesOnly,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	type fileHitBody struct {
		AttachmentID string `json:"attachment_id"`
		MessageID    string `json:"message_id"`
		GroupID      string `json:"group_id"`
		ChannelID    string `json:"channel_id"`
		Filename     string `json:"filename"`
		ContentType  string `json:"content_type"`
		CreatedAt    string `json:"created_at"`
	}
	hits := make([]fileHitBody, 0, len(out.GetHits()))
	for _, hit := range out.GetHits() {
		if hit == nil {
			continue
		}
		hits = append(hits, fileHitBody{
			AttachmentID: hit.GetAttachmentId(),
			MessageID:    hit.GetMessageId(),
			GroupID:      hit.GetGroupId(),
			ChannelID:    hit.GetChannelId(),
			Filename:     hit.GetFilename(),
			ContentType:  hit.GetContentType(),
			CreatedAt:    hit.GetCreatedAt(),
		})
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(map[string]any{"hits": hits}))
}
