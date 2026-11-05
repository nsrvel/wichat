// Package msNotification is the outbound facade for the ms-notification microservice.
package msNotification

import (
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-notification/grpc/general"
	"github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-notification/grpc/notify"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	"google.golang.org/grpc"
)

// Service groups outbound access to ms-notification.
type Service struct {
	GRPC GRPC
}

// GRPC groups gRPC clients by domain.
type GRPC struct {
	Notify  notify.Client
	General general.Client
}

// NewGRPC wires gRPC clients on an existing connection.
func NewGRPC(conn *grpc.ClientConn, settings utils.Settings, log logger.Logger) GRPC {
	return GRPC{
		Notify:  notify.New(conn, settings, log),
		General: general.New(conn, settings, log),
	}
}
