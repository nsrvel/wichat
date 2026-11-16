package test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/wichat/wichat/backend/api-gateway/internal/server"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
)

func TestHealthAndReadyEndpoints(t *testing.T) {
	t.Parallel()

	var fail atomic.Bool
	fail.Store(true)

	mon := microservices.NewHealthMonitor(
		logger.New(logger.Options{Service: "test", Env: "test", Level: "error"}),
		microservices.HealthTarget{
			Name: "ms-user",
			Check: func(context.Context) error {
				if fail.Load() {
					return errors.New("down")
				}
				return nil
			},
		},
	)
	mon.ProbeOnce(context.Background())

	mux := http.NewServeMux()
	server.RegisterHealthRoutes(mux, mon)

	// /api/v1/healthz is always 200 OK while gateway is alive.
	hzReq := httptest.NewRequest(http.MethodGet, "/api/v1/healthz", nil)
	hzRec := httptest.NewRecorder()
	mux.ServeHTTP(hzRec, hzReq)
	if hzRec.Code != http.StatusOK {
		t.Fatalf("healthz status = %d, want 200", hzRec.Code)
	}

	// /api/v1/readyz returns 503 when ms-user is down.
	rzReq := httptest.NewRequest(http.MethodGet, "/api/v1/readyz", nil)
	rzRec := httptest.NewRecorder()
	mux.ServeHTTP(rzRec, rzReq)
	if rzRec.Code != http.StatusServiceUnavailable {
		t.Fatalf("readyz degraded status = %d, want 503", rzRec.Code)
	}

	// Recover ms-user and re-probe -> /api/v1/readyz returns 200.
	fail.Store(false)
	mon.ProbeOnce(context.Background())

	rzRec2 := httptest.NewRecorder()
	mux.ServeHTTP(rzRec2, httptest.NewRequest(http.MethodGet, "/api/v1/readyz", nil))
	if rzRec2.Code != http.StatusOK {
		t.Fatalf("readyz recovered status = %d, want 200", rzRec2.Code)
	}
	var payload struct {
		Status   string          `json:"status"`
		Services map[string]bool `json:"services"`
	}
	if err := json.Unmarshal(rzRec2.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal readyz: %v", err)
	}
	if payload.Status != "ok" || !payload.Services["ms-user"] {
		t.Fatalf("unexpected readyz payload: %+v", payload)
	}
}
