package test

import (
	"strings"
	"testing"

	"github.com/wichat/wichat/backend/ms-user/internal/config"
)

func TestConfigValidateProduction(t *testing.T) {
	t.Parallel()

	dev := config.Defaults
	if err := dev.Validate(); err != nil {
		t.Fatalf("dev defaults should validate: %v", err)
	}

	jwtSecret := strings.Repeat("a", 32)
	mfaSecret := strings.Repeat("b", 32)

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
				c.AuthJWTSecret = jwtSecret
				c.MFAEncryptionKey = mfaSecret
				c.InternalGrpcToken = "prod-grpc-secret"
			},
		},
		{
			name: "http web origin rejected in prod",
			mutate: func(c *config.Config) {
				c.Env = "production"
				c.WebOrigin = "http://chat.example.com"
				c.AuthJWTSecret = jwtSecret
				c.MFAEncryptionKey = mfaSecret
			},
			wantErr: true,
		},
		{
			name: "missing mfa key rejected in prod",
			mutate: func(c *config.Config) {
				c.Env = "production"
				c.WebOrigin = "https://chat.example.com"
				c.AuthJWTSecret = jwtSecret
				c.MFAEncryptionKey = ""
				c.MFAEnabled = true
			},
			wantErr: true,
		},
		{
			name: "prod with mfa disabled skips mfa key",
			mutate: func(c *config.Config) {
				c.Env = "production"
				c.WebOrigin = "https://chat.example.com"
				c.AuthJWTSecret = jwtSecret
				c.MFAEnabled = false
				c.InternalGrpcToken = "prod-grpc-secret"
			},
		},
		{
			name: "mfa key equal to jwt secret rejected in prod",
			mutate: func(c *config.Config) {
				c.Env = "production"
				c.WebOrigin = "https://chat.example.com"
				c.AuthJWTSecret = jwtSecret
				c.MFAEncryptionKey = jwtSecret
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
