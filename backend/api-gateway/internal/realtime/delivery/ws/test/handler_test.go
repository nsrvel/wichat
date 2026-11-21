package test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/wichat/wichat/backend/api-gateway/internal/realtime/delivery/ws"
	"github.com/wichat/wichat/backend/api-gateway/internal/realtime/hub"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type stubPresenceClient struct {
	mu           sync.Mutex
	tokens       []string
	heartbeatFn  func(in *userv1.HeartbeatPresenceRequest) (*userv1.PresenceTransitionResponse, error)
	disconnectFn func(in *userv1.DisconnectPresenceRequest) (*userv1.PresenceTransitionResponse, error)
}

func (s *stubPresenceClient) HeartbeatPresence(_ context.Context, in *userv1.HeartbeatPresenceRequest) (*userv1.PresenceTransitionResponse, error) {
	s.mu.Lock()
	s.tokens = append(s.tokens, in.GetAccessToken())
	fn := s.heartbeatFn
	s.mu.Unlock()
	if fn != nil {
		return fn(in)
	}
	return &userv1.PresenceTransitionResponse{
		UserId:          "user-1",
		PreviousStatus:  "online",
		EffectiveStatus: "online",
		IsOnline:        true,
	}, nil
}

func (s *stubPresenceClient) DisconnectPresence(_ context.Context, in *userv1.DisconnectPresenceRequest) (*userv1.PresenceTransitionResponse, error) {
	if s.disconnectFn != nil {
		return s.disconnectFn(in)
	}
	return &userv1.PresenceTransitionResponse{}, nil
}

func (s *stubPresenceClient) SetManualPresence(context.Context, *userv1.SetManualPresenceRequest) (*userv1.SetManualPresenceResponse, error) {
	return &userv1.SetManualPresenceResponse{}, nil
}

func (s *stubPresenceClient) GetMyPresence(context.Context, *userv1.GetMyPresenceRequest) (*userv1.GetMyPresenceResponse, error) {
	return &userv1.GetMyPresenceResponse{}, nil
}

type stubGroupClient struct {
	groups []*userv1.GroupSummary
}

func (s stubGroupClient) ListMyGroups(context.Context, *userv1.ListMyGroupsRequest) (*userv1.ListMyGroupsResponse, error) {
	return &userv1.ListMyGroupsResponse{Groups: s.groups}, nil
}

func (stubGroupClient) CreateGroup(context.Context, *userv1.CreateGroupRequest) (*userv1.CreateGroupResponse, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (stubGroupClient) DeleteGroup(context.Context, *userv1.DeleteGroupRequest) (*userv1.DeleteGroupResponse, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (stubGroupClient) UploadGroupIcon(context.Context, *userv1.UploadGroupIconRequest) (*userv1.UploadGroupIconResponse, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (stubGroupClient) UpdateGroupProfile(context.Context, *userv1.UpdateGroupProfileRequest) (*userv1.UpdateGroupProfileResponse, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (stubGroupClient) UploadGroupBanner(context.Context, *userv1.UploadGroupBannerRequest) (*userv1.UploadGroupBannerResponse, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (stubGroupClient) ReorderMyGroups(context.Context, *userv1.ReorderMyGroupsRequest) (*userv1.ReorderMyGroupsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (stubGroupClient) ListGroupActivityLogs(context.Context, *userv1.ListGroupActivityLogsRequest) (*userv1.ListGroupActivityLogsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func TestRealtimeHeartbeatAuthRefreshAndInitialSync(t *testing.T) {
	t.Parallel()

	stub := &stubPresenceClient{}
	stub.heartbeatFn = func(in *userv1.HeartbeatPresenceRequest) (*userv1.PresenceTransitionResponse, error) {
		if in.GetAccessToken() == "expired-token" {
			return nil, status.Error(codes.Unauthenticated, "expired")
		}
		return &userv1.PresenceTransitionResponse{
			UserId:          "user-1",
			PreviousStatus:  "online",
			EffectiveStatus: "online",
			IsOnline:        true,
		}, nil
	}

	ms := &microservices.Clients{}
	ms.User.GRPC.Presence = stub

	h := ws.NewHandler(ms, hub.New(), logger.New(logger.Options{Service: "gateway-test", Env: "test", Level: "error"}), "")
	srv := httptest.NewServer(http.HandlerFunc(h.Realtime))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := strings.Replace(srv.URL, "http://", "ws://", 1) + "?access_token=expired-token"
	conn, resp, err := websocket.Dial(ctx, wsURL, nil)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer func() { _ = conn.Close(websocket.StatusNormalClosure, "") }()

	// 1. Heartbeat with expired initial token triggers presence.auth_required.
	if err := conn.Write(ctx, websocket.MessageText, []byte(`{"type":"presence.heartbeat","activity":"active"}`)); err != nil {
		t.Fatalf("write expired heartbeat: %v", err)
	}
	_, msg1, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read auth_required: %v", err)
	}
	var authFrame struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(msg1, &authFrame); err != nil {
		t.Fatalf("unmarshal auth_required: %v", err)
	}
	if authFrame.Type != "presence.auth_required" {
		t.Fatalf("expected presence.auth_required, got %q", authFrame.Type)
	}

	// 2. Heartbeat with refreshed access_token succeeds and emits initial presence.updated even when prev == effective.
	if err := conn.Write(ctx, websocket.MessageText, []byte(`{"type":"presence.heartbeat","activity":"active","access_token":"fresh-token"}`)); err != nil {
		t.Fatalf("write fresh heartbeat: %v", err)
	}
	_, msg2, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read presence.updated: %v", err)
	}
	var updatedFrame struct {
		Type           string `json:"type"`
		UserID         string `json:"user_id"`
		PresenceStatus string `json:"presence_status"`
		IsOnline       bool   `json:"is_online"`
	}
	if err := json.Unmarshal(msg2, &updatedFrame); err != nil {
		t.Fatalf("unmarshal presence.updated: %v", err)
	}
	if updatedFrame.Type != "presence.updated" || updatedFrame.UserID != "user-1" || !updatedFrame.IsOnline || updatedFrame.PresenceStatus != "online" {
		t.Fatalf("unexpected presence.updated frame: %+v", updatedFrame)
	}

	// 3. Subsequent heartbeat without access_token reuses updated token ("fresh-token").
	if err := conn.Write(ctx, websocket.MessageText, []byte(`{"type":"presence.heartbeat","activity":"active"}`)); err != nil {
		t.Fatalf("write third heartbeat: %v", err)
	}

	// Wait briefly for server to process third heartbeat before checking recorded tokens.
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		stub.mu.Lock()
		n := len(stub.tokens)
		stub.mu.Unlock()
		if n >= 3 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	stub.mu.Lock()
	defer stub.mu.Unlock()
	if len(stub.tokens) < 3 || stub.tokens[2] != "fresh-token" {
		t.Fatalf("expected 3rd heartbeat to use fresh-token, got %v", stub.tokens)
	}
}

func TestSubscribeGroupForbiddenKeepsConnectionAlive(t *testing.T) {
	t.Parallel()

	stub := &stubPresenceClient{}
	stub.heartbeatFn = func(in *userv1.HeartbeatPresenceRequest) (*userv1.PresenceTransitionResponse, error) {
		return &userv1.PresenceTransitionResponse{
			UserId:          "user-1",
			PreviousStatus:  "offline",
			EffectiveStatus: "online",
			IsOnline:        true,
		}, nil
	}

	ms := &microservices.Clients{}
	ms.User.GRPC.Presence = stub
	ms.User.GRPC.Group = stubGroupClient{}

	h := ws.NewHandler(ms, hub.New(), logger.New(logger.Options{Service: "gateway-test", Env: "test", Level: "error"}), "")
	srv := httptest.NewServer(http.HandlerFunc(h.Realtime))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := strings.Replace(srv.URL, "http://", "ws://", 1) + "?access_token=test-token"
	conn, resp, err := websocket.Dial(ctx, wsURL, nil)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer func() { _ = conn.Close(websocket.StatusNormalClosure, "") }()

	deniedGroup := "aaaaaaaaaaaaaaaaaaaaaaaa"
	if err := conn.Write(ctx, websocket.MessageText, []byte(`{"type":"subscribe_group","group_id":"`+deniedGroup+`"}`)); err != nil {
		t.Fatalf("write subscribe_group: %v", err)
	}
	_, msg, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read subscribe.denied: %v", err)
	}
	var denied struct {
		Type    string `json:"type"`
		Scope   string `json:"scope"`
		GroupID string `json:"group_id"`
		Code    string `json:"code"`
	}
	if err := json.Unmarshal(msg, &denied); err != nil {
		t.Fatalf("unmarshal subscribe.denied: %v", err)
	}
	if denied.Type != "subscribe.denied" || denied.Code != "forbidden" || denied.GroupID != deniedGroup {
		t.Fatalf("unexpected subscribe.denied: %+v", denied)
	}

	if err := conn.Write(ctx, websocket.MessageText, []byte(`{"type":"presence.heartbeat","activity":"active"}`)); err != nil {
		t.Fatalf("write heartbeat after denied subscribe: %v", err)
	}
	_, msg2, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read presence.updated after denied subscribe: %v", err)
	}
	var updated struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(msg2, &updated); err != nil {
		t.Fatalf("unmarshal presence.updated: %v", err)
	}
	if updated.Type != "presence.updated" {
		t.Fatalf("expected presence.updated after denied subscribe, got %q", updated.Type)
	}
}
