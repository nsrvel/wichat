package http

import (
	"encoding/json"
	"net/http"
	"strings"

	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"github.com/wichat/wichat/backend/wi-shared/response"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

func (h handler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	// Auth
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	// Input
	var body authHttp.UpdatePasswordBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Service
	_, err := h.ms.User.GRPC.User.UpdatePassword(r.Context(), &userv1.UpdatePasswordRequest{
		AccessToken:     token,
		CurrentPassword: body.CurrentPassword,
		NewPassword:     body.NewPassword,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.OKBody{OK: true}))
}

func (h handler) GetMFAStatus(w http.ResponseWriter, r *http.Request) {
	// Auth
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	// Service
	out, err := h.ms.User.GRPC.User.GetMFAStatus(r.Context(), &userv1.GetMFAStatusRequest{
		AccessToken: token,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	keys := make([]authHttp.SecurityKeyBody, len(out.GetSecurityKeys()))
	for i, k := range out.GetSecurityKeys() {
		keys[i] = authHttp.SecurityKeyFromPB(k)
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.MFAStatusBody{
		TOTPEnabled:            out.GetTotpEnabled(),
		SecurityKeys:           keys,
		RecoveryCodesRemaining: out.GetRecoveryCodesRemaining(),
	}))
}

func (h handler) BeginTOTPSetup(w http.ResponseWriter, r *http.Request) {
	// Auth
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	// Input
	var body authHttp.CurrentPasswordBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Service
	out, err := h.ms.User.GRPC.User.BeginTOTPSetup(r.Context(), &userv1.BeginTOTPSetupRequest{
		AccessToken:     token,
		CurrentPassword: body.CurrentPassword,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.TOTPSetupBeginData{
		SetupToken: out.GetSetupToken(),
		Secret:     out.GetSecret(),
		OtpauthURI: out.GetOtpauthUri(),
	}))
}

func (h handler) ConfirmTOTPSetup(w http.ResponseWriter, r *http.Request) {
	// Auth
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	// Input
	var body authHttp.TOTPSetupConfirmBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Service
	out, err := h.ms.User.GRPC.User.ConfirmTOTPSetup(r.Context(), &userv1.ConfirmTOTPSetupRequest{
		AccessToken: token,
		SetupToken:  body.SetupToken,
		Code:        body.Code,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	codes := out.GetRecoveryCodes()
	if codes == nil {
		codes = []string{}
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.RecoveryCodesData{
		RecoveryCodes: codes,
	}))
}

func (h handler) DisableTOTP(w http.ResponseWriter, r *http.Request) {
	// Auth
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	// Input
	var body authHttp.CurrentPasswordBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Service
	_, err := h.ms.User.GRPC.User.DisableTOTP(r.Context(), &userv1.DisableTOTPRequest{
		AccessToken:     token,
		CurrentPassword: body.CurrentPassword,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.OKBody{OK: true}))
}

func (h handler) BeginWebAuthnRegister(w http.ResponseWriter, r *http.Request) {
	// Auth
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	// Input
	var body authHttp.BeginWebAuthnRegisterBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Service
	out, err := h.ms.User.GRPC.User.BeginWebAuthnRegister(r.Context(), &userv1.BeginWebAuthnRegisterRequest{
		AccessToken:     token,
		CurrentPassword: body.CurrentPassword,
		Name:            body.Name,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.BeginWebAuthnRegisterData{
		ChallengeID: out.GetChallengeId(),
		Options:     json.RawMessage(out.GetOptionsJson()),
	}))
}

func (h handler) FinishWebAuthnRegister(w http.ResponseWriter, r *http.Request) {
	// Auth
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	// Input
	var body authHttp.FinishWebAuthnRegisterBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Service
	out, err := h.ms.User.GRPC.User.FinishWebAuthnRegister(r.Context(), &userv1.FinishWebAuthnRegisterRequest{
		AccessToken:    token,
		ChallengeId:    body.ChallengeID,
		CredentialJson: string(body.Credential),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.FinishWebAuthnRegisterData{
		SecurityKey:   authHttp.SecurityKeyFromPB(out.GetSecurityKey()),
		RecoveryCodes: out.GetRecoveryCodes(),
	}))
}

func (h handler) RenameWebAuthnCredential(w http.ResponseWriter, r *http.Request) {
	// Auth
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	// Input
	var body authHttp.RenameSecurityKeyBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Service
	out, err := h.ms.User.GRPC.User.RenameWebAuthnCredential(r.Context(), &userv1.RenameWebAuthnCredentialRequest{
		AccessToken:  token,
		CredentialId: strings.TrimSpace(r.PathValue("id")),
		Name:         body.Name,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.SecurityKeyFromPB(out.GetSecurityKey())))
}

func (h handler) DeleteWebAuthnCredential(w http.ResponseWriter, r *http.Request) {
	// Auth
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	// Input
	var body authHttp.CurrentPasswordBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Service
	_, err := h.ms.User.GRPC.User.DeleteWebAuthnCredential(r.Context(), &userv1.DeleteWebAuthnCredentialRequest{
		AccessToken:     token,
		CredentialId:    strings.TrimSpace(r.PathValue("id")),
		CurrentPassword: body.CurrentPassword,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.OKBody{OK: true}))
}

func (h handler) RegenerateRecoveryCodes(w http.ResponseWriter, r *http.Request) {
	// Auth
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	// Input
	var body authHttp.CurrentPasswordBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Service
	out, err := h.ms.User.GRPC.User.RegenerateRecoveryCodes(r.Context(), &userv1.RegenerateRecoveryCodesRequest{
		AccessToken:     token,
		CurrentPassword: body.CurrentPassword,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	codes := out.GetRecoveryCodes()
	if codes == nil {
		codes = []string{}
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.RecoveryCodesData{
		RecoveryCodes: codes,
	}))
}
