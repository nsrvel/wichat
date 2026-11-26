package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"github.com/wichat/wichat/backend/wi-shared/response"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

func (h handler) GetMyPresence(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	res, err := h.ms.User.GRPC.Presence.GetMyPresence(r.Context(), &userv1.GetMyPresenceRequest{
		AccessToken: token,
	})
	if err != nil {
		if utils.IsOutboundUnavailable(err) {
			h.log.Warn("get my presence degraded", "err", err)
			responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.DegradedMyPresenceBody()))
			return
		}
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.MyPresenceBody{
		ManualStatus:    res.GetManualStatus(),
		EffectiveStatus: res.GetEffectiveStatus(),
		IsOnline:        res.GetIsOnline(),
		ExpiresAt:       res.GetExpiresAt(),
	}))
}

func (h handler) UpdateMyPresence(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	var body authHttp.UpdateMyPresenceRequestBody
	if err := json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, nil))
		return
	}

	res, err := h.ms.User.GRPC.Presence.SetManualPresence(r.Context(), &userv1.SetManualPresenceRequest{
		AccessToken: token,
		Status:      strings.TrimSpace(body.Status),
		TtlSeconds:  body.TTLSeconds,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	if h.broadcaster != nil && len(res.GetGroupIds()) > 0 {
		if payload, mErr := json.Marshal(struct {
			Type           string `json:"type"`
			UserID         string `json:"user_id"`
			PresenceStatus string `json:"presence_status"`
			IsOnline       bool   `json:"is_online"`
		}{
			Type:           "presence.updated",
			UserID:         res.GetUserId(),
			PresenceStatus: res.GetEffectiveStatus(),
			IsOnline:       res.GetIsOnline(),
		}); mErr == nil {
			for _, gid := range res.GetGroupIds() {
				_ = h.broadcaster.DeliverGroup(gid, payload)
			}
		}
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.MyPresenceBody{
		ManualStatus:    res.GetManualStatus(),
		EffectiveStatus: res.GetEffectiveStatus(),
		IsOnline:        res.GetIsOnline(),
		ExpiresAt:       res.GetExpiresAt(),
	}))
}
