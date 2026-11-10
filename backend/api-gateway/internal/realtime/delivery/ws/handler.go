package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/coder/websocket"
	"github.com/wichat/wichat/backend/api-gateway/internal/realtime/hub"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

var publicEntityID = regexp.MustCompile(`^[0-9a-f]{24}$`)

type clientMessage struct {
	Type      string `json:"type"`
	GroupID   string `json:"group_id"`
	ChannelID string `json:"channel_id"`
}

// Handler is the chat realtime WebSocket delivery.
type Handler struct {
	ms        *microservices.Clients
	hub       *hub.Hub
	log       logger.Logger
	webOrigin string
}

// NewHandler constructs realtime WebSocket handlers.
func NewHandler(ms *microservices.Clients, rtHub *hub.Hub, log logger.Logger, webOrigin string) Handler {
	return Handler{ms: ms, hub: rtHub, log: log, webOrigin: webOrigin}
}

func (h Handler) Realtime(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.URL.Query().Get("access_token"))
	if token == "" {
		authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
		const prefix = "Bearer "
		if strings.HasPrefix(authHeader, prefix) {
			token = strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
		}
	}
	if token == "" {
		responsehttp.WriteError(w, r, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	opts := &websocket.AcceptOptions{}
	if origin := h.allowedOriginPattern(); origin != "" {
		opts.OriginPatterns = []string{origin}
	}

	conn, err := websocket.Accept(w, r, opts)
	if err != nil {
		h.log.Error("websocket accept", "err", err)
		return
	}
	defer func() { _ = conn.Close(websocket.StatusNormalClosure, "") }()

	connCtx, cancel := context.WithCancel(r.Context())
	defer cancel()

	var subMu sync.Mutex
	var subCancel context.CancelFunc

	for {
		_, data, err := conn.Read(connCtx)
		if err != nil {
			return
		}

		var msg clientMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}
		if msg.Type != "subscribe" {
			continue
		}

		groupID := strings.ToLower(strings.TrimSpace(msg.GroupID))
		channelID := strings.ToLower(strings.TrimSpace(msg.ChannelID))
		if !publicEntityID.MatchString(groupID) || !publicEntityID.MatchString(channelID) {
			continue
		}

		_, err = h.ms.User.GRPC.Channel.AssertGroupChannelAccess(connCtx, &userv1.AssertGroupChannelAccessRequest{
			AccessToken: token,
			GroupId:     groupID,
			ChannelId:   channelID,
		})
		if err != nil {
			_ = conn.Close(websocket.StatusPolicyViolation, "forbidden")
			return
		}

		subMu.Lock()
		if subCancel != nil {
			subCancel()
		}
		fanoutCtx, fanoutStop := context.WithCancel(connCtx)
		subCancel = fanoutStop
		subMu.Unlock()

		go h.hub.Subscribe(fanoutCtx, conn, groupID, channelID)
	}
}

func (h Handler) allowedOriginPattern() string {
	raw := strings.TrimSpace(h.webOrigin)
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	if u.Host == "" {
		return ""
	}

	return u.Host
}
