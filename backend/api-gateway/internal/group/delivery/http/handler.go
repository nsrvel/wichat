package http

import (
	"io"
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

// Handler is the public group lifecycle & profile HTTP delivery.
type Handler interface {
	ListMyGroups(w http.ResponseWriter, r *http.Request)
	ReorderMyGroups(w http.ResponseWriter, r *http.Request)
	ListJoinableGroups(w http.ResponseWriter, r *http.Request)
	CreateGroup(w http.ResponseWriter, r *http.Request)
	DeleteGroup(w http.ResponseWriter, r *http.Request)
	UploadGroupIcon(w http.ResponseWriter, r *http.Request)
	UpdateGroupProfile(w http.ResponseWriter, r *http.Request)
	UploadGroupBanner(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	ms  *microservices.Clients
	log logger.Logger
}

// NewHandler wires group HTTP to the ms-user client.
func NewHandler(ms *microservices.Clients, log logger.Logger) Handler {
	return handler{ms: ms, log: log}
}

func (h handler) ListMyGroups(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	out, err := h.ms.User.GRPC.Group.ListMyGroups(r.Context(), &userv1.ListMyGroupsRequest{AccessToken: token})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.GroupsListBody{
		Groups: authHttp.GroupsFromPB(out.GetGroups()),
	}))
}

func (h handler) ReorderMyGroups(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	var body authHttp.ReorderMyGroupsBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}
	if len(body.GroupIDs) == 0 {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_ids"}))
		return
	}

	_, err := h.ms.User.GRPC.Group.ReorderMyGroups(r.Context(), &userv1.ReorderMyGroupsRequest{
		AccessToken: token,
		GroupIds:    body.GroupIDs,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, map[string]any{"data": authHttp.OKBody{OK: true}})
}

func (h handler) ListJoinableGroups(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	out, err := h.ms.User.GRPC.Group.ListJoinableGroups(r.Context(), &userv1.ListJoinableGroupsRequest{AccessToken: token})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.GroupsListBody{
		Groups: authHttp.GroupsFromPB(out.GetGroups()),
	}))
}

func (h handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	var body authHttp.CreateGroupBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	name := strings.TrimSpace(body.Name)
	if name == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "name"}))
		return
	}

	out, err := h.ms.User.GRPC.Group.CreateGroup(r.Context(), &userv1.CreateGroupRequest{
		AccessToken: token,
		Name:        name,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusCreated, response.NewDataEnvelope(authHttp.GroupFromPB(out.GetGroup())))
}

func (h handler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
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

	_, err := h.ms.User.GRPC.Group.DeleteGroup(r.Context(), &userv1.DeleteGroupRequest{
		AccessToken: token,
		GroupId:     groupID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.OKBody{OK: true}))
}

func (h handler) UploadGroupIcon(w http.ResponseWriter, r *http.Request) {
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

	r.Body = http.MaxBytesReader(w, r.Body, authHttp.MaxGroupIconBytes+(1<<20))
	if err := r.ParseMultipartForm(authHttp.MaxGroupIconBytes + (1 << 10)); err != nil {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "file"}))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "file"}))
		return
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(io.LimitReader(file, authHttp.MaxGroupIconBytes+1))
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "file"}))
		return
	}
	if len(data) == 0 || len(data) > authHttp.MaxGroupIconBytes {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "file"}))
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	out, err := h.ms.User.GRPC.Group.UploadGroupIcon(r.Context(), &userv1.UploadGroupIconRequest{
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
