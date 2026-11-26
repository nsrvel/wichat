package http

import (
	"net/http"

	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

func (h handler) ForgotPasswordRequest(w http.ResponseWriter, r *http.Request) {
	// Bind
	var body authHttp.ForgotPasswordRequestBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}
	if !h.verifyTurnstile(w, r, body.TurnstileToken) {
		return
	}

	// Service
	out, err := h.ms.User.GRPC.Auth.RequestPasswordReset(r.Context(), &userv1.RequestPasswordResetRequest{
		Email:  body.Email,
		Locale: body.Locale,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, map[string]any{
		"data": authHttp.ForgotPasswordRequestData{
			OK:                    true,
			ResendCooldownSeconds: int(out.GetResendCooldownSeconds()),
		},
	})
}

func (h handler) ForgotPasswordVerify(w http.ResponseWriter, r *http.Request) {
	// Bind
	var body authHttp.ForgotPasswordVerifyBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}
	if !h.verifyTurnstile(w, r, body.TurnstileToken) {
		return
	}

	// Service
	out, err := h.ms.User.GRPC.Auth.VerifyPasswordReset(r.Context(), &userv1.VerifyPasswordResetRequest{
		Email: body.Email,
		Code:  body.Code,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, map[string]any{
		"data": authHttp.ForgotPasswordVerifyData{ResetToken: out.GetResetToken()},
	})
}

func (h handler) ForgotPasswordReset(w http.ResponseWriter, r *http.Request) {
	// Bind
	var body authHttp.ForgotPasswordResetBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Service
	_, err := h.ms.User.GRPC.Auth.CompletePasswordReset(r.Context(), &userv1.CompletePasswordResetRequest{
		ResetToken: body.ResetToken,
		Password:   body.Password,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, map[string]any{"data": authHttp.OKBody{OK: true}})
}
