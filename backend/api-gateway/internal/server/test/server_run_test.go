package test

import (
	"context"
	"testing"

	"github.com/wichat/wichat/backend/api-gateway/internal/config"
	"github.com/wichat/wichat/backend/api-gateway/internal/server"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

func TestServerRunInvalidConfig(t *testing.T) {
	srv := server.NewServer(config.Config{}, logger.New(logger.Options{Service: "test", Env: "test"}))
	err := srv.Run(context.Background())
	if err == nil {
		t.Fatal("expected error on empty config, got nil")
	}
}
