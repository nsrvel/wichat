package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wichat/wichat/backend/api-gateway/internal/middleware"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils/correlation"
)

func TestChainSetsRequestID(t *testing.T) {
	var got string
	h := middleware.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = correlation.RequestIDFromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got == "" {
		t.Fatal("expected request id in context")
	}
	if rec.Header().Get(correlation.RequestIDHeader) != got {
		t.Fatalf("header %q ctx %q", rec.Header().Get(correlation.RequestIDHeader), got)
	}
}
