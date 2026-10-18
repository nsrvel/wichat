// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

package server

import (
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	msgrpc "github.com/wichat/wichat/backend/wi-shared/microservices/utils/grpc"
	"google.golang.org/grpc"
)

func newGRPCServer(log logger.Logger) *grpc.Server {
	return msgrpc.NewServer(log)
}
