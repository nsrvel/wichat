package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	wsdelivery "github.com/wichat/wichat/backend/api-gateway/internal/realtime/delivery/ws"
)

func TestBearerTokenFromRequestSubprotocol(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/realtime", nil)
	req.Header.Set("Sec-WebSocket-Protocol", "json, bearer.ws-token-123")
	got := wsdelivery.BearerTokenFromRequest(req)
	if got != "ws-token-123" {
		t.Fatalf("got %q", got)
	}
}

func TestBearerTokenFromRequestQueryFallback(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/realtime?access_token=legacy", nil)
	if wsdelivery.BearerTokenFromRequest(req) != "legacy" {
		t.Fatal("expected query token")
	}
}

func TestBearerTokenFromRequestAuthorization(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/realtime", nil)
	req.Header.Set("Authorization", "Bearer hdr")
	if wsdelivery.BearerTokenFromRequest(req) != "hdr" {
		t.Fatal("expected header token")
	}
}
