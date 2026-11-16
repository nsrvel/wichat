package test

import (
	"testing"

	"github.com/wichat/wichat/backend/api-gateway/internal/config"
)

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
