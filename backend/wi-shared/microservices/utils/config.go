// Package utils holds shared outbound call settings, retries, and gRPC helpers.
package utils

import "time"

// RetryPolicy controls whether CallUnary may repeat failed attempts.
type RetryPolicy int

const (
	// RetryNone never retries.
	RetryNone RetryPolicy = 0
	// RetryIdempotent retries transport/unavailable-class errors (reads, Health, etc.).
	RetryIdempotent RetryPolicy = 1
	// RetrySafe reserved for idempotency-key writes; same retriable set as RetryIdempotent today.
	RetrySafe RetryPolicy = 2
)

// Settings holds default timeout/retry values (mapstructure on per-service config).
type Settings struct {
	AttemptTimeoutMs        int `mapstructure:"ms_attempt_timeout_ms"`
	OverallTimeoutMs        int `mapstructure:"ms_overall_timeout_ms"`
	MaxRetries              int `mapstructure:"ms_max_retries"`
	RetryBackoffMs          int `mapstructure:"ms_retry_backoff_ms"`
	BreakerFailureThreshold int `mapstructure:"ms_breaker_failure_threshold"`
	BreakerOpenMs           int `mapstructure:"ms_breaker_open_ms"`
}

// DefaultSettings matches Nest-style microservice client defaults.
var DefaultSettings = Settings{
	AttemptTimeoutMs:        5000,
	OverallTimeoutMs:        15000,
	MaxRetries:              3,
	RetryBackoffMs:          100,
	BreakerFailureThreshold: 5,
	BreakerOpenMs:           30000,
}

// CallConfig is the per-call override for CallUnary.
type CallConfig struct {
	PerAttemptTimeout time.Duration
	OverallTimeout    time.Duration
	MaxRetries        int
	RetryBackoff      time.Duration
	RetryPolicy       RetryPolicy
	Breaker           *Breaker
}

// CallConfigFromSettings builds CallConfig with the given retry policy.
func CallConfigFromSettings(s Settings, policy RetryPolicy) CallConfig {
	return CallConfig{
		PerAttemptTimeout: time.Duration(s.AttemptTimeoutMs) * time.Millisecond,
		OverallTimeout:    time.Duration(s.OverallTimeoutMs) * time.Millisecond,
		MaxRetries:        s.MaxRetries,
		RetryBackoff:      time.Duration(s.RetryBackoffMs) * time.Millisecond,
		RetryPolicy:       policy,
	}
}
