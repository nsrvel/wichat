package server

import (
	"context"

	"github.com/wichat/wichat/backend/api-gateway/internal/config"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// Server runs the public HTTP API.
type Server interface {
	Run(ctx context.Context) error
}

type server struct {
	cfg config.Config
	log logger.Logger
}

// NewServer constructs the api-gateway composition root.
func NewServer(cfg config.Config, log logger.Logger) Server {
	return &server{cfg: cfg, log: log}
}
