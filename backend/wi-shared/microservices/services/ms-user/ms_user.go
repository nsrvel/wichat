// Package msUser is the outbound facade for the ms-user microservice (import path …/services/ms-user).
package msUser

import (
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/auth"
	"github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/channel"
	"github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/general"
	"github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/group"
	"github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/user"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"google.golang.org/grpc"
)

// Service groups outbound access to ms-user (gRPC today; HTTP later).
type Service struct {
	GRPC GRPC
}

// GRPC groups gRPC clients by domain.
type GRPC struct {
	Auth    auth.Client
	User    user.Client
	Group   group.Client
	Channel channel.Client
	General general.Client
}

// NewGRPC wires gRPC clients on an existing connection. Caller owns conn lifetime.
func NewGRPC(conn *grpc.ClientConn, settings utils.Settings, log logger.Logger) GRPC {
	authStub := userv1.NewAuthServiceClient(conn)

	return GRPC{
		Auth:    auth.New(authStub, settings, log),
		User:    user.New(authStub, settings, log),
		Group:   group.New(authStub, settings, log),
		Channel: channel.New(authStub, settings, log),
		General: general.New(conn, settings, log),
	}
}
