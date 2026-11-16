package test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
)

func TestHealthMonitorTransitionsAndRecovery(t *testing.T) {
	t.Parallel()

	var fail atomic.Bool
	fail.Store(true)

	mon := microservices.NewHealthMonitor(
		logger.New(logger.Options{Service: "test", Env: "test", Level: "error"}),
		microservices.HealthTarget{
			Name: "ms-user",
			Check: func(context.Context) error {
				if fail.Load() {
					return errors.New("connection refused")
				}
				return nil
			},
		},
	)

	if ok := mon.ProbeOnce(context.Background()); ok {
		t.Fatal("expected ProbeOnce to return false when target fails")
	}
	snap, ready := mon.Snapshot()
	if ready || snap["ms-user"] {
		t.Fatalf("expected degraded snapshot, got ready=%v snap=%v", ready, snap)
	}

	fail.Store(false)
	if ok := mon.ProbeOnce(context.Background()); !ok {
		t.Fatal("expected ProbeOnce to return true after target recovers")
	}
	snap, ready = mon.Snapshot()
	if !ready || !snap["ms-user"] {
		t.Fatalf("expected healthy snapshot after recovery, got ready=%v snap=%v", ready, snap)
	}
}
