package grpc

import (
	"context"

	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

// Server is the general gRPC delivery (operational RPCs).
type Server interface {
	userv1.GeneralServiceServer
}

type server struct {
	userv1.UnimplementedGeneralServiceServer
	serviceName string
}

// NewServer binds operational handlers for the given process name (e.g. "ms-user").
func NewServer(serviceName string) Server {
	return &server{serviceName: serviceName}
}

func (s *server) Health(context.Context, *userv1.HealthRequest) (*userv1.HealthResponse, error) {
	return &userv1.HealthResponse{
		Status:  userv1.ServingStatus_SERVING_STATUS_SERVING,
		Service: s.serviceName,
	}, nil
}
