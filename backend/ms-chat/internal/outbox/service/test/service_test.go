package service_test

import (
	"context"
	"testing"
	"time"

	outboxmock "github.com/wichat/wichat/backend/ms-chat/internal/outbox/repository/mock"
	outboxservice "github.com/wichat/wichat/backend/ms-chat/internal/outbox/service"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

func TestRunRelay_noRows(t *testing.T) {
	st := outboxmock.NewState()
	appLog := logger.New(logger.Options{Level: "error", Service: "ms-chat-test"})
	svc := outboxservice.NewService(outboxmock.NewPgRepositoryMock(st), nil, "chat-events", appLog)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_ = svc.RunRelay(ctx)
}
