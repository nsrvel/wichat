package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	"github.com/wichat/wichat/backend/api-gateway/pkg/cookie"
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
	UploadAvatar(w http.ResponseWriter, r *http.Request)
	UpdateProfile(w http.ResponseWriter, r *http.Request)
	UpdateUsername(w http.ResponseWriter, r *http.Request)
	RequestEmailChange(w http.ResponseWriter, r *http.Request)
	VerifyEmailChange(w http.ResponseWriter, r *http.Request)
	CompleteEmailChange(w http.ResponseWriter, r *http.Request)
	UpdatePassword(w http.ResponseWriter, r *http.Request)
	GetMFAStatus(w http.ResponseWriter, r *http.Request)
	BeginTOTPSetup(w http.ResponseWriter, r *http.Request)
	ConfirmTOTPSetup(w http.ResponseWriter, r *http.Request)
	DisableTOTP(w http.ResponseWriter, r *http.Request)
	BeginWebAuthnRegister(w http.ResponseWriter, r *http.Request)
	FinishWebAuthnRegister(w http.ResponseWriter, r *http.Request)
	RenameWebAuthnCredential(w http.ResponseWriter, r *http.Request)
	DeleteWebAuthnCredential(w http.ResponseWriter, r *http.Request)
	RegenerateRecoveryCodes(w http.ResponseWriter, r *http.Request)
	DisableAccount(w http.ResponseWriter, r *http.Request)
	DeleteAccount(w http.ResponseWriter, r *http.Request)
	ListDevices(w http.ResponseWriter, r *http.Request)
	RevokeDevice(w http.ResponseWriter, r *http.Request)
	RevokeOtherDevices(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	ms               *microservices.Clients
	log              logger.Logger
	turnstileSiteKey string
}

// NewHandler wires user HTTP to the ms-user client.
func NewHandler(ms *microservices.Clients, log logger.Logger, turnstileSiteKey ...string) Handler {
	siteKey := ""
	if len(turnstileSiteKey) > 0 {
		siteKey = strings.TrimSpace(turnstileSiteKey[0])
	}
	return handler{ms: ms, log: log, turnstileSiteKey: siteKey}
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
		TurnstileSiteKey: h.turnstileSiteKey,
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

func (h handler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	// Auth
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	// Input
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, nil))
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, nil))
		return
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(io.LimitReader(file, (10<<20)+1))
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, nil))
		return
	}
	contentType := ""
	if header != nil {
		contentType = strings.TrimSpace(header.Header.Get("Content-Type"))
	}
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}

	// Service
	out, err := h.ms.User.GRPC.User.UploadUserAvatar(r.Context(), &userv1.UploadUserAvatarRequest{
		AccessToken: token,
		ContentType: contentType,
		Data:        data,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.UserFromPB(out.GetUser())))
}

func (h handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	// Auth
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	// Input
	var body authHttp.UpdateUserProfileBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Service
	out, err := h.ms.User.GRPC.User.UpdateUserProfile(r.Context(), &userv1.UpdateUserProfileRequest{
		AccessToken: token,
		DisplayName: body.DisplayName,
		Pronouns:    body.Pronouns,
		Bio:         body.Bio,
		BannerColor: body.BannerColor,
		ClearAvatar: body.ClearAvatar,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.UserFromPB(out.GetUser())))
}

func (h handler) UpdateUsername(w http.ResponseWriter, r *http.Request) {
	// Auth
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	// Input
	var body authHttp.UpdateUsernameBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Service
	out, err := h.ms.User.GRPC.User.UpdateUsername(r.Context(), &userv1.UpdateUsernameRequest{
		AccessToken:     token,
		Username:        body.Username,
		CurrentPassword: body.CurrentPassword,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.UserFromPB(out.GetUser())))
}

func (h handler) RequestEmailChange(w http.ResponseWriter, r *http.Request) {
	// Auth
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	// Input
	var body authHttp.RequestEmailChangeBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Service
	out, err := h.ms.User.GRPC.User.RequestEmailChange(r.Context(), &userv1.RequestEmailChangeRequest{
		AccessToken: token,
		Locale:      body.Locale,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.AuthChallengeBody{
		OTPRequired:        true,
		ChallengeID:        out.GetChallengeId(),
		EmailHint:          out.GetEmailHint(),
		ResendAfterSeconds: out.GetResendCooldownSeconds(),
	}))
}

func (h handler) VerifyEmailChange(w http.ResponseWriter, r *http.Request) {
	// Auth
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	// Input
	var body authHttp.VerifyEmailChangeBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Service
	out, err := h.ms.User.GRPC.User.VerifyEmailChange(r.Context(), &userv1.VerifyEmailChangeRequest{
		AccessToken: token,
		ChallengeId: body.ChallengeID,
		Code:        body.Code,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.VerifyEmailChangeData{
		ChangeToken: out.GetChangeToken(),
	}))
}

func (h handler) CompleteEmailChange(w http.ResponseWriter, r *http.Request) {
	// Auth
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	// Input
	var body authHttp.CompleteEmailChangeBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Service
	out, err := h.ms.User.GRPC.User.CompleteEmailChange(r.Context(), &userv1.CompleteEmailChangeRequest{
		AccessToken:     token,
		ChangeToken:     body.ChangeToken,
		NewEmail:        body.NewEmail,
		CurrentPassword: body.CurrentPassword,
		Reason:          body.Reason,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.UserFromPB(out.GetUser())))
}

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

func (h handler) DisableAccount(w http.ResponseWriter, r *http.Request) {
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
	if _, err := h.ms.User.GRPC.User.DisableAccount(r.Context(), &userv1.DisableAccountRequest{
		AccessToken:     token,
		CurrentPassword: body.CurrentPassword,
	}); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	cookie.ClearRefresh(w, r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https"))
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.OKBody{OK: true}))
}

func (h handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
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
	if _, err := h.ms.User.GRPC.User.DeleteAccount(r.Context(), &userv1.DeleteAccountRequest{
		AccessToken:     token,
		CurrentPassword: body.CurrentPassword,
	}); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	cookie.ClearRefresh(w, r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https"))
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.OKBody{OK: true}))
}

type deviceSessionItem struct {
	ID            string `json:"id"`
	Platform      string `json:"platform"`
	DeviceLabel   string `json:"device_label"`
	LocationLabel string `json:"location_label"`
	UserAgent     string `json:"user_agent"`
	LastUsedAt    string `json:"last_used_at"`
	CreatedAt     string `json:"created_at"`
	IsCurrent     bool   `json:"is_current"`
}

func (h handler) ListDevices(w http.ResponseWriter, r *http.Request) {
	// Auth
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	// Service
	out, err := h.ms.User.GRPC.User.ListSessions(r.Context(), &userv1.ListSessionsRequest{AccessToken: token})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	items := make([]deviceSessionItem, 0, len(out.GetSessions()))
	for _, s := range out.GetSessions() {
		items = append(items, deviceSessionItem{
			ID:            s.GetId(),
			Platform:      s.GetPlatform(),
			DeviceLabel:   s.GetDeviceLabel(),
			LocationLabel: s.GetLocationLabel(),
			UserAgent:     s.GetUserAgent(),
			LastUsedAt:    s.GetLastUsedAt(),
			CreatedAt:     s.GetCreatedAt(),
			IsCurrent:     s.GetIsCurrent(),
		})
	}
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(struct {
		Devices []deviceSessionItem `json:"devices"`
	}{Devices: items}))
}

func (h handler) RevokeDevice(w http.ResponseWriter, r *http.Request) {
	// Auth
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	// Input
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "id"}))
		return
	}
	var body authHttp.CurrentPasswordBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Service
	if _, err := h.ms.User.GRPC.User.RevokeSession(r.Context(), &userv1.RevokeSessionRequest{
		AccessToken:     token,
		SessionId:       id,
		CurrentPassword: body.CurrentPassword,
	}); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.OKBody{OK: true}))
}

func (h handler) RevokeOtherDevices(w http.ResponseWriter, r *http.Request) {
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
	if _, err := h.ms.User.GRPC.User.RevokeOtherSessions(r.Context(), &userv1.RevokeOtherSessionsRequest{
		AccessToken:     token,
		CurrentPassword: body.CurrentPassword,
	}); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.OKBody{OK: true}))
}
