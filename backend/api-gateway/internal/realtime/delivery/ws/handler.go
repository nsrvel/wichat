package ws

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/wichat/wichat/backend/api-gateway/internal/realtime/hub"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var publicEntityID = regexp.MustCompile(`^[0-9a-f]{24}$`)

type clientMessage struct {
	Type        string `json:"type"`
	GroupID     string `json:"group_id"`
	ChannelID   string `json:"channel_id"`
	Activity    string `json:"activity"`
	AccessToken string `json:"access_token,omitempty"`
}

type presenceUpdatedEvent struct {
	Type           string `json:"type"`
	UserID         string `json:"user_id"`
	PresenceStatus string `json:"presence_status"`
	IsOnline       bool   `json:"is_online"`
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

func newConnID() string {
	var b [12]byte
	if _, err := rand.Read(b[:]); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(b[:])
}

func (h Handler) broadcastPresenceTransition(tr *userv1.PresenceTransitionResponse) {
	if tr == nil || tr.GetEffectiveStatus() == tr.GetPreviousStatus() || len(tr.GetGroupIds()) == 0 {
		return
	}
	payload, err := json.Marshal(presenceUpdatedEvent{
		Type:           "presence.updated",
		UserID:         tr.GetUserId(),
		PresenceStatus: tr.GetEffectiveStatus(),
		IsOnline:       tr.GetIsOnline(),
	})
	if err != nil {
		return
	}
	for _, gid := range tr.GetGroupIds() {
		h.hub.DeliverGroup(gid, payload)
	}
}

func (h Handler) isGroupMember(ctx context.Context, token, groupID string) bool {
	res, err := h.ms.User.GRPC.Group.ListMyGroups(ctx, &userv1.ListMyGroupsRequest{AccessToken: token})
	if err != nil {
		return false
	}
	for _, g := range res.GetGroups() {
		if strings.EqualFold(g.GetId(), groupID) {
			return true
		}
	}
	return false
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

	connID := newConnID()
	var userIDMu sync.Mutex
	var authedUserID string

	defer func() {
		userIDMu.Lock()
		uid := authedUserID
		userIDMu.Unlock()
		if uid == "" || h.ms == nil {
			return
		}
		discCtx, discCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer discCancel()
		if tr, discErr := h.ms.User.GRPC.Presence.DisconnectPresence(discCtx, &userv1.DisconnectPresenceRequest{
			UserId: uid,
			ConnId: connID,
		}); discErr == nil {
			h.broadcastPresenceTransition(tr)
		}
	}()

	var subMu sync.Mutex
	var subCancel context.CancelFunc
	var groupSubCancel context.CancelFunc

	for {
		_, data, err := conn.Read(connCtx)
		if err != nil {
			return
		}

		var msg clientMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "presence.heartbeat":
			if h.ms == nil {
				continue
			}
			hbToken := token
			if trimmed := strings.TrimSpace(msg.AccessToken); trimmed != "" {
				hbToken = trimmed
			}
			tr, hbErr := h.ms.User.GRPC.Presence.HeartbeatPresence(connCtx, &userv1.HeartbeatPresenceRequest{
				AccessToken: hbToken,
				ConnId:      connID,
				Activity:    msg.Activity,
			})
			if hbErr != nil {
				if status.Code(hbErr) == codes.Unauthenticated {
					if authPayload, mErr := json.Marshal(struct {
						Type string `json:"type"`
					}{Type: "presence.auth_required"}); mErr == nil {
						_ = conn.Write(connCtx, websocket.MessageText, authPayload)
					}
				}
				continue
			}
			token = hbToken
			userIDMu.Lock()
			isFirstHeartbeat := authedUserID == ""
			authedUserID = tr.GetUserId()
			userIDMu.Unlock()
			if isFirstHeartbeat || tr.GetEffectiveStatus() != tr.GetPreviousStatus() {
				if selfPayload, mErr := json.Marshal(presenceUpdatedEvent{
					Type:           "presence.updated",
					UserID:         tr.GetUserId(),
					PresenceStatus: tr.GetEffectiveStatus(),
					IsOnline:       tr.GetIsOnline(),
				}); mErr == nil {
					_ = conn.Write(connCtx, websocket.MessageText, selfPayload)
				}
			}
			h.broadcastPresenceTransition(tr)

		case "subscribe_group":
			groupID := strings.ToLower(strings.TrimSpace(msg.GroupID))
			if !publicEntityID.MatchString(groupID) {
				continue
			}
			if !h.isGroupMember(connCtx, token, groupID) {
				_ = conn.Close(websocket.StatusPolicyViolation, "forbidden")
				return
			}
			subMu.Lock()
			if groupSubCancel != nil {
				groupSubCancel()
			}
			gCtx, gStop := context.WithCancel(connCtx)
			groupSubCancel = gStop
			subMu.Unlock()
			go h.hub.SubscribeGroup(gCtx, conn, groupID)

		case "subscribe":
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
