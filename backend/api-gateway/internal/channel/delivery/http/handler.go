package http

import (
	"net/http"
	"strings"

	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
)

// Handler is the public channel HTTP delivery.
type Handler interface {
	ListGroupChannels(w http.ResponseWriter, r *http.Request)
	CreateGroupChannel(w http.ResponseWriter, r *http.Request)
	UpdateGroupChannel(w http.ResponseWriter, r *http.Request)
	DeleteGroupChannel(w http.ResponseWriter, r *http.Request)
	ListChannelPermissions(w http.ResponseWriter, r *http.Request)
	UpsertChannelPermission(w http.ResponseWriter, r *http.Request)
	DeleteChannelPermission(w http.ResponseWriter, r *http.Request)
	SyncChannelPermissions(w http.ResponseWriter, r *http.Request)
	ReorderGroupChannels(w http.ResponseWriter, r *http.Request)
	EnsureGroupDmChannel(w http.ResponseWriter, r *http.Request)
	WithGroupBroadcaster(b GroupRealtimeBroadcaster) Handler
}

// GroupRealtimeBroadcaster pushes group-scoped events to subscribed WebSocket clients.
type GroupRealtimeBroadcaster interface {
	DeliverGroup(groupID string, payload []byte) error
}

type handler struct {
	ms          *microservices.Clients
	log         logger.Logger
	broadcaster GroupRealtimeBroadcaster
}

// NewHandler wires channel HTTP to the ms-user client.
func NewHandler(ms *microservices.Clients, log logger.Logger) Handler {
	return handler{ms: ms, log: log}
}

func (h handler) WithGroupBroadcaster(b GroupRealtimeBroadcaster) Handler {
	h.broadcaster = b
	return h
}

func (h handler) requireBearerAndGroupID(w http.ResponseWriter, r *http.Request) (string, string, bool) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return "", "", false
	}

	groupID := strings.TrimSpace(r.PathValue("id"))
	if groupID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "id"}))
		return "", "", false
	}

	return token, groupID, true
}
