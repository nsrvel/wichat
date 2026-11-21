package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wichat/wichat/backend/api-gateway/internal/middleware"
)

func TestCrossOriginStateChange(t *testing.T) {
	t.Parallel()
	origin := "http://localhost:8080"
	h := middleware.CrossOriginStateChange(origin, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.Header.Set("Origin", "https://evil.example")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("got %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req2.Header.Set("Origin", origin)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusNoContent {
		t.Fatalf("allowed: %d", rec2.Code)
	}
}
