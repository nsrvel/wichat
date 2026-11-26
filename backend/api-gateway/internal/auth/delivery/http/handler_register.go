package http

import (
	"net/http"

	"github.com/wichat/wichat/backend/api-gateway/internal/middleware"
	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"github.com/wichat/wichat/backend/wi-shared/response"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

func (h handler) Setup(w http.ResponseWriter, r *http.Request) {
	h.Register(w, r)
}

func (h handler) Register(w http.ResponseWriter, r *http.Request) {
	// Bind
	var body authHttp.RegisterBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	if err := h.turnstile.Verify(r.Context(), body.TurnstileToken, middleware.ClientIP(r)); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Service
	out, err := h.ms.User.GRPC.Auth.Register(r.Context(), &userv1.RegisterRequest{
		Email:       body.Email,
		Username:    body.Username,
		Password:    body.Password,
		DisplayName: body.DisplayName,
		BirthDate:   body.BirthDate,
		Locale:      body.Locale,
		InviteToken: body.InviteToken,
		UserAgent:   middleware.UserAgent(r),
		ClientIp:    middleware.ClientIP(r),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	authHttp.WritePasswordAuth(w, out.GetSession(), out.GetChallenge(), h.secureCookie)
}

func (h handler) VerifyRegister(w http.ResponseWriter, r *http.Request) {
	var body authHttp.VerifyEmailCodeBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}
	if !h.verifyTurnstile(w, r, body.TurnstileToken) {
		return
	}

	out, err := h.ms.User.GRPC.Auth.VerifyRegister(r.Context(), &userv1.VerifyRegisterRequest{
		ChallengeId: body.ChallengeID,
		Code:        body.Code,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	authHttp.WriteSession(w, out.GetSession(), h.secureCookie)
}

func (h handler) ResendRegisterCode(w http.ResponseWriter, r *http.Request) {
	var body authHttp.ResendEmailCodeBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}
	if !h.verifyTurnstile(w, r, body.TurnstileToken) {
		return
	}

	out, err := h.ms.User.GRPC.Auth.ResendRegisterCode(r.Context(), &userv1.ResendRegisterCodeRequest{
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
