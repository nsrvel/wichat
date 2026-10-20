package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	authgrpc "github.com/wichat/wichat/backend/ms-user/internal/auth/delivery/grpc"
	authrepo "github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	authservice "github.com/wichat/wichat/backend/ms-user/internal/auth/service"
	generalgrpc "github.com/wichat/wichat/backend/ms-user/internal/general/delivery/grpc"
	groupservice "github.com/wichat/wichat/backend/ms-user/internal/group/service"
	userrepo "github.com/wichat/wichat/backend/ms-user/internal/user/repository"
	userservice "github.com/wichat/wichat/backend/ms-user/internal/user/service"
	googleoauth "github.com/wichat/wichat/backend/ms-user/pkg/oauth"
	msNotification "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-notification"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	msgrpc "github.com/wichat/wichat/backend/wi-shared/microservices/utils/grpc"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"google.golang.org/grpc"
)

const serveShutdownTimeout = 5 * time.Second

func (s *server) Run(ctx context.Context) error {
	// Database
	if strings.TrimSpace(s.cfg.DatabaseURL) == "" {
		return errors.New("DATABASE_URL is required")
	}
	if len(s.cfg.AuthJWTSecret) < 32 {
		return errors.New("AUTH_JWT_SECRET must be at least 32 characters")
	}

	sqldb, err := sql.Open("pgx", s.cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("postgres: %w", err)
	}
	defer func() { _ = sqldb.Close() }()

	if err := sqldb.PingContext(ctx); err != nil {
		return fmt.Errorf("postgres ping: %w", err)
	}

	db := bun.NewDB(sqldb, pgdialect.New())
	defer func() { _ = db.Close() }()

	// Redis
	redisURL := strings.TrimSpace(s.cfg.RedisURL)
	if redisURL == "" {
		return errors.New("redis url empty")
	}
	redisOpts, err := redis.ParseURL(redisURL)
	if err != nil {
		return fmt.Errorf("redis: %w", err)
	}
	rdb := redis.NewClient(redisOpts)
	defer func() { _ = rdb.Close() }()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping: %w", err)
	}

	// Notification
	var notifConn *grpc.ClientConn
	var notifyClient authservice.Notifier
	if strings.TrimSpace(s.cfg.NotificationHost) != "" {
		notifConn, err = msgrpc.Dial(s.cfg.NotificationHost, msgrpc.DefaultDialOptions)
		if err != nil {
			return fmt.Errorf("ms-notification dial: %w", err)
		}
		defer func() { _ = notifConn.Close() }()
		notifyClient = msNotification.NewGRPC(notifConn, utils.DefaultSettings, s.log).Notify
	}

	// Services
	userPg := userrepo.NewPgRepository(db)
	authPg := authrepo.NewPgRepository(db)
	secret := []byte(s.cfg.AuthJWTSecret)
	userSvc := userservice.NewService(userPg, authPg, secret, s.log)
	groupSvc := groupservice.NewService(userPg, authPg, secret, s.log)
	authSvc, err := authservice.NewService(userPg, authPg, secret, s.log, googleoauth.Config{
		ClientID:     s.cfg.GoogleClientID,
		ClientSecret: s.cfg.GoogleClientSecret,
		RedirectURL:  s.cfg.GoogleRedirectURL,
	}, authrepo.NewRedisPasswordResetStore(rdb), authrepo.NewRedisEmailAuthStore(rdb), notifyClient, userSvc)
	if err != nil {
		return err
	}

	// gRPC
	addr := s.cfg.ListenAddr()
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	grpcSrv := newGRPCServer(s.log)
	userv1.RegisterAuthServiceServer(grpcSrv, authgrpc.NewServer(authSvc, userSvc, groupSvc))
	userv1.RegisterGeneralServiceServer(grpcSrv, generalgrpc.NewServer("ms-user"))

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
