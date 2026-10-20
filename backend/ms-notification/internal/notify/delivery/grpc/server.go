package grpc

import (
	"context"

	"github.com/wichat/wichat/backend/ms-notification/internal/models"
	"github.com/wichat/wichat/backend/ms-notification/internal/notify/service"
	notificationv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/notification/v1"
	responsegrpc "github.com/wichat/wichat/backend/wi-shared/response/grpc"
)

// Server is the notify gRPC delivery.
type Server interface {
	notificationv1.NotifyServiceServer
}

type server struct {
	notificationv1.UnimplementedNotifyServiceServer
	svc service.Service
}

// NewServer binds the notify service to gRPC.
func NewServer(svc service.Service) Server {
	return &server{svc: svc}
}

func (s *server) Send(ctx context.Context, req *notificationv1.SendRequest) (*notificationv1.SendResponse, error) {
	// Bind
	in := models.SendInput{
		Channel:    req.GetChannel(),
		TemplateID: req.GetTemplateId(),
		Locale:     req.GetLocale(),
		Vars:       req.GetVars(),
		EmailTo:    "",
	}
	if req.GetEmail() != nil {
		in.EmailTo = req.GetEmail().GetTo()
	}

	// Service
	if err := s.svc.Send(ctx, in); err != nil {
		return nil, responsegrpc.ToStatus(err)
	}

	// Response
	return &notificationv1.SendResponse{Ok: true}, nil
}
