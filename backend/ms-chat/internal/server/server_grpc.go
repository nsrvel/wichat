package server

import (
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	msgrpc "github.com/wichat/wichat/backend/wi-shared/microservices/utils/grpc"
	"google.golang.org/grpc"
)

func newGRPCServer(log logger.Logger, internalToken string) *grpc.Server {
	return msgrpc.NewServer(log, internalToken, grpc.MaxRecvMsgSize(32<<20))
}
