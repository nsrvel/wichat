package utils_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
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

func TestCallUnaryResultReturnsValue(t *testing.T) {
	cfg := utils.CallConfig{
		PerAttemptTimeout: 50 * time.Millisecond,
		OverallTimeout:    200 * time.Millisecond,
		MaxRetries:        0,
		RetryPolicy:       utils.RetryNone,
	}
	out, err := utils.CallUnaryResult(context.Background(), cfg, nil, "test.Method", func(context.Context) (string, error) {
		return "ok", nil
	})
	if err != nil || out != "ok" {
		t.Fatalf("unexpected out=%q err=%v", out, err)
	}
}

func TestCallUnaryResourceExhaustedRetries(t *testing.T) {
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
		return status.Error(codes.ResourceExhausted, "throttled")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls.Load() != 3 {
		t.Fatalf("calls=%d", calls.Load())
	}
}

func TestCallUnaryWithBreakerSuccess(t *testing.T) {
	br := utils.NewBreaker(utils.Settings{BreakerFailureThreshold: 2, BreakerOpenMs: 1000})
	cfg := utils.CallConfig{
		PerAttemptTimeout: 50 * time.Millisecond,
		OverallTimeout:    200 * time.Millisecond,
		MaxRetries:        0,
		RetryPolicy:       utils.RetryNone,
		Breaker:           br,
	}
	if err := utils.CallUnary(context.Background(), cfg, nil, "ok", func(context.Context) error { return nil }); err != nil {
		t.Fatal(err)
	}
}

func TestCallUnaryRejectsInvalidTimeouts(t *testing.T) {
	err := utils.CallUnary(context.Background(), utils.CallConfig{}, nil, "x", func(context.Context) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCallUnaryBreakerOpen(t *testing.T) {
	br := utils.NewBreaker(utils.Settings{BreakerFailureThreshold: 1, BreakerOpenMs: 60_000})
	br.Record(status.Error(codes.Unavailable, "down"))

	cfg := utils.CallConfig{
		PerAttemptTimeout: 50 * time.Millisecond,
		OverallTimeout:    200 * time.Millisecond,
		MaxRetries:        0,
		RetryPolicy:       utils.RetryNone,
		Breaker:           br,
	}
	err := utils.CallUnary(context.Background(), cfg, nil, "x", func(context.Context) error { return nil })
	if err == nil {
		t.Fatal("expected breaker error")
	}
}

func TestCallUnaryBreakerRecordsFailure(t *testing.T) {
	br := utils.NewBreaker(utils.Settings{BreakerFailureThreshold: 5, BreakerOpenMs: 60_000})
	cfg := utils.CallConfig{
		PerAttemptTimeout: 50 * time.Millisecond,
		OverallTimeout:    200 * time.Millisecond,
		MaxRetries:        0,
		RetryPolicy:       utils.RetryNone,
		Breaker:           br,
	}
	err := utils.CallUnary(context.Background(), cfg, nil, "x", func(context.Context) error {
		return status.Error(codes.NotFound, "missing")
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCallUnaryDeadlineExceededRetries(t *testing.T) {
	cfg := utils.CallConfig{
		PerAttemptTimeout: 50 * time.Millisecond,
		OverallTimeout:    500 * time.Millisecond,
		MaxRetries:        2,
		RetryBackoff:      1 * time.Millisecond,
		RetryPolicy:       utils.RetryIdempotent,
	}
	var calls atomic.Int32
	err := utils.CallUnary(context.Background(), cfg, logger.New(logger.Options{Service: "t", Env: "test", Level: "warn"}), "x", func(context.Context) error {
		calls.Add(1)
		return status.Error(codes.DeadlineExceeded, "timeout")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls.Load() != 3 {
		t.Fatalf("calls=%d", calls.Load())
	}
}

func TestCallUnaryContextCanceledNoRetry(t *testing.T) {
	cfg := utils.CallConfig{
		PerAttemptTimeout: 50 * time.Millisecond,
		OverallTimeout:    500 * time.Millisecond,
		MaxRetries:        3,
		RetryBackoff:      1 * time.Millisecond,
		RetryPolicy:       utils.RetryIdempotent,
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var calls atomic.Int32
	err := utils.CallUnary(ctx, cfg, nil, "x", func(context.Context) error {
		calls.Add(1)
		return status.Error(codes.Aborted, "aborted")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls.Load() != 1 {
		t.Fatalf("calls=%d", calls.Load())
	}
}

func TestCallUnaryOverallCanceledDuringBackoff(t *testing.T) {
	cfg := utils.CallConfig{
		PerAttemptTimeout: 50 * time.Millisecond,
		OverallTimeout:    80 * time.Millisecond,
		MaxRetries:        5,
		RetryBackoff:      50 * time.Millisecond,
		RetryPolicy:       utils.RetryIdempotent,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()

	err := utils.CallUnary(ctx, cfg, nil, "x", func(context.Context) error {
		return status.Error(codes.Unavailable, "down")
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
