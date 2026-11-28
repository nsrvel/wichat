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
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	generalgrpc "github.com/wichat/wichat/backend/ms-chat/internal/general/delivery/grpc"
	mentionrepo "github.com/wichat/wichat/backend/ms-chat/internal/mention/repository"
	mentionservice "github.com/wichat/wichat/backend/ms-chat/internal/mention/service"
	messagegrpc "github.com/wichat/wichat/backend/ms-chat/internal/message/delivery/grpc"
	messagerepo "github.com/wichat/wichat/backend/ms-chat/internal/message/repository"
	messageservice "github.com/wichat/wichat/backend/ms-chat/internal/message/service"
	outboxrepo "github.com/wichat/wichat/backend/ms-chat/internal/outbox/repository"
	outboxservice "github.com/wichat/wichat/backend/ms-chat/internal/outbox/service"
	searchindexservice "github.com/wichat/wichat/backend/ms-chat/internal/search-index/service"
	searchgrpc "github.com/wichat/wichat/backend/ms-chat/internal/search/delivery/grpc"
	searchkafka "github.com/wichat/wichat/backend/ms-chat/internal/search/delivery/kafka"
	searchrepo "github.com/wichat/wichat/backend/ms-chat/internal/search/repository"
	searchservice "github.com/wichat/wichat/backend/ms-chat/internal/search/service"
	"github.com/wichat/wichat/backend/wi-shared/infra/elasticsearch"
	wikafka "github.com/wichat/wichat/backend/wi-shared/infra/kafka"
	"github.com/wichat/wichat/backend/wi-shared/infra/minio"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
	msUser "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	msgrpc "github.com/wichat/wichat/backend/wi-shared/microservices/utils/grpc"
	chatv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/chat/v1"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

const serveShutdownTimeout = 5 * time.Second

func (s *server) Run(ctx context.Context) error {
	if err := s.cfg.Validate(); err != nil {
		return err
	}

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

	// User service
	if strings.TrimSpace(s.cfg.MSUserHost) == "" {
		return errors.New("MS_USER_HOST is required")
	}

	userConn, err := msgrpc.Dial(s.cfg.MSUserHost, msgrpc.DialOptions{
		Insecure:      true,
		InternalToken: s.cfg.InternalGrpcToken,
	})
	if err != nil {
		return fmt.Errorf("ms-user dial: %w", err)
	}
	defer func() { _ = userConn.Close() }()

	userGRPC := msUser.NewGRPC(userConn, utils.DefaultSettings, s.log)

	// Kafka
	kafkaBrokers := wikafka.ParseBrokers(s.cfg.KafkaBrokers)
	if len(kafkaBrokers) == 0 {
		return errors.New("KAFKA_BROKERS is required")
	}
	topic := strings.TrimSpace(s.cfg.KafkaChatEventsTopic)
	if topic == "" {
		return errors.New("KAFKA_CHAT_EVENTS_TOPIC is required")
	}

	kafkaProducer, err := wikafka.NewProducer(wikafka.Config{Brokers: kafkaBrokers}, s.log)
	if err != nil {
		return fmt.Errorf("kafka producer: %w", err)
	}
	if kafkaProducer == nil {
		return errors.New("kafka producer: unavailable")
	}
	defer func() { _ = kafkaProducer.Close() }()

	// Elasticsearch indexer (search query API/UI deferred; index stays wired)
	esURL := strings.TrimSpace(s.cfg.ElasticsearchURL)
	if esURL == "" {
		return errors.New("ELASTICSEARCH_URL is required")
	}
	esIndex := strings.TrimSpace(s.cfg.ElasticsearchMessagesIdx)
	if esIndex == "" {
		return errors.New("ELASTICSEARCH_MESSAGES_INDEX is required")
	}

	esClient, err := elasticsearch.New(elasticsearch.Config{
		URL:   esURL,
		Index: esIndex,
	}, s.log)
	if err != nil {
		return fmt.Errorf("elasticsearch: %w", err)
	}

	groupID := strings.TrimSpace(s.cfg.KafkaConsumerGroup)
	if groupID == "" {
		return errors.New("KAFKA_CONSUMER_GROUP is required")
	}

	kafkaConsumer, err := wikafka.NewConsumer(wikafka.Config{
		Brokers: kafkaBrokers,
		Topic:   topic,
		GroupID: groupID,
	}, s.log)
	if err != nil {
		return fmt.Errorf("kafka consumer: %w", err)
	}
	defer func() { _ = kafkaConsumer.Close() }()

	// MinIO (channel attachments; empty endpoint disables uploads)
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

	// Services
	outboxPg := outboxrepo.NewPgRepository(db, s.log)
	messagePg := messagerepo.NewPgRepository(db, outboxPg, s.log)
	messageMinio := messagerepo.NewMinioRepository(minioClient)
	outboxSvc := outboxservice.NewService(outboxPg, kafkaProducer, topic, s.log)
	esRepo := searchrepo.NewEsRepository(esClient)
	searchIndexSvc := searchindexservice.NewService(esRepo, s.log)
	mentionPg := mentionrepo.NewPgRepository(db, s.log)
	mentionSvc := mentionservice.NewService(mentionPg, userGRPC.Channel, userGRPC.Auth, s.log)
	searchQuerySvc := searchservice.NewQueryService(esRepo, userGRPC.Channel, s.log)
	searchIndexer := searchkafka.NewSearchIndexer(kafkaConsumer, searchIndexSvc, s.log)
	messageSvc := messageservice.NewService(messagePg, mentionPg, messageMinio, userGRPC.Auth, userGRPC.Channel, userGRPC.GroupRole, outboxPg, s.log)

	// gRPC
	addr := s.cfg.ListenAddr()
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	grpcSrv := newGRPCServer(s.log, s.cfg.InternalGrpcToken)
	chatv1.RegisterMessageServiceServer(grpcSrv, messagegrpc.NewServer(messageSvc))
	chatv1.RegisterSearchServiceServer(grpcSrv, searchgrpc.NewServer(searchQuerySvc, mentionSvc))
	chatv1.RegisterGeneralServiceServer(grpcSrv, generalgrpc.NewServer("ms-chat"))

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
		delay := 2 * time.Second
		const maxDelay = 10 * time.Second
		for {
			err := searchIndexer.Run(ctx)
			if err == nil || ctx.Err() != nil {
				return
			}
			s.log.Warn("search indexer stopped; retrying", "err", err, "retry_in", delay.String())
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
			delay *= 2
			if delay > maxDelay {
				delay = maxDelay
			}
		}
	}()

	go func() {
		delay := 2 * time.Second
		const maxDelay = 10 * time.Second
		for {
			err := outboxSvc.RunRelay(ctx)
			if err == nil || ctx.Err() != nil {
				return
			}
			s.log.Warn("outbox relay stopped; retrying", "err", err, "retry_in", delay.String())
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
			delay *= 2
			if delay > maxDelay {
				delay = maxDelay
			}
		}
	}()

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
