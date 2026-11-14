package turnstile_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wichat/wichat/backend/api-gateway/pkg/turnstile"
	"github.com/wichat/wichat/backend/wi-shared/exception"
)

func TestTurnstileDisabledBypass(t *testing.T) {
	v := turnstile.NewVerifier("")
	if v.Enabled() {
		t.Fatal("expected disabled verifier")
	}
	if err := v.Verify(context.Background(), "", "127.0.0.1"); err != nil {
		t.Fatalf("expected bypass nil error, got %v", err)
	}
}

func TestTurnstileVerifySuccessAndFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		if r.Form.Get("response") == "good-token" {
			_, _ = w.Write([]byte(`{"success":true}`))
			return
		}
		_, _ = w.Write([]byte(`{"success":false}`))
	}))
	defer srv.Close()

	v := turnstile.NewVerifierWithEndpoint("sec-123", srv.URL, srv.Client())
	if !v.Enabled() {
		t.Fatal("expected enabled verifier")
	}

	var ex *exception.Error
	if err := v.Verify(context.Background(), "", "1.2.3.4"); !errors.As(err, &ex) || ex.Kind != exception.KindPermissionDenied {
		t.Fatalf("expected permission denied for empty token, got %v", err)
	}
	if err := v.Verify(context.Background(), "bad-token", "1.2.3.4"); !errors.As(err, &ex) || ex.Kind != exception.KindPermissionDenied {
		t.Fatalf("expected permission denied for bad token, got %v", err)
	}
	if err := v.Verify(context.Background(), "good-token", "1.2.3.4"); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}
