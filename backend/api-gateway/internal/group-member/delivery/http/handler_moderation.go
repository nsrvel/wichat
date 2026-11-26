package http

import (
	"net/http"
	"strconv"
	"strings"

	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"github.com/wichat/wichat/backend/wi-shared/response"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

func (h handler) BanGroupMembers(w http.ResponseWriter, r *http.Request) {
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

	var body banGroupMembersBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}
	if len(body.UserIDs) == 0 {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "user_ids"}))
		return
	}

	out, err := h.ms.User.GRPC.GroupMember.BanGroupMembers(r.Context(), &userv1.BanGroupMembersRequest{
		AccessToken:          token,
		GroupId:              groupID,
		UserIds:              body.UserIDs,
		Reason:               strings.TrimSpace(body.Reason),
		DeleteMessageSeconds: body.DeleteMessageSeconds,
		DurationSeconds:      body.DurationSeconds,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(banGroupMembersResponseBody{
		OK:          true,
		BannedCount: int(out.GetBannedCount()),
	}))
}

type timeoutGroupMemberBody struct {
	DurationSeconds int32  `json:"duration_seconds"`
	Reason          string `json:"reason"`
}

type timeoutGroupMemberResponseBody struct {
	TimeoutUntil string `json:"timeout_until"`
}

func (h handler) TimeoutGroupMember(w http.ResponseWriter, r *http.Request) {
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

	targetUserID := strings.TrimSpace(r.PathValue("userId"))
	if targetUserID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "userId"}))
		return
	}

	var body timeoutGroupMemberBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}
	if body.DurationSeconds < 0 {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "duration_seconds"}))
		return
	}

	out, err := h.ms.User.GRPC.GroupMember.TimeoutGroupMember(r.Context(), &userv1.TimeoutGroupMemberRequest{
		AccessToken:     token,
		GroupId:         groupID,
		UserId:          targetUserID,
		DurationSeconds: body.DurationSeconds,
		Reason:          strings.TrimSpace(body.Reason),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(timeoutGroupMemberResponseBody{
		TimeoutUntil: out.GetTimeoutUntil(),
	}))
}

type groupBanEntryBody struct {
	UserID              string `json:"user_id"`
	Username            string `json:"username"`
	DisplayName         string `json:"display_name"`
	AvatarURL           string `json:"avatar_url"`
	Reason              string `json:"reason"`
	BannedByUserID      string `json:"banned_by_user_id"`
	BannedByUsername    string `json:"banned_by_username"`
	BannedByDisplayName string `json:"banned_by_display_name"`
	ExpiresAt           string `json:"expires_at"`
	CreatedAt           string `json:"created_at"`
}

type listGroupBansResponseBody struct {
	Bans       []groupBanEntryBody `json:"bans"`
	NextCursor string              `json:"next_cursor,omitempty"`
	Total      int                 `json:"total"`
}

func (h handler) ListGroupBans(w http.ResponseWriter, r *http.Request) {
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

	var limit int32
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		n, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || n < 1 || n > constants.MaxMemberListLimit {
			authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "limit"}))
			return
		}
		limit = int32(n) //nolint:gosec // G115 — bounded by MaxMemberListLimit
	}

	out, err := h.ms.User.GRPC.GroupMember.ListGroupBans(r.Context(), &userv1.ListGroupBansRequest{
		AccessToken: token,
		GroupId:     groupID,
		Search:      strings.TrimSpace(r.URL.Query().Get("search")),
		Limit:       limit,
		Cursor:      strings.TrimSpace(r.URL.Query().Get("cursor")),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	bans := make([]groupBanEntryBody, 0, len(out.GetBans()))
	for _, b := range out.GetBans() {
		bans = append(bans, groupBanEntryBody{
			UserID:              b.GetUserId(),
			Username:            b.GetUsername(),
			DisplayName:         b.GetDisplayName(),
			AvatarURL:           b.GetAvatarUrl(),
			Reason:              b.GetReason(),
			BannedByUserID:      b.GetBannedByUserId(),
			BannedByUsername:    b.GetBannedByUsername(),
			BannedByDisplayName: b.GetBannedByDisplayName(),
			ExpiresAt:           b.GetExpiresAt(),
			CreatedAt:           b.GetCreatedAt(),
		})
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(listGroupBansResponseBody{
		Bans:       bans,
		NextCursor: out.GetNextCursor(),
		Total:      int(out.GetTotal()),
	}))
}
