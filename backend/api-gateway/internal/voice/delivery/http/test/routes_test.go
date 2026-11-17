package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	voicedelivery "github.com/wichat/wichat/backend/api-gateway/internal/voice/delivery/http"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

func TestRegisterRoutes(t *testing.T) {
	mux := http.NewServeMux()
	voicedelivery.Register(mux, voicedelivery.NewHandler(nil, logger.New(logger.Options{Service: "gw-test", Env: "test", Level: "error"})))

	cases := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/channels/ch-1/voice/token"},
		{http.MethodPost, "/api/v1/groups/g-1/channels/ch-1/voice/token"},
		{http.MethodGet, "/api/v1/groups/g-1/voice/occupancy"},
		{http.MethodPost, "/api/v1/channels/ch-1/voice/moderate"},
		{http.MethodPost, "/api/v1/channels/ch-1/voice/keepalive"},
		{http.MethodPost, "/api/v1/channels/ch-1/voice/recording/start"},
		{http.MethodPost, "/api/v1/channels/ch-1/voice/recording/stop"},
		{http.MethodGet, "/api/v1/channels/ch-1/voice/recordings"},
		{http.MethodGet, "/api/v1/voice/quota"},
		{http.MethodPost, "/api/v1/webhooks/livekit"},
	}

	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code == http.StatusNotFound || rec.Code == http.StatusMethodNotAllowed {
			t.Fatalf("%s %s returned %d", tc.method, tc.path, rec.Code)
		}
	}
}
