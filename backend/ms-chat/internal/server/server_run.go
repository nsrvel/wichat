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
	kafkaconsumers "github.com/wichat/wichat/backend/ms-chat/internal/kafka/consumers"
	kafkaproducers "github.com/wichat/wichat/backend/ms-chat/internal/kafka/producers"
	messagegrpc "github.com/wichat/wichat/backend/ms-chat/internal/message/delivery/grpc"
	messagerepo "github.com/wichat/wichat/backend/ms-chat/internal/message/repository"
	messageservice "github.com/wichat/wichat/backend/ms-chat/internal/message/service"
	searchrepo "github.com/wichat/wichat/backend/ms-chat/internal/search/repository"
	"github.com/wichat/wichat/backend/wi-shared/infra/elasticsearch"
	wikafka "github.com/wichat/wichat/backend/wi-shared/infra/kafka"
	msUser "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	msgrpc "github.com/wichat/wichat/backend/wi-shared/microservices/utils/grpc"
	chatv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/chat/v1"
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

	searchIndexer := kafkaconsumers.NewSearchIndexer(
		kafkaConsumer,
		searchrepo.NewEsRepository(esClient),
		s.log,
	)

	// Services
	messagePg := messagerepo.NewPgRepository(db, s.log)
	eventPub := kafkaproducers.NewMessageCreated(kafkaProducer, topic, s.log)
	messageSvc := messageservice.NewService(messagePg, userGRPC.Channel, eventPub, s.log)

	// gRPC
	addr := s.cfg.ListenAddr()
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	grpcSrv := newGRPCServer(s.log)
	chatv1.RegisterMessageServiceServer(grpcSrv, messagegrpc.NewServer(messageSvc, userGRPC.Channel))
	chatv1.RegisterGeneralServiceServer(grpcSrv, generalgrpc.NewServer("ms-chat"))

	// Listen
	errCh := make(chan error, 2)
	go func() {
		if err := searchIndexer.Run(ctx); err != nil {
			errCh <- fmt.Errorf("search indexer: %w", err)
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
