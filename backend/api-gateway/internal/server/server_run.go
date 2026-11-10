package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	kafkaconsumers "github.com/wichat/wichat/backend/api-gateway/internal/kafka/consumers"
	"github.com/wichat/wichat/backend/api-gateway/internal/realtime/hub"
	wikafka "github.com/wichat/wichat/backend/wi-shared/infra/kafka"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
)

const httpShutdownTimeout = 5 * time.Second

func (s *server) Run(ctx context.Context) error {
	// User service
	ms, err := microservices.Open(microservices.Config{
		Log:              s.log,
		Settings:         s.cfg.Microservices,
		UserAddr:         s.cfg.UserHost,
		ChatAddr:         s.cfg.ChatHost,
		NotificationAddr: s.cfg.NotificationHost,
	})
	if err != nil {
		return err
	}
	defer func() { _ = ms.Close() }()

	// Kafka
	kafkaBrokers := wikafka.ParseBrokers(s.cfg.KafkaBrokers)
	if len(kafkaBrokers) == 0 {
		return errors.New("KAFKA_BROKERS is required")
	}
	topic := strings.TrimSpace(s.cfg.KafkaChatEventsTopic)
	if topic == "" {
		return errors.New("KAFKA_CHAT_EVENTS_TOPIC is required")
	}

	kafkaConsumer, err := wikafka.NewConsumer(wikafka.Config{
		Brokers: kafkaBrokers,
		Topic:   topic,
		GroupID: s.cfg.RealtimeKafkaGroupID(),
	}, s.log)
	if err != nil {
		return fmt.Errorf("kafka consumer: %w", err)
	}
	defer func() { _ = kafkaConsumer.Close() }()

	// Realtime hub
	rtHub := hub.New()
	fanout := kafkaconsumers.NewRealtimeFanout(kafkaConsumer, rtHub, s.log)

	// Handler
	handler, err := s.buildHandler(ms, rtHub)
	if err != nil {
		return err
	}

	// HTTP server
	httpSrv := &http.Server{
		Addr:              s.cfg.HTTPListenAddr(),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Listen
	errCh := make(chan error, 2)
	go func() {
		if err := fanout.Run(ctx); err != nil {
			errCh <- fmt.Errorf("realtime fanout: %w", err)
		}
	}()

	go func() {
		s.log.Info("http listening", "addr", s.cfg.HTTPListenAddr(), "kafka_group", s.cfg.RealtimeKafkaGroupID())
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("http serve: %w", err)
		}
	}()

	// Shutdown
	shutdown := func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), httpShutdownTimeout)
		defer cancel()
		_ = httpSrv.Shutdown(stopCtx)
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
