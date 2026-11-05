package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/wichat/wichat/backend/wi-shared/microservices"
)

const httpShutdownTimeout = 5 * time.Second

func (s *server) Run(ctx context.Context) error {
	// User service
	ms, err := microservices.Open(microservices.Config{
		Log:              s.log,
		Settings:         s.cfg.Microservices,
		UserAddr:         s.cfg.UserHost,
		NotificationAddr: s.cfg.NotificationHost,
	})
	if err != nil {
		return err
	}
	defer func() { _ = ms.Close() }()

	// Handler
	handler, err := s.buildHandler(ms)
	if err != nil {
		return err
	}

	// HTTP server
	httpSrv := &http.Server{
		Addr:              s.cfg.HTTPListenAddr(),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Listen
	errCh := make(chan error, 1)
	go func() {
		s.log.Info("http listening", "addr", s.cfg.HTTPListenAddr())
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
