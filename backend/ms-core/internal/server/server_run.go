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
	generalgrpc "github.com/wichat/wichat/backend/ms-core/internal/general/delivery/grpc"
	"github.com/wichat/wichat/backend/ms-core/internal/kafka/producers"
	voicegrpc "github.com/wichat/wichat/backend/ms-core/internal/voice/delivery/grpc"
	voicerepo "github.com/wichat/wichat/backend/ms-core/internal/voice/repository"
	voiceservice "github.com/wichat/wichat/backend/ms-core/internal/voice/service"
	livekitclient "github.com/wichat/wichat/backend/ms-core/pkg/livekit-client"
	wikafka "github.com/wichat/wichat/backend/wi-shared/infra/kafka"
	wiminio "github.com/wichat/wichat/backend/wi-shared/infra/minio"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
	msUser "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	msgrpc "github.com/wichat/wichat/backend/wi-shared/microservices/utils/grpc"
	corev1 "github.com/wichat/wichat/backend/wi-shared/models/gen/core/v1"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

const serveShutdownTimeout = 5 * time.Second

func (s *server) Run(ctx context.Context) error {
	// Database
	if strings.TrimSpace(s.cfg.DatabaseURL) == "" {
		return errors.New("DATABASE_URL is required")
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
	redisOpts, err := redis.ParseURL(s.cfg.RedisURL)
	if err != nil {
		return fmt.Errorf("redis parse url: %w", err)
	}
	redisClient := redis.NewClient(redisOpts)
	defer func() { _ = redisClient.Close() }()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping: %w", err)
	}

	// MinIO
	minioClient, err := wiminio.New(wiminio.Config{
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

	// Kafka
	var brokers []string
	for _, b := range strings.Split(s.cfg.KafkaBrokers, ",") {
		if trimmed := strings.TrimSpace(b); trimmed != "" {
			brokers = append(brokers, trimmed)
		}
	}
	kafkaProducer, err := wikafka.NewProducer(wikafka.Config{Brokers: brokers}, s.log)
	if err != nil {
		return fmt.Errorf("kafka producer: %w", err)
	}
	if kafkaProducer != nil {
		defer func() { _ = kafkaProducer.Close() }()
	}

	// User service
	if strings.TrimSpace(s.cfg.MSUserHost) == "" {
		return errors.New("MS_USER_HOST is required")
	}

	userConn, err := msgrpc.Dial(s.cfg.MSUserHost, msgrpc.DefaultDialOptions)
	if err != nil {
		return fmt.Errorf("ms-user dial: %w", err)
	}
	defer func() { _ = userConn.Close() }()

	userGRPC := msUser.NewGRPC(userConn, utils.DefaultSettings, s.log)

	// Voice
	voicePg := voicerepo.NewPgRepository(db, s.log)
	voiceRedis := voicerepo.NewRedisOccupancy(redisClient)
	voiceMinio := voicerepo.NewMinioRepository(minioClient)
	voiceEvents := producers.NewVoiceEventsPublisher(kafkaProducer, s.cfg.KafkaChatEventsTopic, s.log)
	lkClient := livekitclient.New(s.cfg.LiveKitURL, s.cfg.LiveKitAPIKey, s.cfg.LiveKitAPISecret)
	voiceSvc := voiceservice.NewService(
		voicePg,
		voiceRedis,
		voiceMinio,
		lkClient,
		userGRPC.Channel,
		userGRPC.Auth,
		userGRPC.Presence,
		voiceEvents,
		voiceservice.Options{
			LiveKitURL:       s.cfg.LiveKitURL,
			MonthlyQuotaMins: s.cfg.LiveKitMonthlyQuotaMinutes,
			EnforceHardCap:   s.cfg.LiveKitEnforceHardCap,
			MinioEndpoint:    s.cfg.MinioEndpoint,
			MinioAccessKey:   s.cfg.MinioAccessKey,
			MinioSecretKey:   s.cfg.MinioSecretKey,
			MinioBucket:      s.cfg.MinioBucket,
			MinioRegion:      s.cfg.MinioRegion,
			MinioUseSSL:      s.cfg.MinioUseSSL,
		},
		s.log,
	)

	// gRPC
	addr := s.cfg.ListenAddr()
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	grpcSrv := newGRPCServer(s.log)
	corev1.RegisterGeneralServiceServer(grpcSrv, generalgrpc.NewServer("ms-core"))
	corev1.RegisterVoiceServiceServer(grpcSrv, voicegrpc.NewServer(voiceSvc))

	// Idle-kick monitor
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if idleErr := voiceSvc.RunIdleCheck(ctx); idleErr != nil && s.log != nil {
					s.log.Warn("voice idle check", "err", idleErr)
				}
			}
		}
	}()

	// Listen
	go microservices.NewHealthMonitor(s.log, microservices.HealthTarget{
		Name: "ms-user",
		Conn: userConn,
		Check: func(probeCtx context.Context) error {
			_, hErr := userGRPC.General.Health(probeCtx, &userv1.HealthRequest{})
			return hErr
		},
	}).Run(ctx)

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
