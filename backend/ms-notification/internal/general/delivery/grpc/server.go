package grpc

import (
	"context"

	notificationv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/notification/v1"
)

// Server is the general gRPC delivery (operational RPCs).
type Server interface {
	notificationv1.GeneralServiceServer
}

type server struct {
	notificationv1.UnimplementedGeneralServiceServer
	serviceName string
}

// NewServer binds operational handlers for the given process name.
func NewServer(serviceName string) Server {
	return &server{serviceName: serviceName}
}

func (s *server) Health(context.Context, *notificationv1.HealthRequest) (*notificationv1.HealthResponse, error) {
	return &notificationv1.HealthResponse{
		Status:  notificationv1.ServingStatus_SERVING_STATUS_SERVING,
		Service: s.serviceName,
	}, nil
}
