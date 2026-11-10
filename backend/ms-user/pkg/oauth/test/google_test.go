package google_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	googleoauth "github.com/wichat/wichat/backend/ms-user/pkg/oauth"
)

func TestPrepareConfiguredPKCEAndAuthorizeURL(t *testing.T) {
	t.Parallel()

	cfg := googleoauth.Prepare(googleoauth.Config{
		ClientID:     "client-id",
		ClientSecret: "secret",
		RedirectURL:  "http://localhost/callback",
	})
	if !googleoauth.Configured(cfg) {
		t.Fatal("expected configured")
	}

	state, err := googleoauth.RandomToken()
	if err != nil || state == "" {
		t.Fatalf("state %q %v", state, err)
	}
	verifier, challenge, err := googleoauth.PKCE()
	if err != nil || verifier == "" || challenge == "" {
		t.Fatalf("pkce %v", err)
	}

	url := googleoauth.AuthorizeURL(cfg, state, challenge)
	if !strings.Contains(url, "client_id=client-id") || !strings.Contains(url, "code_challenge=") {
		t.Fatalf("url %q", url)
	}
}

func TestDisplayName(t *testing.T) {
	t.Parallel()

	if got := googleoauth.DisplayName("Ada", "a@x.com", 30); got != "Ada" {
		t.Fatalf("%q", got)
	}
	if got := googleoauth.DisplayName("", "ada.lovelace@example.com", 30); got != "ada.lovelace" {
		t.Fatalf("%q", got)
	}
	if got := googleoauth.DisplayName("", "@", 30); got != "WiChat" {
		t.Fatalf("%q", got)
	}
	long := strings.Repeat("山", 40)
	if utf8Len(googleoauth.DisplayName(long, "", 5)) != 5 {
		t.Fatal("truncate")
	}
}

func utf8Len(s string) int {
	return len([]rune(s))
}

func TestExchangeRoundTrip(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/token" {
			_, _ = w.Write([]byte(`{"access_token":"tok"}`))
			return
		}
		_, _ = w.Write([]byte(`{"sub":"sub-1","email":"a@x.com","name":"Ada","email_verified":"true"}`))
	}))
	t.Cleanup(srv.Close)

	cfg := googleoauth.Prepare(googleoauth.Config{
		ClientID:     "id",
		ClientSecret: "sec",
		RedirectURL:  "http://localhost/cb",
		TokenURL:     srv.URL + "/token",
		UserInfoURL:  srv.URL + "/info",
	})

	profile, err := googleoauth.Exchange(context.Background(), cfg, "code", "verifier")
	if err != nil || profile.Subject != "sub-1" || !profile.Verified || profile.Email != "a@x.com" {
		t.Fatalf("%+v %v", profile, err)
	}
}

func TestExchangeDeniedOnBadToken(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "bad", http.StatusBadRequest)
	}))
	t.Cleanup(srv.Close)

	cfg := googleoauth.Prepare(googleoauth.Config{
		ClientID: "id", ClientSecret: "sec", RedirectURL: "http://cb",
		TokenURL: srv.URL,
	})
	_, err := googleoauth.Exchange(context.Background(), cfg, "code", "verifier")
	if !errors.Is(err, googleoauth.ErrDenied) {
		t.Fatalf("%v", err)
	}
}

func TestExchangeUserInfoDenied(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			_, _ = w.Write([]byte(`{"access_token":"tok"}`))
			return
		}
		http.Error(w, "nope", http.StatusUnauthorized)
	}))
	t.Cleanup(srv.Close)

	cfg := googleoauth.Prepare(googleoauth.Config{
		ClientID: "id", ClientSecret: "sec", RedirectURL: "http://cb",
		TokenURL: srv.URL + "/token", UserInfoURL: srv.URL + "/info",
	})
	_, err := googleoauth.Exchange(context.Background(), cfg, "code", "verifier")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConfiguredFalseForEmpty(t *testing.T) {
	t.Parallel()

	if googleoauth.Configured(googleoauth.Prepare(googleoauth.Config{})) {
		t.Fatal("expected not configured")
	}
}

func TestExchangeRejectsUnverifiedEmailFlag(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/token" {
			_, _ = w.Write([]byte(`{"access_token":"tok"}`))
			return
		}
		_, _ = w.Write([]byte(`{"sub":"s1","email":"a@x.com","email_verified":false}`))
	}))
	t.Cleanup(srv.Close)

	cfg := googleoauth.Prepare(googleoauth.Config{
		ClientID: "id", ClientSecret: "s", RedirectURL: "http://cb",
		TokenURL: srv.URL + "/token", UserInfoURL: srv.URL + "/info",
	})
	prof, err := googleoauth.Exchange(context.Background(), cfg, "code", "verifier")
	if err != nil || prof.Verified {
		t.Fatalf("verified=%v err=%v", prof.Verified, err)
	}
}

func TestEmailVerifiedStringTrue(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/token" {
			_, _ = w.Write([]byte(`{"access_token":"tok"}`))
			return
		}
		_, _ = w.Write([]byte(`{"sub":"s1","email":"a@x.com","email_verified":"true"}`))
	}))
	t.Cleanup(srv.Close)

	cfg := googleoauth.Prepare(googleoauth.Config{
		ClientID: "id", ClientSecret: "s", RedirectURL: "http://cb",
		TokenURL: srv.URL + "/token", UserInfoURL: srv.URL + "/info",
	})
	prof, err := googleoauth.Exchange(context.Background(), cfg, "code", "verifier")
	if err != nil || !prof.Verified {
		t.Fatalf("%+v %v", prof, err)
	}
}
