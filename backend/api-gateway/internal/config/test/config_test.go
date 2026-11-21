package test

import (
	"testing"

	"github.com/wichat/wichat/backend/api-gateway/internal/config"
)

func TestTurnstileVerifySecretToggle(t *testing.T) {
	t.Parallel()

	cfg := config.Defaults
	cfg.TurnstileSecretKey = "secret-key"
	cfg.TurnstileSiteKey = "site-key"
	if got := cfg.TurnstileVerifySecret(); got != "secret-key" {
		t.Fatalf("enabled: got %q", got)
	}
	if got := cfg.TurnstileClientSiteKey(); got != "site-key" {
		t.Fatalf("enabled site key: got %q", got)
	}

	cfg.TurnstileEnabled = false
	if got := cfg.TurnstileVerifySecret(); got != "" {
		t.Fatalf("disabled verify secret: got %q", got)
	}
	cfg.TurnstileSiteKey = "site-key"
	if got := cfg.TurnstileClientSiteKey(); got != "" {
		t.Fatalf("disabled client site key: got %q", got)
	}
}

func TestConfigValidateProduction(t *testing.T) {
	t.Parallel()

	dev := config.Defaults
	if err := dev.Validate(); err != nil {
		t.Fatalf("dev defaults should validate: %v", err)
	}

	cases := []struct {
		name    string
		mutate  func(*config.Config)
		wantErr bool
	}{
		{
			name: "valid production",
			mutate: func(c *config.Config) {
				c.Env = "production"
				c.WebOrigin = "https://chat.example.com"
				c.TurnstileSecretKey = "0x4AAAAAAA_real_secret_key"
				c.InternalGrpcToken = "prod-grpc-secret"
			},
		},
		{
			name: "http web origin rejected in prod",
			mutate: func(c *config.Config) {
				c.Env = "production"
				c.WebOrigin = "http://chat.example.com"
				c.TurnstileSecretKey = "0x4AAAAAAA_real_secret_key"
			},
			wantErr: true,
		},
		{
			name: "empty turnstile secret rejected in prod",
			mutate: func(c *config.Config) {
				c.Env = "production"
				c.WebOrigin = "https://chat.example.com"
				c.TurnstileSecretKey = ""
			},
			wantErr: true,
		},
		{
			name: "dummy turnstile test secret rejected in prod",
			mutate: func(c *config.Config) {
				c.Env = "production"
				c.WebOrigin = "https://chat.example.com"
				c.TurnstileSecretKey = "1x0000000000000000000000000000000AA"
			},
			wantErr: true,
		},
		{
			name: "turnstile disabled rejected in prod",
			mutate: func(c *config.Config) {
				c.Env = "production"
				c.WebOrigin = "https://chat.example.com"
				c.TurnstileEnabled = false
				c.TurnstileSecretKey = "0x4AAAAAAA_real_secret_key"
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := config.Defaults
			tc.mutate(&cfg)
			err := cfg.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
