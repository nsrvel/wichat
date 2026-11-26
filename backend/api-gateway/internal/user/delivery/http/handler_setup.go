package http

import (
	"net/http"
	"strings"

	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"github.com/wichat/wichat/backend/wi-shared/response"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

func (h handler) SetupStatus(w http.ResponseWriter, r *http.Request) {
	// Service
	out, err := h.ms.User.GRPC.User.SetupStatus(r.Context(), &userv1.SetupStatusRequest{})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.SetupStatusBody{
		RegistrationMode:  out.GetRegistrationMode(),
		InviteOnly:        h.inviteOnly,
		TurnstileSiteKey:  h.turnstileSiteKey,
		TurnstileRequired: h.turnstileRequired,
		MfaEnabled:        out.GetMfaEnabled(),
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
