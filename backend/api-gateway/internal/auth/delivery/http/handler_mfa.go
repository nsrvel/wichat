package http

import (
	"net/http"

	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"github.com/wichat/wichat/backend/wi-shared/response"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

func (h handler) BeginLoginWebAuthn(w http.ResponseWriter, r *http.Request) {
	var body authHttp.BeginLoginWebAuthnBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	out, err := h.ms.User.GRPC.Auth.BeginLoginWebAuthn(r.Context(), &userv1.BeginLoginWebAuthnRequest{
		ChallengeId: body.ChallengeID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.WebAuthnOptionsData{
		WebAuthnOptions: []byte(out.GetWebauthnOptionsJson()),
	}))
}

func (h handler) VerifyLoginMFA(w http.ResponseWriter, r *http.Request) {
	var body authHttp.VerifyLoginMFABody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}
	if !h.verifyTurnstile(w, r, body.TurnstileToken) {
		return
	}

	out, err := h.ms.User.GRPC.Auth.VerifyLoginMFA(r.Context(), &userv1.VerifyLoginMFARequest{
		ChallengeId:   body.ChallengeID,
		Method:        body.Method,
		Code:          body.Code,
		AssertionJson: string(body.Assertion),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	authHttp.WriteSession(w, out.GetSession(), h.secureCookie)
}
