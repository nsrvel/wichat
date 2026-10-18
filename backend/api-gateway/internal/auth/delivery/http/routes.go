// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

package http

import (
	"net/http"
	"time"

	"github.com/wichat/wichat/backend/api-gateway/internal/middleware"
)

// Register mounts auth routes on mux.
func Register(mux *http.ServeMux, h Handler) {
	setupLimit := middleware.NewLimiter(5, time.Hour)
	loginLimit := middleware.NewLimiter(10, time.Minute)
	refreshLimit := middleware.NewLimiter(30, time.Minute)
	googleLimit := middleware.NewLimiter(20, time.Minute)

	mux.HandleFunc("GET /api/v1/auth/setup-status", h.SetupStatus)
	mux.Handle("POST /api/v1/auth/setup", setupLimit.Wrap(http.HandlerFunc(h.Setup)))
	mux.Handle("POST /api/v1/auth/login", loginLimit.Wrap(http.HandlerFunc(h.Login)))
	googleStart := googleLimit.Wrap(http.HandlerFunc(h.GoogleStart))
	googleCallback := googleLimit.Wrap(http.HandlerFunc(h.GoogleCallback))
	mux.Handle("GET /api/v1/auth/google", googleStart)
	mux.Handle("GET /api/v1/auth/google/callback", googleCallback)
	mux.Handle("GET /api/auth/callback/google", googleCallback)
	mux.Handle("POST /api/v1/auth/refresh", refreshLimit.Wrap(http.HandlerFunc(h.Refresh)))
	mux.HandleFunc("POST /api/v1/auth/logout", h.Logout)
	mux.HandleFunc("GET /api/v1/auth/me", h.Me)
}
