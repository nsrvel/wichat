package http

import (
	"net/http"
	"net/url"
	"strings"

	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"github.com/wichat/wichat/backend/wi-shared/response"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

// Handler is the public invite HTTP delivery.
type Handler interface {
	GetGroupInviteLink(w http.ResponseWriter, r *http.Request)
	SendGroupInvites(w http.ResponseWriter, r *http.Request)
	PreviewGroupInvite(w http.ResponseWriter, r *http.Request)
	RedeemGroupInvite(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	ms        *microservices.Clients
	log       logger.Logger
	webOrigin string
}

// NewHandler wires invite HTTP to the ms-user client.
func NewHandler(ms *microservices.Clients, log logger.Logger, webOrigin string) Handler {
	return handler{
		ms:        ms,
		log:       log,
		webOrigin: strings.TrimRight(strings.TrimSpace(webOrigin), "/"),
	}
}

func (h handler) GetGroupInviteLink(w http.ResponseWriter, r *http.Request) {
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

	out, err := h.ms.User.GRPC.Invite.GetGroupInviteLink(r.Context(), &userv1.GetGroupInviteLinkRequest{
		AccessToken: token,
		GroupId:     groupID,
		ChannelId:   strings.TrimSpace(r.URL.Query().Get("channel_id")),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	inviteURL := h.webOrigin + "/invite/" + url.PathEscape(out.GetToken())

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.GroupInviteLinkBody{
		Token:     out.GetToken(),
		URL:       inviteURL,
		ExpiresAt: out.GetExpiresAt(),
	}))
}

func (h handler) SendGroupInvites(w http.ResponseWriter, r *http.Request) {
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

	var body authHttp.SendGroupInvitesBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	out, err := h.ms.User.GRPC.Invite.SendGroupInvites(r.Context(), &userv1.SendGroupInvitesRequest{
		AccessToken: token,
		GroupId:     groupID,
		ChannelId:   strings.TrimSpace(body.ChannelID),
		UserIds:     body.UserIDs,
		Emails:      body.Emails,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.SendGroupInvitesResultBody{
		Added:   int(out.GetAdded()),
		Emailed: int(out.GetEmailed()),
		Skipped: int(out.GetSkipped()),
	}))
}

func (h handler) PreviewGroupInvite(w http.ResponseWriter, r *http.Request) {
	inviteToken := strings.TrimSpace(r.URL.Query().Get("invite"))
	if inviteToken == "" {
		inviteToken = strings.TrimSpace(r.URL.Query().Get("token"))
	}
	if inviteToken == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "invite"}))
		return
	}

	out, err := h.ms.User.GRPC.Invite.PreviewGroupInvite(r.Context(), &userv1.PreviewGroupInviteRequest{
		Token: inviteToken,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.InvitePreviewBody{
		Valid:        out.GetValid(),
		GroupName:    out.GetGroupName(),
		ChannelName:  out.GetChannelName(),
		GroupLogoURL: out.GetGroupLogoUrl(),
		MemberCount:  int(out.GetMemberCount()),
		OnlineCount:  int(out.GetOnlineCount()),
		ExpiresAt:    out.GetExpiresAt(),
		Reason:       out.GetReason(),
	}))
}

func (h handler) RedeemGroupInvite(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	var body authHttp.RedeemInviteBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	out, err := h.ms.User.GRPC.Invite.RedeemGroupInvite(r.Context(), &userv1.RedeemGroupInviteRequest{
		AccessToken: token,
		Token:       strings.TrimSpace(body.Token),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.RedeemInviteResultBody{
		GroupID:   out.GetGroupId(),
		ChannelID: out.GetChannelId(),
	}))
}
