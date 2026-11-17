package config_test

import (
	"testing"

	"github.com/wichat/wichat/backend/ms-core/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Port != "3003" {
		t.Fatalf("Port = %q, want 3003", cfg.Port)
	}
	if cfg.ListenAddr() != ":3003" {
		t.Fatalf("ListenAddr() = %q, want :3003", cfg.ListenAddr())
	}
	if cfg.LiveKitMonthlyQuotaMinutes != 5000 {
		t.Fatalf("LiveKitMonthlyQuotaMinutes = %d, want 5000", cfg.LiveKitMonthlyQuotaMinutes)
	}
}
