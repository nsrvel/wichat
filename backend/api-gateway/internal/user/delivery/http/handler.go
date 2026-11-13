package http

import (
	"net/http"
	"strconv"
	"strings"

	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"github.com/wichat/wichat/backend/wi-shared/response"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

// Handler is the public user & instance-setup HTTP delivery.
type Handler interface {
	SetupStatus(w http.ResponseWriter, r *http.Request)
	InstanceSetup(w http.ResponseWriter, r *http.Request)
	CheckUsername(w http.ResponseWriter, r *http.Request)
	SearchUsers(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	ms  *microservices.Clients
	log logger.Logger
}

// NewHandler wires user HTTP to the ms-user client.
func NewHandler(ms *microservices.Clients, log logger.Logger) Handler {
	return handler{ms: ms, log: log}
}

func (h handler) SetupStatus(w http.ResponseWriter, r *http.Request) {
	// Service
	out, err := h.ms.User.GRPC.User.SetupStatus(r.Context(), &userv1.SetupStatusRequest{})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.SetupStatusBody{
		RegistrationMode: out.GetRegistrationMode(),
	}))
}

func (h handler) InstanceSetup(w http.ResponseWriter, r *http.Request) {
	// Bind
	token := authHttp.Bearer(r)
	if token == "" {
		responsehttp.WriteError(w, r, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	var body authHttp.InstanceSetupBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Service
	_, err := h.ms.User.GRPC.User.CompleteInstanceSetup(r.Context(), &userv1.CompleteInstanceSetupRequest{
		AccessToken:      token,
		Name:             body.Name,
		RegistrationMode: body.RegistrationMode,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.OKBody{OK: true}))
}

func (h handler) CheckUsername(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimSpace(r.URL.Query().Get("username"))
	out, err := h.ms.User.GRPC.User.CheckUsername(r.Context(), &userv1.CheckUsernameRequest{
		Username: raw,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.CheckUsernameBody{
		Available:  out.GetAvailable(),
		Normalized: out.GetNormalized(),
		Reason:     out.GetReason(),
	}))
}

func (h handler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	// Bind
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	groupID := strings.TrimSpace(r.URL.Query().Get("exclude_group_id"))
	limit := 20
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}

	// Service
	out, err := h.ms.User.GRPC.User.SearchUsers(r.Context(), &userv1.SearchUsersRequest{
		AccessToken: token,
		Query:       query,
		Limit:       int32(limit), //nolint:gosec // G115 — query limit clamped above
		GroupId:     groupID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.UserSearchBody{
		Users: authHttp.UserSearchFromPB(out.GetUsers()),
	}))
}
