package server

import (
	"github.com/wichat/wichat/backend/ms-core/internal/config"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

type server struct {
	cfg config.Config
	log logger.Logger
}

// NewServer constructs the ms-core gRPC server.
func NewServer(cfg config.Config, log logger.Logger) *server {
	return &server{cfg: cfg, log: log}
}
