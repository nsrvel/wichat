package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/wichat/wichat/backend/api-gateway/internal/middleware"
)

func TestLimiterBlocksAfterLimit(t *testing.T) {
	limiter := middleware.NewLimiter(2, time.Hour)
	if !limiter.Allow("1.2.3.4") {
		t.Fatal("expected the first hit")
	}
	if !limiter.Allow("1.2.3.4") {
		t.Fatal("expected the second hit")
	}
	if limiter.Allow("1.2.3.4") {
		t.Fatal("expected the third hit to be blocked")
	}
	if !limiter.Allow("5.6.7.8") {
		t.Fatal("expected a different key to be allowed")
	}
}

func TestLimiterWrapSetsRetryAfter(t *testing.T) {
	limiter := middleware.NewLimiter(1, 30*time.Second)
	h := limiter.Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r1 := httptest.NewRequest(http.MethodPost, "/", nil)
	r1.RemoteAddr = "198.51.100.10:1234"
	w1 := httptest.NewRecorder()
	h.ServeHTTP(w1, r1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w1.Code)
	}

	r2 := httptest.NewRequest(http.MethodPost, "/", nil)
	r2.RemoteAddr = "198.51.100.10:1234"
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, r2)
	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w2.Code)
	}
	if w2.Header().Get("Retry-After") != "30" {
		t.Fatalf("expected Retry-After=30, got %q", w2.Header().Get("Retry-After"))
	}
}
