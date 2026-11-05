package middleware_test

import (
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
