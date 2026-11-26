package http

import (
	"net/http"
	"strings"

	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"github.com/wichat/wichat/backend/wi-shared/response"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

func (h handler) ListChannelPermissions(w http.ResponseWriter, r *http.Request) {
	token, groupID, ok := h.requireBearerAndGroupID(w, r)
	if !ok {
		return
	}
	channelID := strings.TrimSpace(r.PathValue("channelId"))
	if channelID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "channelId"}))
		return
	}

	out, err := h.ms.User.GRPC.Channel.ListChannelPermissions(r.Context(), &userv1.ListChannelPermissionsRequest{
		AccessToken: token,
		GroupId:     groupID,
		ChannelId:   channelID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	items := make([]authHttp.ChannelCategoryPermissionOverrideBody, len(out.GetOverrides()))
	for i, o := range out.GetOverrides() {
		items[i] = authHttp.CategoryPermissionOverrideFromPB(o)
	}
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.CategoryPermissionsListResponse{
		Overrides: items,
	}))
}

func (h handler) UpsertChannelPermission(w http.ResponseWriter, r *http.Request) {
	token, groupID, ok := h.requireBearerAndGroupID(w, r)
	if !ok {
		return
	}
	channelID := strings.TrimSpace(r.PathValue("channelId"))
	if channelID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "channelId"}))
		return
	}

	var body authHttp.UpsertCategoryPermissionOverrideBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	out, err := h.ms.User.GRPC.Channel.UpsertChannelPermission(r.Context(), &userv1.UpsertChannelPermissionRequest{
		AccessToken:      token,
		GroupId:          groupID,
		ChannelId:        channelID,
		TargetType:       body.TargetType,
		TargetId:         body.TargetID,
		AllowPermissions: body.AllowPermissions,
		DenyPermissions:  body.DenyPermissions,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.CategoryPermissionOverrideResponse{
		Override: authHttp.CategoryPermissionOverrideFromPB(out.GetOverride()),
	}))
}

func (h handler) DeleteChannelPermission(w http.ResponseWriter, r *http.Request) {
	token, groupID, ok := h.requireBearerAndGroupID(w, r)
	if !ok {
		return
	}
	channelID := strings.TrimSpace(r.PathValue("channelId"))
	targetType := strings.TrimSpace(r.PathValue("targetType"))
	targetID := strings.TrimSpace(r.PathValue("targetId"))
	if channelID == "" || targetType == "" || targetID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, nil))
		return
	}

	_, err := h.ms.User.GRPC.Channel.DeleteChannelPermission(r.Context(), &userv1.DeleteChannelPermissionRequest{
		AccessToken: token,
		GroupId:     groupID,
		ChannelId:   channelID,
		TargetType:  targetType,
		TargetId:    targetID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.OKBody{OK: true}))
}

func (h handler) SyncChannelPermissions(w http.ResponseWriter, r *http.Request) {
	token, groupID, ok := h.requireBearerAndGroupID(w, r)
	if !ok {
		return
	}
	channelID := strings.TrimSpace(r.PathValue("channelId"))
	if channelID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "channelId"}))
		return
	}

	out, err := h.ms.User.GRPC.Channel.SyncChannelPermissions(r.Context(), &userv1.SyncChannelPermissionsRequest{
		AccessToken: token,
		GroupId:     groupID,
		ChannelId:   channelID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.ChannelResponse{
		Channel: authHttp.ChannelFromPB(out.GetChannel()),
	}))
}
