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

func (h handler) ListRelationships(w http.ResponseWriter, r *http.Request) {
	// Auth
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	// Service
	out, err := h.ms.User.GRPC.User.ListUserRelationships(r.Context(), &userv1.ListUserRelationshipsRequest{
		AccessToken: token,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	items := make([]authHttp.UserRelationshipBody, 0, len(out.GetRelationships()))
	for _, rel := range out.GetRelationships() {
		items = append(items, authHttp.UserRelationshipBody{
			TargetUserID: rel.GetTargetUserId(),
			Kind:         rel.GetKind(),
			CreatedAt:    rel.GetCreatedAt(),
		})
	}
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(struct {
		Relationships []authHttp.UserRelationshipBody `json:"relationships"`
	}{Relationships: items}))
}

func (h handler) SetRelationship(w http.ResponseWriter, r *http.Request) {
	// Auth
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	// Input
	targetUserID := strings.TrimSpace(r.PathValue("userId"))
	if targetUserID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "userId"}))
		return
	}
	var body authHttp.SetUserRelationshipBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Service
	out, err := h.ms.User.GRPC.User.SetUserRelationship(r.Context(), &userv1.SetUserRelationshipRequest{
		AccessToken:  token,
		TargetUserId: targetUserID,
		Kind:         body.Kind,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.SetUserRelationshipResponseBody{
		OK:   out.GetOk(),
		Kind: out.GetKind(),
	}))
}
