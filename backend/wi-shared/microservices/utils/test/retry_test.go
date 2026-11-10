package utils_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
