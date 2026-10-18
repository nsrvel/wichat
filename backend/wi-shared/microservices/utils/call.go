// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

package utils

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// CallUnary runs fn with per-attempt and overall deadlines and optional retries.
func CallUnary(ctx context.Context, cfg CallConfig, log logger.Logger, method string, fn func(context.Context) error) error {
	if cfg.PerAttemptTimeout <= 0 || cfg.OverallTimeout <= 0 {
		return fmt.Errorf("microservices/utils: invalid call timeouts")
	}

	if cfg.Breaker != nil {
		if err := cfg.Breaker.BeforeCall(); err != nil {
			return err
		}
	}

	overallCtx, cancelOverall := context.WithTimeout(ctx, cfg.OverallTimeout)
	defer cancelOverall()

	var lastErr error
	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		attemptCtx, cancelAttempt := context.WithTimeout(overallCtx, cfg.PerAttemptTimeout)
		lastErr = fn(attemptCtx)
		cancelAttempt()

		if lastErr == nil {
			if cfg.Breaker != nil {
				cfg.Breaker.Record(nil)
			}
			return nil
		}
		if !shouldRetry(cfg.RetryPolicy, lastErr, attempt, cfg.MaxRetries) {
			if cfg.Breaker != nil {
				cfg.Breaker.Record(lastErr)
			}
			return lastErr
		}

		delay := retryDelay(attempt, cfg.RetryBackoff, lastErr)
		if log != nil {
			code := codes.Unknown
			if st, ok := status.FromError(lastErr); ok {
				code = st.Code()
			}
			log.Warn("grpc call retry",
				"method", method,
				"attempt", attempt+1,
				"max_retries", cfg.MaxRetries,
				"delay_ms", delay.Milliseconds(),
				"code", code.String(),
				"err", lastErr,
			)
		}

		if delay > 0 {
			timer := time.NewTimer(delay)
			select {
			case <-overallCtx.Done():
				timer.Stop()
				return lastErr
			case <-timer.C:
			}
		}
	}

	if cfg.Breaker != nil {
		cfg.Breaker.Record(lastErr)
	}

	return lastErr
}
