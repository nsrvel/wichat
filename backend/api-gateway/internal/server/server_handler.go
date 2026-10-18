// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

package server

import (
	"encoding/json"
	"net/http"

	authdelivery "github.com/wichat/wichat/backend/api-gateway/internal/auth/delivery/http"
	"github.com/wichat/wichat/backend/api-gateway/internal/middleware"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
)

func (s *server) buildHandler(ms *microservices.Clients) (http.Handler, error) {
	// Routes
	mux := http.NewServeMux()
	s.mapHandlers(mux, ms)

	// Middleware
	return middleware.CORS(s.cfg.WebOrigin, middleware.Chain(mux)), nil
}

func (s *server) mapHandlers(mux *http.ServeMux, ms *microservices.Clients) {
	// APIs
	authdelivery.Register(mux, authdelivery.NewHandler(ms, s.log, s.cfg.SecureCookies(), s.cfg.WebOrigin))

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
