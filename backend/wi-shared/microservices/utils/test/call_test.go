// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

package utils_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils/correlation"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRequestIDContext(t *testing.T) {
	ctx := correlation.WithRequestID(context.Background(), "abc")
	if correlation.RequestIDFromContext(ctx) != "abc" {
		t.Fatalf("expected request id abc")
	}
	if correlation.RequestIDFromContext(context.Background()) != "" {
		t.Fatal("expected empty id")
	}
}

func TestCallUnaryNoRetryOnClientError(t *testing.T) {
	cfg := utils.CallConfig{
		PerAttemptTimeout: 50 * time.Millisecond,
		OverallTimeout:    200 * time.Millisecond,
		MaxRetries:        3,
		RetryBackoff:      1 * time.Millisecond,
		RetryPolicy:       utils.RetryIdempotent,
	}
	var calls atomic.Int32
	err := utils.CallUnary(context.Background(), cfg, nil, "test.Method", func(context.Context) error {
		calls.Add(1)
		return status.Error(codes.InvalidArgument, "bad input")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls.Load() != 1 {
		t.Fatalf("expected 1 call, got %d", calls.Load())
	}
}

func TestCallUnaryRetriesUnavailable(t *testing.T) {
	cfg := utils.CallConfig{
		PerAttemptTimeout: 50 * time.Millisecond,
		OverallTimeout:    500 * time.Millisecond,
		MaxRetries:        2,
		RetryBackoff:      1 * time.Millisecond,
		RetryPolicy:       utils.RetryIdempotent,
	}
	var calls atomic.Int32
	err := utils.CallUnary(context.Background(), cfg, nil, "test.Method", func(context.Context) error {
		calls.Add(1)
		return status.Error(codes.Unavailable, "down")
	})
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.Unavailable {
		t.Fatalf("unexpected err %v", err)
	}
	if calls.Load() != 3 {
		t.Fatalf("expected 3 calls (1 + 2 retries), got %d", calls.Load())
	}
}

func TestCallConfigFromSettings(t *testing.T) {
	cfg := utils.CallConfigFromSettings(utils.DefaultSettings, utils.RetryIdempotent)
	if cfg.MaxRetries != 3 || cfg.PerAttemptTimeout != 5*time.Second {
		t.Fatalf("unexpected cfg %+v", cfg)
	}
}
