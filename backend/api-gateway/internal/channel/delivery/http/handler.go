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
	CreateGroupCategory(w http.ResponseWriter, r *http.Request)
	ReorderChannelCategories(w http.ResponseWriter, r *http.Request)
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

func (h handler) CreateGroupCategory(w http.ResponseWriter, r *http.Request) {
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

	var body authHttp.CreateChannelCategoryBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	out, err := h.ms.User.GRPC.Channel.CreateChannelCategory(r.Context(), &userv1.CreateChannelCategoryRequest{
		AccessToken: token,
		GroupId:     groupID,
		Name:        body.Name,
		IsPrivate:   body.IsPrivate,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusCreated, response.NewDataEnvelope(authHttp.ChannelCategoryResponse{
		Category: authHttp.ChannelCategoryFromPB(out.GetCategory()),
	}))
}

func (h handler) ReorderChannelCategories(w http.ResponseWriter, r *http.Request) {
	token, groupID, ok := h.requireBearerAndGroupID(w, r)
	if !ok {
		return
	}

	var body authHttp.ReorderChannelCategoriesBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}
	if len(body.CategoryIDs) == 0 {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "category_ids"}))
		return
	}

	_, err := h.ms.User.GRPC.Channel.ReorderChannelCategories(r.Context(), &userv1.ReorderChannelCategoriesRequest{
		AccessToken: token,
		GroupId:     groupID,
		CategoryIds: body.CategoryIDs,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.OKBody{OK: true}))
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
