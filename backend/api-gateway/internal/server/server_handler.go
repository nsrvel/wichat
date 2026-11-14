package server

import (
	"encoding/json"
	"net/http"

	authdelivery "github.com/wichat/wichat/backend/api-gateway/internal/auth/delivery/http"
	channeldelivery "github.com/wichat/wichat/backend/api-gateway/internal/channel/delivery/http"
	chatdelivery "github.com/wichat/wichat/backend/api-gateway/internal/chat/delivery/http"
	groupmemberdelivery "github.com/wichat/wichat/backend/api-gateway/internal/group-member/delivery/http"
	groupdelivery "github.com/wichat/wichat/backend/api-gateway/internal/group/delivery/http"
	invitedelivery "github.com/wichat/wichat/backend/api-gateway/internal/invite/delivery/http"
	"github.com/wichat/wichat/backend/api-gateway/internal/middleware"
	realtimeWs "github.com/wichat/wichat/backend/api-gateway/internal/realtime/delivery/ws"
	"github.com/wichat/wichat/backend/api-gateway/internal/realtime/hub"
	userdelivery "github.com/wichat/wichat/backend/api-gateway/internal/user/delivery/http"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
)

func (s *server) buildHandler(ms *microservices.Clients, rtHub *hub.Hub) (http.Handler, error) {
	if err := middleware.ConfigureTrustedProxies(s.cfg.TrustedProxies); err != nil {
		return nil, err
	}

	// Routes
	mux := http.NewServeMux()
	s.mapHandlers(mux, ms, rtHub)

	// Middleware
	return middleware.CORS(s.cfg.WebOrigin, middleware.Chain(mux)), nil
}

func (s *server) mapHandlers(mux *http.ServeMux, ms *microservices.Clients, rtHub *hub.Hub) {
	// APIs
	authdelivery.Register(mux, authdelivery.NewHandler(ms, s.log, s.cfg.SecureCookies(), s.cfg.WebOrigin, s.cfg.TurnstileSecretKey))
	userdelivery.Register(mux, userdelivery.NewHandler(ms, s.log, s.cfg.TurnstileSiteKey))
	groupdelivery.Register(mux, groupdelivery.NewHandler(ms, s.log))
	groupmemberdelivery.Register(mux, groupmemberdelivery.NewHandler(ms, s.log))
	channeldelivery.Register(mux, channeldelivery.NewHandler(ms, s.log))
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
