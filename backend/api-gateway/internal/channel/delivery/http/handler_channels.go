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
