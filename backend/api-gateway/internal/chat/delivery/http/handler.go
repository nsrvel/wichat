package http

import (
	"net/http"
	"strings"

	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
)

const (
	maxMessageBodyBytes      = 1 << 14
	maxAttachmentUploadBytes = 25 << 20
	maxMultipartOverhead     = 1 << 20
)

// Handler is the public chat message HTTP delivery.
type Handler interface {
	ListChannelMessages(w http.ResponseWriter, r *http.Request)
	CreateChannelMessage(w http.ResponseWriter, r *http.Request)
	UploadChannelAttachment(w http.ResponseWriter, r *http.Request)
	UpdateChannelMessage(w http.ResponseWriter, r *http.Request)
	DeleteChannelMessage(w http.ResponseWriter, r *http.Request)
	ToggleMessageReaction(w http.ResponseWriter, r *http.Request)
	ListPinnedChannelMessages(w http.ResponseWriter, r *http.Request)
	PinChannelMessage(w http.ResponseWriter, r *http.Request)
	UnpinChannelMessage(w http.ResponseWriter, r *http.Request)
	ResolveMessageChannelSeq(w http.ResponseWriter, r *http.Request)
	ResolveMessageChannelSeqGlobal(w http.ResponseWriter, r *http.Request)
	SearchGroupMessages(w http.ResponseWriter, r *http.Request)
	SearchGroupFiles(w http.ResponseWriter, r *http.Request)
	ListGroupMentions(w http.ResponseWriter, r *http.Request)
	MarkGroupMentionsRead(w http.ResponseWriter, r *http.Request)
	CountGroupMentionsUnread(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	ms  *microservices.Clients
	log logger.Logger
}

// NewHandler constructs chat HTTP handlers.
func NewHandler(ms *microservices.Clients, log logger.Logger) Handler {
	return &handler{ms: ms, log: log}
}

func (h handler) requireChannelContext(w http.ResponseWriter, r *http.Request) (token, groupID, channelID string, ok bool) {
	token = authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return "", "", "", false
	}

	groupID = strings.TrimSpace(r.PathValue("groupId"))
	channelID = strings.TrimSpace(r.PathValue("channelId"))
	if groupID == "" || channelID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_id"}))
		return "", "", "", false
	}

	return token, groupID, channelID, true
}

func (h handler) requireMessageContext(w http.ResponseWriter, r *http.Request) (token, groupID, channelID, messageID string, ok bool) {
	token, groupID, channelID, ok = h.requireChannelContext(w, r)
	if !ok {
		return "", "", "", "", false
	}
	messageID = strings.TrimSpace(r.PathValue("messageId"))
	if messageID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "message_id"}))
		return "", "", "", "", false
	}

	return token, groupID, channelID, messageID, true
}
