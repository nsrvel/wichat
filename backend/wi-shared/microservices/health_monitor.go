package microservices

import (
	"context"
	"sync"
	"time"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	chatv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/chat/v1"
	notificationv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/notification/v1"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

const (
	defaultHealthyPollInterval   = 5 * time.Second
	defaultUnhealthyPollInterval = 2 * time.Second
	defaultHealthCheckTimeout    = 4 * time.Second
)

// HealthTarget describes one downstream gRPC service monitored via GeneralService.Health.
type HealthTarget struct {
	Name  string
	Conn  *grpc.ClientConn
	Check func(ctx context.Context) error
}

// HealthMonitor periodically probes downstream gRPC services and resets connect backoff on recovery.
type HealthMonitor struct {
	targets         []HealthTarget
	log             logger.Logger
	healthyInterval time.Duration
	unhealthyIntvl  time.Duration
	checkTimeout    time.Duration

	mu     sync.RWMutex
	status map[string]bool
}

// NewHealthMonitor constructs a HealthMonitor for the given gRPC targets.
func NewHealthMonitor(log logger.Logger, targets ...HealthTarget) *HealthMonitor {
	status := make(map[string]bool, len(targets))
	filtered := make([]HealthTarget, 0, len(targets))
	for _, t := range targets {
		if t.Name == "" || t.Check == nil {
			continue
		}
		filtered = append(filtered, t)
		status[t.Name] = true
	}
	return &HealthMonitor{
		targets:         filtered,
		log:             log,
		healthyInterval: defaultHealthyPollInterval,
		unhealthyIntvl:  defaultUnhealthyPollInterval,
		checkTimeout:    defaultHealthCheckTimeout,
		status:          status,
	}
}

// Run polls targets until ctx is canceled.
func (m *HealthMonitor) Run(ctx context.Context) {
	if m == nil || len(m.targets) == 0 {
		return
	}

	allHealthy := m.ProbeOnce(ctx)
	for {
		interval := m.healthyInterval
		if !allHealthy {
			interval = m.unhealthyIntvl
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
		allHealthy = m.ProbeOnce(ctx)
	}
}

// ProbeOnce runs one health-check pass across all targets and returns true if all are healthy.
func (m *HealthMonitor) ProbeOnce(ctx context.Context) bool {
	if m == nil || len(m.targets) == 0 {
		return true
	}

	allHealthy := true
	for _, target := range m.targets {
		if target.Conn != nil && target.Conn.GetState() == connectivity.TransientFailure {
			target.Conn.ResetConnectBackoff()
			target.Conn.Connect()
		}

		probeCtx, cancel := context.WithTimeout(ctx, m.checkTimeout)
		err := target.Check(probeCtx)
		cancel()

		healthy := err == nil
		if !healthy {
			allHealthy = false
			if target.Conn != nil {
				target.Conn.ResetConnectBackoff()
				target.Conn.Connect()
			}
		}

		m.mu.Lock()
		prev, existed := m.status[target.Name]
		m.status[target.Name] = healthy
		m.mu.Unlock()

		if m.log != nil && (!existed || prev != healthy) {
			if healthy {
				m.log.Info("downstream service healthy", "service", target.Name)
			} else {
				m.log.Warn("downstream service unhealthy", "service", target.Name, "err", err)
			}
		}
	}

	return allHealthy
}

// Snapshot returns a copy of the latest per-service health map and overall readiness.
func (m *HealthMonitor) Snapshot() (map[string]bool, bool) {
	if m == nil {
		return map[string]bool{}, true
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make(map[string]bool, len(m.status))
	ready := true
	for k, ok := range m.status {
		out[k] = ok
		if !ok {
			ready = false
		}
	}
	return out, ready
}

// StartHealthMonitor starts background GeneralService.Health monitoring for all dialed services in Clients.
func (c *Clients) StartHealthMonitor(ctx context.Context, log logger.Logger) *HealthMonitor {
	if c == nil {
		return NewHealthMonitor(log)
	}
	targets := make([]HealthTarget, 0, 3)
	if c.User.GRPC.General != nil {
		targets = append(targets, HealthTarget{
			Name: "ms-user",
			Conn: c.userConn,
			Check: func(probeCtx context.Context) error {
				_, err := c.User.GRPC.General.Health(probeCtx, &userv1.HealthRequest{})
				return err
			},
		})
	}
	if c.Chat.GRPC.General != nil && c.chatConn != nil {
		targets = append(targets, HealthTarget{
			Name: "ms-chat",
			Conn: c.chatConn,
			Check: func(probeCtx context.Context) error {
				_, err := c.Chat.GRPC.General.Health(probeCtx, &chatv1.HealthRequest{})
				return err
			},
		})
	}
	if c.Notification.GRPC.General != nil && c.notificationConn != nil {
		targets = append(targets, HealthTarget{
			Name: "ms-notification",
			Conn: c.notificationConn,
			Check: func(probeCtx context.Context) error {
				_, err := c.Notification.GRPC.General.Health(probeCtx, &notificationv1.HealthRequest{})
				return err
			},
		})
	}

	monitor := NewHealthMonitor(log, targets...)
	go monitor.Run(ctx)
	return monitor
}
