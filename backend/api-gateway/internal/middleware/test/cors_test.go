package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wichat/wichat/backend/api-gateway/internal/middleware"
)

func TestCORSOptionsIncludesDELETE(t *testing.T) {
	var nextCalled bool
	h := middleware.CORS("http://localhost:8080", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalled = true
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/groups/x", nil)
	req.Header.Set("Origin", "http://localhost:8080")
	req.Header.Set("Access-Control-Request-Method", "DELETE")

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if nextCalled {
		t.Fatal("OPTIONS preflight must not reach the handler")
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status: got %d", rec.Code)
	}
	methods := rec.Header().Get("Access-Control-Allow-Methods")
	for _, want := range []string{"DELETE", "PATCH", "PUT"} {
		if !strings.Contains(methods, want) {
			t.Fatalf("Allow-Methods %q must include %s", methods, want)
		}
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:8080" {
		t.Fatalf("origin header: %q", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}
