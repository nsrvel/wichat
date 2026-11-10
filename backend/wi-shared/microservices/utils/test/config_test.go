package utils_test

import (
	"testing"
	"time"

	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
)

func TestCallConfigFromSettings(t *testing.T) {
	cfg := utils.CallConfigFromSettings(utils.DefaultSettings, utils.RetryIdempotent)
	if cfg.MaxRetries != 3 || cfg.PerAttemptTimeout != 5*time.Second {
		t.Fatalf("unexpected cfg %+v", cfg)
	}
}
