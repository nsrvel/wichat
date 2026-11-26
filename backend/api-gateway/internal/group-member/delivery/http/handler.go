package http

import (
	"net/http"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
)

// Handler is the public group-member HTTP delivery.
type Handler interface {
	JoinGroup(w http.ResponseWriter, r *http.Request)
	LeaveGroup(w http.ResponseWriter, r *http.Request)
	ListGroupMembers(w http.ResponseWriter, r *http.Request)
	KickGroupMember(w http.ResponseWriter, r *http.Request)
	BanGroupMembers(w http.ResponseWriter, r *http.Request)
	ListGroupBans(w http.ResponseWriter, r *http.Request)
	UnbanGroupMember(w http.ResponseWriter, r *http.Request)
	TimeoutGroupMember(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	ms  *microservices.Clients
	log logger.Logger
}

// NewHandler wires group-member HTTP to the ms-user client.
func NewHandler(ms *microservices.Clients, log logger.Logger) Handler {
	return handler{ms: ms, log: log}
}
