package correlation_test

import (
	"context"
	"testing"

	"github.com/wichat/wichat/backend/wi-shared/microservices/utils/correlation"
)

func TestRequestIDRoundTrip(t *testing.T) {
	t.Parallel()

	base := context.Background()
	if correlation.RequestIDFromContext(base) != "" {
		t.Fatal("expected empty")
	}
	ctx := correlation.WithRequestID(base, "req-1")
	if correlation.RequestIDFromContext(ctx) != "req-1" {
		t.Fatalf("got %q", correlation.RequestIDFromContext(ctx))
	}
	if correlation.WithRequestID(base, "") == base {
		return
	}
	t.Fatal("empty id should leave ctx unchanged")
}
