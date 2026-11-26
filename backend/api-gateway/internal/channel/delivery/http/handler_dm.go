package http

import (
	"net/http"
	"strings"

	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"github.com/wichat/wichat/backend/wi-shared/realtime"
	"github.com/wichat/wichat/backend/wi-shared/response"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

func (h handler) EnsureGroupDmChannel(w http.ResponseWriter, r *http.Request) {
	token, groupID, ok := h.requireBearerAndGroupID(w, r)
	if !ok {
		return
	}

	var body authHttp.EnsureGroupDmBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	out, err := h.ms.User.GRPC.Channel.EnsureGroupDmChannel(r.Context(), &userv1.EnsureGroupDmChannelRequest{
		AccessToken: token,
		GroupId:     groupID,
		PeerUserId:  strings.TrimSpace(body.UserID),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	status := http.StatusOK
	if out.GetCreated() {
		status = http.StatusCreated
		if h.broadcaster != nil {
			ch := out.GetChannel()
			if payload, mErr := realtime.MarshalChannelsUpdated(
				groupID,
				realtime.ChannelsUpdateReasonDmCreated,
				ch.GetId(),
			); mErr == nil {
				_ = h.broadcaster.DeliverGroup(groupID, payload)
			}
		}
	}

	responsehttp.WriteJSON(w, status, response.NewDataEnvelope(authHttp.EnsureGroupDmResponse{
		Channel: authHttp.ChannelFromPB(out.GetChannel()),
		Created: out.GetCreated(),
	}))
}
