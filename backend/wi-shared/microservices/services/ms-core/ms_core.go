// Package msCore is the outbound facade for the ms-core microservice.
package msCore

import (
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-core/grpc/general"
	"github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-core/grpc/voice"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	"google.golang.org/grpc"
)

// Service groups outbound access to ms-core.
type Service struct {
	GRPC GRPC
}

// GRPC groups gRPC clients by domain.
type GRPC struct {
	Voice   voice.Client
	General general.Client
}

// NewGRPC wires gRPC clients on an existing connection.
func NewGRPC(conn *grpc.ClientConn, settings utils.Settings, log logger.Logger) GRPC {
	return GRPC{
		Voice:   voice.New(conn, settings, log),
		General: general.New(conn, settings, log),
	}
}
