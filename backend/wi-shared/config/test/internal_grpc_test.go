package config_test

import (
	"testing"

	sharedconfig "github.com/wichat/wichat/backend/wi-shared/config"
)

func TestRequireInternalGrpcToken(t *testing.T) {
	t.Parallel()
	if err := sharedconfig.RequireInternalGrpcToken("development", ""); err != nil {
		t.Fatalf("dev: %v", err)
	}
	if err := sharedconfig.RequireInternalGrpcToken("production", "secret"); err != nil {
		t.Fatalf("prod ok: %v", err)
	}
	if err := sharedconfig.RequireInternalGrpcToken("production", ""); err == nil {
		t.Fatal("expected error when token missing in prod")
	}
}
