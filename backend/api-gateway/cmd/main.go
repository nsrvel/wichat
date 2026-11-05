package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/wichat/wichat/backend/api-gateway/internal/config"
	"github.com/wichat/wichat/backend/api-gateway/internal/server"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

func main() {
	// Config
	cfg, err := config.Load(".env")
	if err != nil {
		log.Printf("config: %v", err)
		os.Exit(1)
	}

	// Logger
	appLog := logger.New(logger.Options{
		Service: "api-gateway",
		Env:     cfg.Env,
		Level:   cfg.LogLevel,
	})

	// Signals
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	appLog.Info("api-gateway starting", "port", cfg.Port)

	// Server
	srv := server.NewServer(cfg, appLog)
	if err := srv.Run(ctx); err != nil {
		appLog.Error("api-gateway stopped", "err", err)
		os.Exit(1)
	}

	appLog.Info("api-gateway shutdown complete")
}
