// Package msChat is the outbound facade for the ms-chat microservice.
package msChat

import (
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-chat/grpc/general"
	"github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-chat/grpc/message"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	"google.golang.org/grpc"
)

// Service groups outbound access to ms-chat.
type Service struct {
	GRPC GRPC
}

// GRPC groups gRPC clients by domain.
type GRPC struct {
	Message message.Client
	General general.Client
}

// NewGRPC wires gRPC clients on an existing connection.
func NewGRPC(conn *grpc.ClientConn, settings utils.Settings, log logger.Logger) GRPC {
	return GRPC{
		Message: message.New(conn, settings, log),
		General: general.New(conn, settings, log),
	}
}
