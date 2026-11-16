package http

import (
	"net/http"
	"strings"

	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"github.com/wichat/wichat/backend/wi-shared/response"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

// Handler is the public channel HTTP delivery.
type Handler interface {
	ListGroupChannels(w http.ResponseWriter, r *http.Request)
	CreateGroupChannel(w http.ResponseWriter, r *http.Request)
	UpdateGroupChannel(w http.ResponseWriter, r *http.Request)
	DeleteGroupChannel(w http.ResponseWriter, r *http.Request)
	ListChannelPermissions(w http.ResponseWriter, r *http.Request)
	UpsertChannelPermission(w http.ResponseWriter, r *http.Request)
	DeleteChannelPermission(w http.ResponseWriter, r *http.Request)
	SyncChannelPermissions(w http.ResponseWriter, r *http.Request)
	ReorderGroupChannels(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	ms  *microservices.Clients
	log logger.Logger
}

// NewHandler wires channel HTTP to the ms-user client.
func NewHandler(ms *microservices.Clients, log logger.Logger) Handler {
	return handler{ms: ms, log: log}
}

func (h handler) ListGroupChannels(w http.ResponseWriter, r *http.Request) {
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

	out, err := h.ms.User.GRPC.Channel.ListGroupChannels(r.Context(), &userv1.ListGroupChannelsRequest{
		AccessToken: token,
		GroupId:     groupID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.ChannelsListBody{
		Channels:   authHttp.ChannelsFromPB(out.GetChannels()),
		Categories: authHttp.ChannelCategoriesFromPB(out.GetCategories()),
	}))
}

func (h handler) CreateGroupChannel(w http.ResponseWriter, r *http.Request) {
	token, groupID, ok := h.requireBearerAndGroupID(w, r)
	if !ok {
		return
	}

	var body authHttp.CreateGroupChannelBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	var categoryID string
	if body.CategoryID != nil {
		categoryID = *body.CategoryID
	}

	out, err := h.ms.User.GRPC.Channel.CreateGroupChannel(r.Context(), &userv1.CreateGroupChannelRequest{
		AccessToken: token,
		GroupId:     groupID,
		Name:        body.Name,
		Type:        body.Type,
		CategoryId:  categoryID,
		IsPrivate:   body.IsPrivate,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusCreated, response.NewDataEnvelope(authHttp.ChannelResponse{
		Channel: authHttp.ChannelFromPB(out.GetChannel()),
	}))
}

func (h handler) ReorderGroupChannels(w http.ResponseWriter, r *http.Request) {
	token, groupID, ok := h.requireBearerAndGroupID(w, r)
	if !ok {
		return
	}

	var body authHttp.ReorderGroupChannelsBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}
	if len(body.ChannelIDs) == 0 {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "channel_ids"}))
		return
	}

	_, err := h.ms.User.GRPC.Channel.ReorderGroupChannels(r.Context(), &userv1.ReorderGroupChannelsRequest{
		AccessToken:    token,
		GroupId:        groupID,
		ChannelIds:     body.ChannelIDs,
		CategoryId:     body.CategoryID,
		SidebarSection: body.SidebarSection,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.OKBody{OK: true}))
}

func (h handler) UpdateGroupChannel(w http.ResponseWriter, r *http.Request) {
	token, groupID, ok := h.requireBearerAndGroupID(w, r)
	if !ok {
		return
	}

	channelID := strings.TrimSpace(r.PathValue("channelId"))
	if channelID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "channelId"}))
		return
	}

	var body authHttp.UpdateGroupChannelBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	out, err := h.ms.User.GRPC.Channel.UpdateGroupChannel(r.Context(), &userv1.UpdateGroupChannelRequest{
		AccessToken:           token,
		GroupId:               groupID,
		ChannelId:             channelID,
		Name:                  body.Name,
		IsPrivate:             body.IsPrivate,
		SlowmodeSeconds:       body.SlowmodeSeconds,
		ThreadSlowmodeSeconds: body.ThreadSlowmodeSeconds,
		ContentVisibility:     body.ContentVisibility,
		ForumDefaultLayout:    body.ForumDefaultLayout,
		ForumSortOrder:        body.ForumSortOrder,
		ForumTagMatching:      body.ForumTagMatching,
		BitrateKbps:           body.BitrateKbps,
		VideoQuality:          body.VideoQuality,
		UserLimit:             body.UserLimit,
		RtcRegion:             body.RTCRegion,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.ChannelResponse{
		Channel: authHttp.ChannelFromPB(out.GetChannel()),
	}))
}

func (h handler) DeleteGroupChannel(w http.ResponseWriter, r *http.Request) {
	token, groupID, ok := h.requireBearerAndGroupID(w, r)
	if !ok {
		return
	}

	channelID := strings.TrimSpace(r.PathValue("channelId"))
	if channelID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "channelId"}))
		return
	}

	_, err := h.ms.User.GRPC.Channel.DeleteGroupChannel(r.Context(), &userv1.DeleteGroupChannelRequest{
		AccessToken: token,
		GroupId:     groupID,
		ChannelId:   channelID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.OKBody{OK: true}))
}

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

func (h handler) requireBearerAndGroupID(w http.ResponseWriter, r *http.Request) (string, string, bool) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return "", "", false
	}

	groupID := strings.TrimSpace(r.PathValue("id"))
	if groupID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "id"}))
		return "", "", false
	}

	return token, groupID, true
}
