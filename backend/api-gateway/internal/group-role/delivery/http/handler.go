package http

import (
	"net/http"

	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"github.com/wichat/wichat/backend/wi-shared/response"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

// Handler is the HTTP delivery interface for group roles and permissions.
type Handler interface {
	ListGroupRoles(w http.ResponseWriter, r *http.Request)
	CreateGroupRole(w http.ResponseWriter, r *http.Request)
	UpdateGroupRole(w http.ResponseWriter, r *http.Request)
	DuplicateGroupRole(w http.ResponseWriter, r *http.Request)
	DeleteGroupRole(w http.ResponseWriter, r *http.Request)
	ListGroupRoleMembers(w http.ResponseWriter, r *http.Request)
	AddGroupRoleMembers(w http.ResponseWriter, r *http.Request)
	RemoveGroupRoleMember(w http.ResponseWriter, r *http.Request)
	UpdateGroupHideRoles(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	ms  *microservices.Clients
	log logger.Logger
}

// NewHandler wires group-role HTTP handlers to the ms-user client.
func NewHandler(ms *microservices.Clients, log logger.Logger) Handler {
	return handler{ms: ms, log: log}
}

func (h handler) requireBearer(w http.ResponseWriter, r *http.Request) (string, bool) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return "", false
	}
	return token, true
}

func (h handler) ListGroupRoles(w http.ResponseWriter, r *http.Request) {
	token, ok := h.requireBearer(w, r)
	if !ok {
		return
	}

	out, err := h.ms.User.GRPC.GroupRole.ListGroupRoles(r.Context(), &userv1.ListGroupRolesRequest{
		AccessToken: token,
		GroupId:     r.PathValue("id"),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.GroupRolesListBody{
		Roles:     authHttp.GroupRolesFromPB(out.GetRoles()),
		HideRoles: out.GetHideRoles(),
	}))
}

func (h handler) CreateGroupRole(w http.ResponseWriter, r *http.Request) {
	token, ok := h.requireBearer(w, r)
	if !ok {
		return
	}

	var body authHttp.CreateGroupRoleBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	out, err := h.ms.User.GRPC.GroupRole.CreateGroupRole(r.Context(), &userv1.CreateGroupRoleRequest{
		AccessToken: token,
		GroupId:     r.PathValue("id"),
		Name:        body.Name,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusCreated, response.NewDataEnvelope(authHttp.GroupRoleResponseBody{
		Role: authHttp.GroupRoleFromPB(out.GetRole()),
	}))
}

func (h handler) UpdateGroupRole(w http.ResponseWriter, r *http.Request) {
	token, ok := h.requireBearer(w, r)
	if !ok {
		return
	}

	var body authHttp.UpdateGroupRoleBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	req := &userv1.UpdateGroupRoleRequest{
		AccessToken: token,
		GroupId:     r.PathValue("id"),
		RoleId:      r.PathValue("roleId"),
	}
	if body.Name != nil {
		req.Name = *body.Name
	}
	if body.Color != nil {
		req.Color = *body.Color
	}
	if body.Hoist != nil {
		req.Hoist = *body.Hoist
	}
	if body.Mentionable != nil {
		req.Mentionable = *body.Mentionable
	}
	if body.Permissions != nil {
		req.Permissions = body.Permissions
		if len(body.Permissions) == 0 {
			req.ClearPermissions = true
		}
	}

	out, err := h.ms.User.GRPC.GroupRole.UpdateGroupRole(r.Context(), req)
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.GroupRoleResponseBody{
		Role: authHttp.GroupRoleFromPB(out.GetRole()),
	}))
}

func (h handler) DuplicateGroupRole(w http.ResponseWriter, r *http.Request) {
	token, ok := h.requireBearer(w, r)
	if !ok {
		return
	}

	out, err := h.ms.User.GRPC.GroupRole.DuplicateGroupRole(r.Context(), &userv1.DuplicateGroupRoleRequest{
		AccessToken: token,
		GroupId:     r.PathValue("id"),
		RoleId:      r.PathValue("roleId"),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusCreated, response.NewDataEnvelope(authHttp.GroupRoleResponseBody{
		Role: authHttp.GroupRoleFromPB(out.GetRole()),
	}))
}

func (h handler) DeleteGroupRole(w http.ResponseWriter, r *http.Request) {
	token, ok := h.requireBearer(w, r)
	if !ok {
		return
	}

	_, err := h.ms.User.GRPC.GroupRole.DeleteGroupRole(r.Context(), &userv1.DeleteGroupRoleRequest{
		AccessToken: token,
		GroupId:     r.PathValue("id"),
		RoleId:      r.PathValue("roleId"),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.OKBody{OK: true}))
}

func (h handler) ListGroupRoleMembers(w http.ResponseWriter, r *http.Request) {
	token, ok := h.requireBearer(w, r)
	if !ok {
		return
	}

	out, err := h.ms.User.GRPC.GroupRole.ListGroupRoleMembers(r.Context(), &userv1.ListGroupRoleMembersRequest{
		AccessToken: token,
		GroupId:     r.PathValue("id"),
		RoleId:      r.PathValue("roleId"),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.GroupRoleMembersBody{
		Members: authHttp.GroupMembersFromPB(out.GetMembers()),
	}))
}

func (h handler) AddGroupRoleMembers(w http.ResponseWriter, r *http.Request) {
	token, ok := h.requireBearer(w, r)
	if !ok {
		return
	}

	var body authHttp.AddGroupRoleMembersBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	out, err := h.ms.User.GRPC.GroupRole.AddGroupRoleMembers(r.Context(), &userv1.AddGroupRoleMembersRequest{
		AccessToken: token,
		GroupId:     r.PathValue("id"),
		RoleId:      r.PathValue("roleId"),
		UserIds:     body.UserIDs,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.GroupRoleResponseBody{
		Role: authHttp.GroupRoleFromPB(out.GetRole()),
	}))
}

func (h handler) RemoveGroupRoleMember(w http.ResponseWriter, r *http.Request) {
	token, ok := h.requireBearer(w, r)
	if !ok {
		return
	}

	out, err := h.ms.User.GRPC.GroupRole.RemoveGroupRoleMember(r.Context(), &userv1.RemoveGroupRoleMemberRequest{
		AccessToken: token,
		GroupId:     r.PathValue("id"),
		RoleId:      r.PathValue("roleId"),
		UserId:      r.PathValue("userId"),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.GroupRoleResponseBody{
		Role: authHttp.GroupRoleFromPB(out.GetRole()),
	}))
}

func (h handler) UpdateGroupHideRoles(w http.ResponseWriter, r *http.Request) {
	token, ok := h.requireBearer(w, r)
	if !ok {
		return
	}

	var body authHttp.UpdateGroupHideRolesBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	out, err := h.ms.User.GRPC.GroupRole.UpdateGroupHideRoles(r.Context(), &userv1.UpdateGroupHideRolesRequest{
		AccessToken: token,
		GroupId:     r.PathValue("id"),
		HideRoles:   body.HideRoles,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.UpdateGroupHideRolesBody{
		HideRoles: out.GetHideRoles(),
	}))
}
