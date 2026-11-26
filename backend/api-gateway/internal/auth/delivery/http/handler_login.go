package http

import (
	"net/http"
	"strings"

	"github.com/wichat/wichat/backend/api-gateway/internal/middleware"
	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"github.com/wichat/wichat/backend/wi-shared/response"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

func (h handler) Login(w http.ResponseWriter, r *http.Request) {
	// Bind
	var body authHttp.LoginBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	if err := h.turnstile.Verify(r.Context(), body.TurnstileToken, middleware.ClientIP(r)); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	login := strings.TrimSpace(body.Login)
	if login == "" {
		login = strings.TrimSpace(body.Email)
	}

	// Service
	out, err := h.ms.User.GRPC.Auth.Login(r.Context(), &userv1.LoginRequest{
		Login:     login,
		Password:  body.Password,
		Locale:    body.Locale,
		UserAgent: middleware.UserAgent(r),
		ClientIp:  middleware.ClientIP(r),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	authHttp.WritePasswordAuth(w, out.GetSession(), out.GetChallenge(), h.secureCookie)
}

func (h handler) VerifyLogin(w http.ResponseWriter, r *http.Request) {
	var body authHttp.VerifyEmailCodeBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}
	if !h.verifyTurnstile(w, r, body.TurnstileToken) {
		return
	}

	out, err := h.ms.User.GRPC.Auth.VerifyLogin(r.Context(), &userv1.VerifyLoginRequest{
		ChallengeId: body.ChallengeID,
		Code:        body.Code,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	authHttp.WriteSession(w, out.GetSession(), h.secureCookie)
}

func (h handler) ResendLoginCode(w http.ResponseWriter, r *http.Request) {
	var body authHttp.ResendEmailCodeBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}
	if !h.verifyTurnstile(w, r, body.TurnstileToken) {
		return
	}

	out, err := h.ms.User.GRPC.Auth.ResendLoginCode(r.Context(), &userv1.ResendLoginCodeRequest{
		ChallengeId: body.ChallengeID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.ResendEmailCodeData{
		ResendAfterSeconds: out.GetResendCooldownSeconds(),
	}))
}
