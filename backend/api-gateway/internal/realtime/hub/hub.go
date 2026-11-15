package hub

import (
	"context"
	"sync"

	"github.com/coder/websocket"
	"github.com/wichat/wichat/backend/wi-shared/realtime"
)

// Hub tracks WebSocket subscribers per group channel and per group for realtime fanout.
type Hub struct {
	mu        sync.RWMutex
	subs      map[string]map[*websocket.Conn]struct{}
	groupSubs map[string]map[*websocket.Conn]struct{}
}

// New builds an in-memory subscription registry.
func New() *Hub {
	return &Hub{
		subs:      make(map[string]map[*websocket.Conn]struct{}),
		groupSubs: make(map[string]map[*websocket.Conn]struct{}),
	}
}

// Subscribe registers conn for groupID+channelID until ctx is canceled.
func (h *Hub) Subscribe(ctx context.Context, conn *websocket.Conn, groupID, channelID string) {
	key := realtime.ChannelKey(groupID, channelID)

	h.mu.Lock()
	if h.subs[key] == nil {
		h.subs[key] = make(map[*websocket.Conn]struct{})
	}
	h.subs[key][conn] = struct{}{}
	h.mu.Unlock()

	<-ctx.Done()

	h.mu.Lock()
	if bucket := h.subs[key]; bucket != nil {
		delete(bucket, conn)
		if len(bucket) == 0 {
			delete(h.subs, key)
		}
	}
	h.mu.Unlock()
}

// SubscribeGroup registers conn for group-level events (e.g. presence) until ctx is canceled.
func (h *Hub) SubscribeGroup(ctx context.Context, conn *websocket.Conn, groupID string) {
	h.mu.Lock()
	if h.groupSubs[groupID] == nil {
		h.groupSubs[groupID] = make(map[*websocket.Conn]struct{})
	}
	h.groupSubs[groupID][conn] = struct{}{}
	h.mu.Unlock()

	<-ctx.Done()

	h.mu.Lock()
	if bucket := h.groupSubs[groupID]; bucket != nil {
		delete(bucket, conn)
		if len(bucket) == 0 {
			delete(h.groupSubs, groupID)
		}
	}
	h.mu.Unlock()
}

func (h *Hub) deliverToBucket(bucketMap map[string]map[*websocket.Conn]struct{}, key string, payload []byte) {
	h.mu.RLock()
	bucket := bucketMap[key]
	targets := make([]*websocket.Conn, 0, len(bucket))
	for conn := range bucket {
		targets = append(targets, conn)
	}
	h.mu.RUnlock()

	if len(targets) == 0 {
		return
	}

	writeCtx := context.Background()
	for _, conn := range targets {
		_ = conn.Write(writeCtx, websocket.MessageText, payload)
	}
}

// Deliver writes payload to all subscribers on the channel (no-op when empty).
func (h *Hub) Deliver(groupID, channelID string, payload []byte) {
	h.deliverToBucket(h.subs, realtime.ChannelKey(groupID, channelID), payload)
}

// DeliverGroup writes payload to all subscribers on the group (no-op when empty).
func (h *Hub) DeliverGroup(groupID string, payload []byte) {
	h.deliverToBucket(h.groupSubs, groupID, payload)
}
