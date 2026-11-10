package middleware_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wichat/wichat/backend/api-gateway/internal/middleware"
	"github.com/wichat/wichat/backend/api-gateway/pkg/constants"
)

func TestClientIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.0.2.1:12345"
	if middleware.ClientIP(req) != "192.0.2.1" {
		t.Fatalf("got %q", middleware.ClientIP(req))
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
