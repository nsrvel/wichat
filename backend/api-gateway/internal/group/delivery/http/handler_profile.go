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

func (h handler) UpdateGroupProfile(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	groupID := strings.TrimSpace(r.PathValue("id"))
	if groupID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "id"}))
		return
	}

	var body authHttp.UpdateGroupProfileBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		return
	}

	out, err := h.ms.User.GRPC.Group.UpdateGroupProfile(r.Context(), &userv1.UpdateGroupProfileRequest{
		AccessToken:      token,
		GroupId:          groupID,
		Name:             body.Name,
		Description:      body.Description,
		IsPrivate:        body.IsPrivate,
		BannerColor:      body.BannerColor,
		ClearLogo:        body.ClearLogo,
		ClearBannerImage: body.ClearBannerImage,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.GroupFromPB(out.GetGroup())))
}

func (h handler) UploadGroupBanner(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	groupID := strings.TrimSpace(r.PathValue("id"))
	if groupID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "id"}))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, authHttp.MaxGroupBannerBytes+(1<<20))
	if err := r.ParseMultipartForm(authHttp.MaxGroupBannerBytes + (1 << 10)); err != nil {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "file"}))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "file"}))
		return
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(io.LimitReader(file, authHttp.MaxGroupBannerBytes+1))
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "file"}))
		return
	}
	if len(data) == 0 || len(data) > authHttp.MaxGroupBannerBytes {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "file"}))
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	out, err := h.ms.User.GRPC.Group.UploadGroupBanner(r.Context(), &userv1.UploadGroupBannerRequest{
		AccessToken: token,
		GroupId:     groupID,
		ContentType: contentType,
		Data:        data,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.GroupFromPB(out.GetGroup())))
}

func (h handler) ListGroupAuditLogs(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}
	groupID := strings.TrimSpace(r.PathValue("id"))
	if groupID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "id"}))
		return
	}

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

	out, err := h.ms.User.GRPC.Group.ListGroupActivityLogs(r.Context(), &userv1.ListGroupActivityLogsRequest{
		AccessToken: token,
		GroupId:     groupID,
		ActorId:     strings.TrimSpace(q.Get("actor_id")),
		Category:    strings.TrimSpace(q.Get("category")),
		Action:      strings.TrimSpace(q.Get("action")),
		Limit:       limit,
		Cursor:      strings.TrimSpace(q.Get("cursor")),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(struct {
		Logs       []authHttp.GroupActivityLogBody `json:"logs"`
		NextCursor string                          `json:"next_cursor"`
	}{
		Logs:       authHttp.GroupActivityLogsFromPB(out.GetLogs()),
		NextCursor: out.GetNextCursor(),
	}))
}
