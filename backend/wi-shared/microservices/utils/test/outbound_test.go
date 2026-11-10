package utils_test

import (
	"context"
	"errors"
	"testing"

	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestIsOutboundUnavailable(t *testing.T) {
	t.Parallel()

	if utils.IsOutboundUnavailable(nil) {
		t.Fatal("nil")
	}
	if !utils.IsOutboundUnavailable(status.Error(codes.Unavailable, "down")) {
		t.Fatal("unavailable")
	}
	if !utils.IsOutboundUnavailable(context.DeadlineExceeded) {
		t.Fatal("deadline")
	}
	if !utils.IsOutboundUnavailable(errors.New("connection refused")) {
		t.Fatal("hint")
	}
	if utils.IsOutboundUnavailable(status.Error(codes.InvalidArgument, "bad")) {
		t.Fatal("invalid arg")
	}
	if !utils.IsOutboundUnavailable(errors.New("connection reset by peer")) {
		t.Fatal("reset")
	}
	if !utils.IsOutboundUnavailable(errors.New("unexpected eof")) {
		t.Fatal("eof")
	}
	if utils.IsOutboundUnavailable(context.Canceled) {
		t.Fatal("canceled")
	}
}
