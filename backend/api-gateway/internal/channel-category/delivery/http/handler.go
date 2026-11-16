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

// Handler is the public channel category HTTP delivery.
type Handler interface {
	CreateGroupCategory(w http.ResponseWriter, r *http.Request)
	UpdateGroupCategory(w http.ResponseWriter, r *http.Request)
	DeleteGroupCategory(w http.ResponseWriter, r *http.Request)
	ListCategoryPermissions(w http.ResponseWriter, r *http.Request)
	UpsertCategoryPermission(w http.ResponseWriter, r *http.Request)
	DeleteCategoryPermission(w http.ResponseWriter, r *http.Request)
	ReorderChannelCategories(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	ms  *microservices.Clients
	log logger.Logger
}

// NewHandler constructs channel category HTTP handlers.
func NewHandler(ms *microservices.Clients, log logger.Logger) Handler {
	return handler{ms: ms, log: log}
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

func (h handler) CreateGroupCategory(w http.ResponseWriter, r *http.Request) {
	token, groupID, ok := h.requireBearerAndGroupID(w, r)
	if !ok {
		return
	}

	var body authHttp.CreateChannelCategoryBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	out, err := h.ms.User.GRPC.ChannelCategory.CreateChannelCategory(r.Context(), &userv1.CreateChannelCategoryRequest{
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

func (h handler) UpdateGroupCategory(w http.ResponseWriter, r *http.Request) {
	token, groupID, ok := h.requireBearerAndGroupID(w, r)
	if !ok {
		return
	}

	categoryID := strings.TrimSpace(r.PathValue("categoryId"))
	if categoryID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "categoryId"}))
		return
	}

	var body authHttp.UpdateChannelCategoryBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	out, err := h.ms.User.GRPC.ChannelCategory.UpdateChannelCategory(r.Context(), &userv1.UpdateChannelCategoryRequest{
		AccessToken: token,
		GroupId:     groupID,
		CategoryId:  categoryID,
		Name:        body.Name,
		IsPrivate:   body.IsPrivate,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.ChannelCategoryResponse{
		Category: authHttp.ChannelCategoryFromPB(out.GetCategory()),
	}))
}

func (h handler) DeleteGroupCategory(w http.ResponseWriter, r *http.Request) {
	token, groupID, ok := h.requireBearerAndGroupID(w, r)
	if !ok {
		return
	}

	categoryID := strings.TrimSpace(r.PathValue("categoryId"))
	if categoryID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "categoryId"}))
		return
	}

	_, err := h.ms.User.GRPC.ChannelCategory.DeleteChannelCategory(r.Context(), &userv1.DeleteChannelCategoryRequest{
		AccessToken: token,
		GroupId:     groupID,
		CategoryId:  categoryID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.OKBody{OK: true}))
}

func (h handler) ListCategoryPermissions(w http.ResponseWriter, r *http.Request) {
	token, groupID, ok := h.requireBearerAndGroupID(w, r)
	if !ok {
		return
	}
	categoryID := strings.TrimSpace(r.PathValue("categoryId"))
	if categoryID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "categoryId"}))
		return
	}

	out, err := h.ms.User.GRPC.ChannelCategory.ListChannelCategoryPermissions(r.Context(), &userv1.ListChannelCategoryPermissionsRequest{
		AccessToken: token,
		GroupId:     groupID,
		CategoryId:  categoryID,
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

func (h handler) UpsertCategoryPermission(w http.ResponseWriter, r *http.Request) {
	token, groupID, ok := h.requireBearerAndGroupID(w, r)
	if !ok {
		return
	}
	categoryID := strings.TrimSpace(r.PathValue("categoryId"))
	if categoryID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "categoryId"}))
		return
	}

	var body authHttp.UpsertCategoryPermissionOverrideBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	out, err := h.ms.User.GRPC.ChannelCategory.UpsertChannelCategoryPermission(r.Context(), &userv1.UpsertChannelCategoryPermissionRequest{
		AccessToken:      token,
		GroupId:          groupID,
		CategoryId:       categoryID,
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

func (h handler) DeleteCategoryPermission(w http.ResponseWriter, r *http.Request) {
	token, groupID, ok := h.requireBearerAndGroupID(w, r)
	if !ok {
		return
	}
	categoryID := strings.TrimSpace(r.PathValue("categoryId"))
	targetType := strings.TrimSpace(r.PathValue("targetType"))
	targetID := strings.TrimSpace(r.PathValue("targetId"))
	if categoryID == "" || targetType == "" || targetID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, nil))
		return
	}

	_, err := h.ms.User.GRPC.ChannelCategory.DeleteChannelCategoryPermission(r.Context(), &userv1.DeleteChannelCategoryPermissionRequest{
		AccessToken: token,
		GroupId:     groupID,
		CategoryId:  categoryID,
		TargetType:  targetType,
		TargetId:    targetID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.OKBody{OK: true}))
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

	_, err := h.ms.User.GRPC.ChannelCategory.ReorderChannelCategories(r.Context(), &userv1.ReorderChannelCategoriesRequest{
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
