package http

import (
	"net/http"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
)

// Handler is the public group lifecycle & profile HTTP delivery.
type Handler interface {
	ListMyGroups(w http.ResponseWriter, r *http.Request)
	ReorderMyGroups(w http.ResponseWriter, r *http.Request)
	CreateGroup(w http.ResponseWriter, r *http.Request)
	DeleteGroup(w http.ResponseWriter, r *http.Request)
	UploadGroupIcon(w http.ResponseWriter, r *http.Request)
	UpdateGroupProfile(w http.ResponseWriter, r *http.Request)
	UploadGroupBanner(w http.ResponseWriter, r *http.Request)
	ListGroupAuditLogs(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	ms  *microservices.Clients
	log logger.Logger
}

// NewHandler wires group HTTP to the ms-user client.
func NewHandler(ms *microservices.Clients, log logger.Logger) Handler {
	return handler{ms: ms, log: log}
}
