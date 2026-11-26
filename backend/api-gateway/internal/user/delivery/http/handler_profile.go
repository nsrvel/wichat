package http

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"github.com/wichat/wichat/backend/wi-shared/response"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

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

func (h handler) UpdateStatusNote(w http.ResponseWriter, r *http.Request) {
	// Auth
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	// Input
	var body authHttp.UpdateUserStatusNoteBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}
	expiresAt := ""
	if body.ExpiresAt != nil {
		expiresAt = *body.ExpiresAt
	}

	// Service
	out, err := h.ms.User.GRPC.User.UpdateUserStatusNote(r.Context(), &userv1.UpdateUserStatusNoteRequest{
		AccessToken: token,
		Text:        body.Text,
		Emoji:       body.Emoji,
		ExpiresAt:   expiresAt,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.UserFromPB(out.GetUser())))
}

func (h handler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	// Auth
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	// Input
	var body authHttp.UpdateUserPreferencesBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Service
	out, err := h.ms.User.GRPC.User.UpdateUserPreferences(r.Context(), &userv1.UpdateUserPreferencesRequest{
		AccessToken: token,
		Locale:      body.Locale,
		TimeFormat:  body.TimeFormat,
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
