package server

import (
	"context"
	"fmt"
	"net"
	"time"

	generalgrpc "github.com/wichat/wichat/backend/ms-notification/internal/general/delivery/grpc"
	notifygrpc "github.com/wichat/wichat/backend/ms-notification/internal/notify/delivery/grpc"
	"github.com/wichat/wichat/backend/ms-notification/internal/notify/service"
	"github.com/wichat/wichat/backend/ms-notification/pkg/mail"
	notificationv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/notification/v1"
)

const serveShutdownTimeout = 5 * time.Second

func (s *server) Run(ctx context.Context) error {
	// Mail
	mailSender, err := mail.NewSender(s.cfg)
	if err != nil {
		return fmt.Errorf("mail: %w", err)
	}

	// Notify
	notifySvc := service.NewService(mailSender, s.log)

	// gRPC
	addr := s.cfg.ListenAddr()
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	grpcSrv := newGRPCServer(s.log)
	notificationv1.RegisterNotifyServiceServer(grpcSrv, notifygrpc.NewServer(notifySvc))
	notificationv1.RegisterGeneralServiceServer(grpcSrv, generalgrpc.NewServer("ms-notification"))

	// Listen
	errCh := make(chan error, 1)
	go func() {
		s.log.Info("listening", "addr", addr)
		if err := grpcSrv.Serve(lis); err != nil {
			errCh <- fmt.Errorf("serve: %w", err)
		}
	}()

	// Shutdown
	shutdown := func() {
		stopped := make(chan struct{})
		go func() {
			grpcSrv.GracefulStop()
			close(stopped)
		}()
		select {
		case <-stopped:
		case <-time.After(serveShutdownTimeout):
			grpcSrv.Stop()
		}
	}

	// Wait
	select {
	case <-ctx.Done():
		shutdown()
		return nil
	case err := <-errCh:
		shutdown()
		return err
	}
}
