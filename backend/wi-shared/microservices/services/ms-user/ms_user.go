// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

// Package msUser is the outbound facade for the ms-user microservice (import path …/services/ms-user).
package msUser

import (
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/auth"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	"google.golang.org/grpc"
)

// Service groups outbound access to ms-user (gRPC today; HTTP later).
type Service struct {
	GRPC GRPC
}

// GRPC groups gRPC clients by domain.
type GRPC struct {
	Auth auth.Client
}

// NewGRPC wires gRPC clients on an existing connection. Caller owns conn lifetime.
func NewGRPC(conn *grpc.ClientConn, settings utils.Settings, log logger.Logger) GRPC {
	return GRPC{Auth: auth.New(conn, settings, log)}
}
