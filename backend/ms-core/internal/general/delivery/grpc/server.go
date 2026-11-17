package grpc

import (
	"context"

	corev1 "github.com/wichat/wichat/backend/wi-shared/models/gen/core/v1"
)

// Server is the general gRPC delivery (operational RPCs).
type Server interface {
	corev1.GeneralServiceServer
}

type server struct {
	corev1.UnimplementedGeneralServiceServer
	serviceName string
}

// NewServer binds operational handlers for the given process name.
func NewServer(serviceName string) Server {
	return &server{serviceName: serviceName}
}

func (s *server) Health(context.Context, *corev1.HealthRequest) (*corev1.HealthResponse, error) {
	return &corev1.HealthResponse{
		Status:  corev1.ServingStatus_SERVING_STATUS_SERVING,
		Service: s.serviceName,
	}, nil
}
