package test

import (
	"context"
	"testing"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
)

func TestClientsOpenFailsOnEmptyAddr(t *testing.T) {
	log := logger.New(logger.Options{Service: "test", Env: "test"})
	_, err := microservices.Open(microservices.Config{
		Log:      log,
		UserAddr: "",
	})
	if err == nil {
		t.Fatal("expected error on empty user addr")
	}
}

func TestClientsStartHealthMonitorNil(t *testing.T) {
	var c *microservices.Clients
	mon := c.StartHealthMonitor(context.Background(), nil)
	if mon == nil {
		t.Fatal("expected non-nil monitor for nil clients")
	}
}
