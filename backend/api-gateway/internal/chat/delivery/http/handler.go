package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	chatHttp "github.com/wichat/wichat/backend/api-gateway/pkg/chat-http"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
	chatv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/chat/v1"
	"github.com/wichat/wichat/backend/wi-shared/response"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

const maxMessageBodyBytes = 1 << 12

// Handler is the public chat message HTTP delivery.
type Handler interface {
	ListChannelMessages(w http.ResponseWriter, r *http.Request)
	CreateChannelMessage(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	ms  *microservices.Clients
	log logger.Logger
}

// NewHandler constructs chat HTTP handlers.
func NewHandler(ms *microservices.Clients, log logger.Logger) Handler {
	return &handler{ms: ms, log: log}
}

func (h handler) ListChannelMessages(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	groupID := strings.TrimSpace(r.PathValue("groupId"))
	channelID := strings.TrimSpace(r.PathValue("channelId"))
	if groupID == "" || channelID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_id"}))
		return
	}

	limit := int32(0)
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || parsed < 0 {
			authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "limit"}))
			return
		}
		limit = int32(parsed)
	}

	out, err := h.ms.Chat.GRPC.Message.ListChannelMessages(r.Context(), &chatv1.ListChannelMessagesRequest{
		AccessToken:     token,
		GroupId:         groupID,
		ChannelId:       channelID,
		Limit:           limit,
		BeforeCreatedAt: strings.TrimSpace(r.URL.Query().Get("before_created_at")),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(chatHttp.MessagesListBody{
		Messages: chatHttp.MessagesFromPB(out.GetMessages()),
	}))
}

func (h handler) CreateChannelMessage(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	groupID := strings.TrimSpace(r.PathValue("groupId"))
	channelID := strings.TrimSpace(r.PathValue("channelId"))
	if groupID == "" || channelID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_id"}))
		return
	}

	// Bind
	r.Body = http.MaxBytesReader(w, r.Body, maxMessageBodyBytes)
	var body chatHttp.CreateMessageBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, nil))
		return
	}

	out, err := h.ms.Chat.GRPC.Message.CreateChannelMessage(r.Context(), &chatv1.CreateChannelMessageRequest{
		AccessToken: token,
		GroupId:     groupID,
		ChannelId:   channelID,
		Body:        body.Body,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusCreated, response.NewDataEnvelope(chatHttp.CreateMessageResponseBody{
		Message: chatHttp.MessageFromPB(out.GetMessage()),
	}))
}
