package test

import (
	"testing"

	"github.com/wichat/wichat/backend/api-gateway/internal/config"
	"github.com/wichat/wichat/backend/api-gateway/internal/server"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

func TestNewServer(t *testing.T) {
	srv := server.NewServer(config.Config{}, logger.New(logger.Options{Service: "test", Env: "test"}))
	if srv == nil {
		t.Fatal("expected non-nil server")
	}
}
