// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

package server

import (
	"context"

	"github.com/wichat/wichat/backend/ms-user/internal/config"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// Server runs ms-user gRPC.
type Server interface {
	Run(ctx context.Context) error
}

type server struct {
	cfg config.Config
	log logger.Logger
}

// NewServer constructs the process composition root.
func NewServer(cfg config.Config, log logger.Logger) Server {
	return &server{cfg: cfg, log: log}
}
