package http

import (
	"net/http"
	"time"

	"github.com/wichat/wichat/backend/api-gateway/internal/middleware"
)

// Register mounts user & instance-setup routes on mux.
func Register(mux *http.ServeMux, h Handler) {
	setupLimit := middleware.NewLimiter(5, time.Hour)
	checkLimit := middleware.NewLimiter(60, time.Minute)
	searchLimit := middleware.NewLimiter(120, time.Hour)

	mux.HandleFunc("GET /api/v1/auth/setup-status", h.SetupStatus)
	mux.Handle("POST /api/v1/auth/instance-setup", setupLimit.Wrap(http.HandlerFunc(h.InstanceSetup)))
	mux.Handle("GET /api/v1/users/check-username", checkLimit.Wrap(http.HandlerFunc(h.CheckUsername)))
	mux.Handle("GET /api/v1/users/search", searchLimit.Wrap(http.HandlerFunc(h.SearchUsers)))
}
