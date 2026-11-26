package http

import (
	"net/http"
	"strconv"
	"strings"

	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"github.com/wichat/wichat/backend/wi-shared/response"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

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
		if utils.IsOutboundUnavailable(err) {
			h.log.Warn("list group members degraded", "err", err, "group_id", groupID)
			responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.DegradedGroupMembersPageBody()))
			return
		}
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
