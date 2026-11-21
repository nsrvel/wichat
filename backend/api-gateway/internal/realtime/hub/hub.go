package hub

import (
	"context"
	"fmt"
	"sync"

	"github.com/coder/websocket"
	"github.com/wichat/wichat/backend/wi-shared/realtime"
)

// Hub tracks WebSocket subscribers per group channel and per group for realtime fanout.
type Hub struct {
	mu        sync.RWMutex
	subs      map[string]map[*websocket.Conn]struct{}
	groupSubs map[string]map[*websocket.Conn]struct{}
	writeMu   map[*websocket.Conn]*sync.Mutex
}

// New builds an in-memory subscription registry.
func New() *Hub {
	return &Hub{
		subs:      make(map[string]map[*websocket.Conn]struct{}),
		groupSubs: make(map[string]map[*websocket.Conn]struct{}),
		writeMu:   make(map[*websocket.Conn]*sync.Mutex),
	}
}

// BindConn registers per-connection write serialization (call once per accepted socket).
func (h *Hub) BindConn(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.writeMu[conn] == nil {
		h.writeMu[conn] = &sync.Mutex{}
	}
}

// UnbindConn drops the write lock entry when the socket closes.
func (h *Hub) UnbindConn(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.writeMu, conn)
	h.mu.Unlock()
}

// WriteText sends a text frame with the connection write lock held.
func (h *Hub) WriteText(ctx context.Context, conn *websocket.Conn, payload []byte) error {
	h.mu.RLock()
	mu := h.writeMu[conn]
	h.mu.RUnlock()
	if mu == nil {
		return fmt.Errorf("hub: connection not bound")
	}
	mu.Lock()
	defer mu.Unlock()

	return conn.Write(ctx, websocket.MessageText, payload)
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

func (h *Hub) deliverToBucket(bucketMap map[string]map[*websocket.Conn]struct{}, key string, payload []byte) error {
	h.mu.RLock()
	bucket := bucketMap[key]
	targets := make([]*websocket.Conn, 0, len(bucket))
	for conn := range bucket {
		targets = append(targets, conn)
	}
	h.mu.RUnlock()

	if len(targets) == 0 {
		return nil
	}

	writeCtx := context.Background()
	var firstErr error
	for _, conn := range targets {
		if err := h.WriteText(writeCtx, conn, payload); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if firstErr != nil {
		return firstErr
	}

	return nil
}

// Deliver writes payload to all subscribers on the channel (no-op when empty).
func (h *Hub) Deliver(groupID, channelID string, payload []byte) error {
	return h.deliverToBucket(h.subs, realtime.ChannelKey(groupID, channelID), payload)
}

// DeliverGroup writes payload to all subscribers on the group (no-op when empty).
func (h *Hub) DeliverGroup(groupID string, payload []byte) error {
	return h.deliverToBucket(h.groupSubs, groupID, payload)
}
