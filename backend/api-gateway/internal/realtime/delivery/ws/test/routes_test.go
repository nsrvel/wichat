package test

import (
	"context"
	"net/http"
	"testing"

	"github.com/wichat/wichat/backend/api-gateway/internal/realtime/delivery/ws"
)

func TestRegister(t *testing.T) {
	mux := http.NewServeMux()
	h := ws.NewHandler(nil, nil, nil, "")
	ws.Register(mux, h)

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/realtime", nil)
	_, pattern := mux.Handler(req)
	if pattern == "" {
		t.Fatal("GET /api/v1/realtime route not registered")
	}
}
