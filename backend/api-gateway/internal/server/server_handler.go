package server

import (
	"encoding/json"
	"net/http"
	"strings"

	authdelivery "github.com/wichat/wichat/backend/api-gateway/internal/auth/delivery/http"
	channelcategorydelivery "github.com/wichat/wichat/backend/api-gateway/internal/channel-category/delivery/http"
	channeldelivery "github.com/wichat/wichat/backend/api-gateway/internal/channel/delivery/http"
	chatdelivery "github.com/wichat/wichat/backend/api-gateway/internal/chat/delivery/http"
	groupmemberdelivery "github.com/wichat/wichat/backend/api-gateway/internal/group-member/delivery/http"
	grouproledelivery "github.com/wichat/wichat/backend/api-gateway/internal/group-role/delivery/http"
	groupdelivery "github.com/wichat/wichat/backend/api-gateway/internal/group/delivery/http"
	invitedelivery "github.com/wichat/wichat/backend/api-gateway/internal/invite/delivery/http"
	"github.com/wichat/wichat/backend/api-gateway/internal/middleware"
	realtimeWs "github.com/wichat/wichat/backend/api-gateway/internal/realtime/delivery/ws"
	"github.com/wichat/wichat/backend/api-gateway/internal/realtime/hub"
	userdelivery "github.com/wichat/wichat/backend/api-gateway/internal/user/delivery/http"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
)

func (s *server) buildHandler(ms *microservices.Clients, rtHub *hub.Hub, monitor *microservices.HealthMonitor) (http.Handler, error) {
	if err := middleware.ConfigureTrustedProxies(s.cfg.TrustedProxies); err != nil {
		return nil, err
	}

	// Routes
	mux := http.NewServeMux()
	s.mapHandlers(mux, ms, rtHub, monitor)

	// Middleware
	return middleware.CORS(s.cfg.WebOrigin, middleware.CrossOriginStateChange(s.cfg.WebOrigin, middleware.Chain(mux))), nil
}

// RegisterHealthRoutes registers /api/v1/healthz and /api/v1/readyz on mux.
func RegisterHealthRoutes(mux *http.ServeMux, monitor *microservices.HealthMonitor) {
	mux.HandleFunc("GET /api/v1/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /api/v1/readyz", func(w http.ResponseWriter, _ *http.Request) {
		services, ready := monitor.Snapshot()
		w.Header().Set("Content-Type", "application/json")
		statusText := "ok"
		if !ready {
			statusText = "degraded"
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":   statusText,
			"services": services,
		})
	})
}

func (s *server) mapHandlers(mux *http.ServeMux, ms *microservices.Clients, rtHub *hub.Hub, monitor *microservices.HealthMonitor) {
	// Health
	RegisterHealthRoutes(mux, monitor)

	// APIs
	dev := strings.EqualFold(strings.TrimSpace(s.cfg.Env), "development")
	authdelivery.Register(mux, authdelivery.NewHandler(ms, s.log, s.cfg.SecureCookies(), s.cfg.WebOrigin, s.cfg.TurnstileVerifySecret()).
		WithInviteOnly(s.cfg.InviteOnly), dev)
	userdelivery.Register(mux, userdelivery.NewHandler(ms, s.log, s.cfg.TurnstileClientSiteKey()).
		WithTurnstileRequired(s.cfg.TurnstileEnabled).
		WithInviteOnly(s.cfg.InviteOnly).
		WithPresenceBroadcaster(rtHub))
	groupdelivery.Register(mux, groupdelivery.NewHandler(ms, s.log))
	groupmemberdelivery.Register(mux, groupmemberdelivery.NewHandler(ms, s.log))
	grouproledelivery.Register(mux, grouproledelivery.NewHandler(ms, s.log))
	channeldelivery.Register(mux, channeldelivery.NewHandler(ms, s.log))
	channelcategorydelivery.Register(mux, channelcategorydelivery.NewHandler(ms, s.log))
	invitedelivery.Register(mux, invitedelivery.NewHandler(ms, s.log, s.cfg.WebOrigin))
	chatdelivery.Register(mux, chatdelivery.NewHandler(ms, s.log))
	realtimeWs.Register(mux, realtimeWs.NewHandler(ms, rtHub, s.log, s.cfg.WebOrigin))

	// Root
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"service": "api-gateway"})
	})
}
