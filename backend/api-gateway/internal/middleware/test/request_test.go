package middleware_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wichat/wichat/backend/api-gateway/internal/middleware"
	"github.com/wichat/wichat/backend/api-gateway/pkg/constants"
)

func TestClientIP(t *testing.T) {
	t.Cleanup(func() {
		_ = middleware.ConfigureTrustedProxies("")
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.0.2.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.9")
	if middleware.ClientIP(req) != "192.0.2.1" {
		t.Fatalf("untrusted proxy should ignore XFF, got %q", middleware.ClientIP(req))
	}

	if err := middleware.ConfigureTrustedProxies("127.0.0.1,10.0.0.0/8"); err != nil {
		t.Fatalf("configure trusted proxies: %v", err)
	}

	proxyReq := httptest.NewRequest("GET", "/", nil)
	proxyReq.RemoteAddr = "10.0.0.5:443"
	proxyReq.Header.Set("X-Forwarded-For", "198.51.100.2, 10.0.0.1")
	if got := middleware.ClientIP(proxyReq); got != "198.51.100.2" {
		t.Fatalf("expected untrusted XFF hop 198.51.100.2, got %q", got)
	}

	cfReq := httptest.NewRequest("GET", "/", nil)
	cfReq.RemoteAddr = "127.0.0.1:9000"
	cfReq.Header.Set("CF-Connecting-IP", "203.0.113.42")
	if got := middleware.ClientIP(cfReq); got != "203.0.113.42" {
		t.Fatalf("expected CF-Connecting-IP 203.0.113.42, got %q", got)
	}

	xriReq := httptest.NewRequest("GET", "/", nil)
	xriReq.RemoteAddr = "127.0.0.1:9000"
	xriReq.Header.Set("X-Real-IP", "203.0.113.77")
	if got := middleware.ClientIP(xriReq); got != "203.0.113.77" {
		t.Fatalf("expected X-Real-IP 203.0.113.77, got %q", got)
	}

	if err := middleware.ConfigureTrustedProxies("not-a-cidr"); err == nil {
		t.Fatal("expected invalid proxy error")
	}
}

func TestUserAgentTruncated(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("User-Agent", strings.Repeat("x", constants.MaxUserAgentLen+10))
	got := middleware.UserAgent(req)
	if len(got) != constants.MaxUserAgentLen {
		t.Fatalf("len %d", len(got))
	}
}
