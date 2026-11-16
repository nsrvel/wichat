package http

import (
	"net/http"
	"strconv"
	"strings"

	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"github.com/wichat/wichat/backend/wi-shared/response"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

// Handler is the public group-member HTTP delivery.
type Handler interface {
	JoinGroup(w http.ResponseWriter, r *http.Request)
	LeaveGroup(w http.ResponseWriter, r *http.Request)
	ListGroupMembers(w http.ResponseWriter, r *http.Request)
	KickGroupMember(w http.ResponseWriter, r *http.Request)
	BanGroupMembers(w http.ResponseWriter, r *http.Request)
	ListGroupBans(w http.ResponseWriter, r *http.Request)
	UnbanGroupMember(w http.ResponseWriter, r *http.Request)
	TimeoutGroupMember(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	ms  *microservices.Clients
	log logger.Logger
}

// NewHandler wires group-member HTTP to the ms-user client.
func NewHandler(ms *microservices.Clients, log logger.Logger) Handler {
	return handler{ms: ms, log: log}
}

func (h handler) JoinGroup(w http.ResponseWriter, r *http.Request) {
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

	_, err := h.ms.User.GRPC.GroupMember.JoinGroup(r.Context(), &userv1.JoinGroupRequest{
		AccessToken: token,
		GroupId:     groupID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.OKBody{OK: true}))
}

func (h handler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
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

	_, err := h.ms.User.GRPC.GroupMember.LeaveGroup(r.Context(), &userv1.LeaveGroupRequest{
		AccessToken: token,
		GroupId:     groupID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.OKBody{OK: true}))
}

func (h handler) ListGroupMembers(w http.ResponseWriter, r *http.Request) {
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

	limit := 20
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}

	out, err := h.ms.User.GRPC.GroupMember.ListGroupMembers(r.Context(), &userv1.ListGroupMembersRequest{
		AccessToken: token,
		GroupId:     groupID,
		Limit:       int32(limit), //nolint:gosec // G115 — query limit clamped above
		Cursor:      strings.TrimSpace(r.URL.Query().Get("cursor")),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.GroupMembersPageBody{
		Members:    authHttp.GroupMembersFromPB(out.GetMembers()),
		NextCursor: out.GetNextCursor(),
		Total:      int(out.GetTotal()),
	}))
}

type kickGroupMemberBody struct {
	Reason string `json:"reason"`
}

type banGroupMembersBody struct {
	UserIDs              []string `json:"user_ids"`
	Reason               string   `json:"reason"`
	DeleteMessageSeconds int32    `json:"delete_message_seconds"`
	DurationSeconds      int64    `json:"duration_seconds"`
}

type banGroupMembersResponseBody struct {
	OK          bool `json:"ok"`
	BannedCount int  `json:"banned_count"`
}

func (h handler) KickGroupMember(w http.ResponseWriter, r *http.Request) {
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

	var body kickGroupMemberBody
	if r.ContentLength > 0 {
		if err := authHttp.DecodeJSON(w, r, &body); err != nil {
			authHttp.WriteOutbound(w, r, h.log, err)
			return
		}
	}

	_, err := h.ms.User.GRPC.GroupMember.KickGroupMember(r.Context(), &userv1.KickGroupMemberRequest{
		AccessToken: token,
		GroupId:     groupID,
		UserId:      targetUserID,
		Reason:      strings.TrimSpace(body.Reason),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.OKBody{OK: true}))
}

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

func (h handler) UnbanGroupMember(w http.ResponseWriter, r *http.Request) {
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

	_, err := h.ms.User.GRPC.GroupMember.UnbanGroupMember(r.Context(), &userv1.UnbanGroupMemberRequest{
		AccessToken: token,
		GroupId:     groupID,
		UserId:      targetUserID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.OKBody{OK: true}))
}
