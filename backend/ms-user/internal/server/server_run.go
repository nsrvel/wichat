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
	channelgrpc "github.com/wichat/wichat/backend/ms-user/internal/channel/delivery/grpc"
	channelrepo "github.com/wichat/wichat/backend/ms-user/internal/channel/repository"
	channelservice "github.com/wichat/wichat/backend/ms-user/internal/channel/service"
	generalgrpc "github.com/wichat/wichat/backend/ms-user/internal/general/delivery/grpc"
	groupmembergrpc "github.com/wichat/wichat/backend/ms-user/internal/group-member/delivery/grpc"
	groupmemberrepo "github.com/wichat/wichat/backend/ms-user/internal/group-member/repository"
	groupmemberservice "github.com/wichat/wichat/backend/ms-user/internal/group-member/service"
	grouprolegrpc "github.com/wichat/wichat/backend/ms-user/internal/group-role/delivery/grpc"
	grouprolerepo "github.com/wichat/wichat/backend/ms-user/internal/group-role/repository"
	grouproleservice "github.com/wichat/wichat/backend/ms-user/internal/group-role/service"
	groupgrpc "github.com/wichat/wichat/backend/ms-user/internal/group/delivery/grpc"
	grouprepo "github.com/wichat/wichat/backend/ms-user/internal/group/repository"
	groupservice "github.com/wichat/wichat/backend/ms-user/internal/group/service"
	invitegrpc "github.com/wichat/wichat/backend/ms-user/internal/invite/delivery/grpc"
	inviterepo "github.com/wichat/wichat/backend/ms-user/internal/invite/repository"
	inviteservice "github.com/wichat/wichat/backend/ms-user/internal/invite/service"
	usergrpc "github.com/wichat/wichat/backend/ms-user/internal/user/delivery/grpc"
	userrepo "github.com/wichat/wichat/backend/ms-user/internal/user/repository"
	userservice "github.com/wichat/wichat/backend/ms-user/internal/user/service"
	googleoauth "github.com/wichat/wichat/backend/ms-user/pkg/oauth"
	"github.com/wichat/wichat/backend/wi-shared/infra/minio"
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

	// Repositories
	userPg := userrepo.NewPgRepository(db)
	groupPg := grouprepo.NewPgRepository(db, s.log)
	groupMemberPg := groupmemberrepo.NewPgRepository(db, s.log)
	groupRolePg := grouprolerepo.NewPgRepository(db, s.log)
	minioClient, err := minio.New(minio.Config{
		Endpoint:      s.cfg.MinioEndpoint,
		AccessKey:     s.cfg.MinioAccessKey,
		SecretKey:     s.cfg.MinioSecretKey,
		Bucket:        s.cfg.MinioBucket,
		PublicBaseURL: s.cfg.MinioPublicBaseURL,
		UseSSL:        s.cfg.MinioUseSSL,
		Region:        s.cfg.MinioRegion,
	}, s.log)
	if err != nil {
		return fmt.Errorf("minio: %w", err)
	}
	if err := minio.EnsurePublicReadBucket(ctx, minioClient); err != nil {
		return fmt.Errorf("minio bootstrap: %w", err)
	}
	groupMinio := grouprepo.NewMinioRepository(minioClient)
	channelPg := channelrepo.NewPgRepository(db, s.log)
	invitePg := inviterepo.NewPgRepository(db, s.log)
	authSessions := authrepo.NewSessionRepository(db)
	authRefresh := authrepo.NewRedisRefreshStore(rdb)
	secret := []byte(s.cfg.AuthJWTSecret)

	// Services
	userSvc := userservice.NewService(userPg, groupPg, groupMemberPg, authSessions, secret, s.log)
	channelSvc := channelservice.NewService(channelPg, groupPg, groupMemberPg, groupRolePg, authSessions, secret, s.log)
	groupSvc := groupservice.NewService(groupPg, groupMemberPg, groupRolePg, channelSvc, userPg, groupMinio, authSessions, secret, s.log)
	groupMemberSvc := groupmemberservice.NewService(groupMemberPg, groupPg, groupRolePg, authSessions, secret, s.log)
	groupRoleSvc := grouproleservice.NewService(groupRolePg, groupPg, groupMemberPg, authSessions, secret, s.log)
	inviteSvc := inviteservice.NewService(
		groupPg, groupMemberPg, groupRolePg, channelPg, invitePg, userPg, authSessions, notifyClient, secret,
		s.cfg.WebOrigin, s.cfg.MinioPublicBaseURL, s.log,
	)
	authSvc, err := authservice.NewService(userPg, authSessions, authRefresh, secret, s.log, googleoauth.Config{
		ClientID:     s.cfg.GoogleClientID,
		ClientSecret: s.cfg.GoogleClientSecret,
		RedirectURL:  s.cfg.GoogleRedirectURL,
	}, authrepo.NewRedisPasswordResetStore(rdb), authrepo.NewRedisEmailAuthStore(rdb), notifyClient, userSvc, nil)
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
	userv1.RegisterAuthServiceServer(grpcSrv, authgrpc.NewServer(authSvc))
	userv1.RegisterUserServiceServer(grpcSrv, usergrpc.NewServer(userSvc))
	userv1.RegisterGroupServiceServer(grpcSrv, groupgrpc.NewServer(groupSvc))
	userv1.RegisterGroupMemberServiceServer(grpcSrv, groupmembergrpc.NewServer(groupMemberSvc))
	userv1.RegisterGroupRoleServiceServer(grpcSrv, grouprolegrpc.NewServer(groupRoleSvc))
	userv1.RegisterInviteServiceServer(grpcSrv, invitegrpc.NewServer(inviteSvc))
	userv1.RegisterChannelServiceServer(grpcSrv, channelgrpc.NewServer(channelSvc))
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
