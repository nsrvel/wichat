// Package msUser is the outbound facade for the ms-user microservice (import path …/services/ms-user).
package msUser

import (
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/auth"
	"github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/channel"
	"github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/general"
	"github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/group"
	groupmember "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/group-member"
	grouprole "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/group-role"
	"github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/invite"
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
	Auth        auth.Client
	User        user.Client
	Group       group.Client
	GroupMember groupmember.Client
	GroupRole   grouprole.Client
	Invite      invite.Client
	Channel     channel.Client
	General     general.Client
}

// NewGRPC wires gRPC clients on an existing connection. Caller owns conn lifetime.
func NewGRPC(conn *grpc.ClientConn, settings utils.Settings, log logger.Logger) GRPC {
	return GRPC{
		Auth:        auth.New(userv1.NewAuthServiceClient(conn), settings, log),
		User:        user.New(userv1.NewUserServiceClient(conn), settings, log),
		Group:       group.New(userv1.NewGroupServiceClient(conn), settings, log),
		GroupMember: groupmember.New(userv1.NewGroupMemberServiceClient(conn), settings, log),
		GroupRole:   grouprole.New(userv1.NewGroupRoleServiceClient(conn), settings, log),
		Invite:      invite.New(userv1.NewInviteServiceClient(conn), settings, log),
		Channel:     channel.New(userv1.NewChannelServiceClient(conn), settings, log),
		General:     general.New(conn, settings, log),
	}
}
