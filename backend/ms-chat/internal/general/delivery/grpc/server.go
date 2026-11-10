package grpc

import (
	"context"

	chatv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/chat/v1"
)

// Server is the general gRPC delivery (operational RPCs).
type Server interface {
	chatv1.GeneralServiceServer
}

type server struct {
	chatv1.UnimplementedGeneralServiceServer
	serviceName string
}

// NewServer binds operational handlers for the given process name.
func NewServer(serviceName string) Server {
	return &server{serviceName: serviceName}
}

func (s *server) Health(context.Context, *chatv1.HealthRequest) (*chatv1.HealthResponse, error) {
	return &chatv1.HealthResponse{
		Status:  chatv1.ServingStatus_SERVING_STATUS_SERVING,
		Service: s.serviceName,
	}, nil
}
