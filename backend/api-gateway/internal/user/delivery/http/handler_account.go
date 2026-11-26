package http

import (
	"net/http"
	"strconv"
	"strings"

	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	"github.com/wichat/wichat/backend/api-gateway/pkg/cookie"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"github.com/wichat/wichat/backend/wi-shared/response"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

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

func (h handler) ListActivityLogs(w http.ResponseWriter, r *http.Request) {
	// Auth
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	// Input
	q := r.URL.Query()
	var limit int32
	if rawLimit := strings.TrimSpace(q.Get("limit")); rawLimit != "" {
		parsed, err := strconv.ParseInt(rawLimit, 10, 32)
		if err != nil || parsed <= 0 {
			authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "limit"}))
			return
		}
		limit = int32(parsed)
	}

	// Service
	res, err := h.ms.User.GRPC.User.ListUserActivityLogs(r.Context(), &userv1.ListUserActivityLogsRequest{
		AccessToken: token,
		Category:    strings.TrimSpace(q.Get("category")),
		Action:      strings.TrimSpace(q.Get("action")),
		Limit:       limit,
		Cursor:      strings.TrimSpace(q.Get("cursor")),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(struct {
		Logs       []authHttp.UserActivityLogBody `json:"logs"`
		NextCursor string                         `json:"next_cursor,omitempty"`
	}{
		Logs:       authHttp.UserActivityLogsFromPB(res.GetLogs()),
		NextCursor: res.GetNextCursor(),
	}))
}
