package utils_test

import (
	"errors"
	"testing"

	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestBreakerOpensAfterThreshold(t *testing.T) {
	b := utils.NewBreaker(utils.Settings{BreakerFailureThreshold: 2, BreakerOpenMs: 1000})
	errDown := status.Error(codes.Unavailable, "down")

	if err := b.BeforeCall(); err != nil {
		t.Fatal(err)
	}
	b.Record(errDown)
	if err := b.BeforeCall(); err != nil {
		t.Fatal(err)
	}
	b.Record(errDown)

	openErr := b.BeforeCall()
	if openErr == nil {
		t.Fatal("expected open breaker")
	}
	if _, ok := exception.As(openErr); !ok {
		t.Fatalf("expected exception error, got %v", openErr)
	}
}

func TestBreakerResetsOnSuccess(t *testing.T) {
	b := utils.NewBreaker(utils.Settings{BreakerFailureThreshold: 2, BreakerOpenMs: 1000})
	b.Record(status.Error(codes.Unavailable, "down"))
	b.Record(nil)

	if err := b.BeforeCall(); err != nil {
		t.Fatalf("expected closed after success: %v", err)
	}
}

func TestNewBreakerDisabled(t *testing.T) {
	if utils.NewBreaker(utils.Settings{BreakerFailureThreshold: 0}) != nil {
		t.Fatal("expected nil breaker")
	}
}

func TestIsOutboundUnavailableTransport(t *testing.T) {
	if !utils.IsOutboundUnavailable(errors.New("connection refused")) {
		t.Fatal("expected transport unavailable")
	}
}
