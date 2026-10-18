package utils

import (
	"sync"
	"time"

	"github.com/wichat/wichat/backend/wi-shared/exception"
)

// Breaker fails fast after consecutive outbound failures (transport / unavailable).
type Breaker struct {
	mu        sync.Mutex
	threshold int
	openFor   time.Duration

	failures  int
	openUntil time.Time
}

// NewBreaker returns nil when settings disable the breaker (threshold <= 0).
func NewBreaker(s Settings) *Breaker {
	if s.BreakerFailureThreshold <= 0 {
		return nil
	}
	openFor := time.Duration(s.BreakerOpenMs) * time.Millisecond
	if openFor <= 0 {
		openFor = 30 * time.Second
	}

	return &Breaker{
		threshold: s.BreakerFailureThreshold,
		openFor:   openFor,
	}
}

// BeforeCall returns an error when the breaker is open.
func (b *Breaker) BeforeCall() error {
	if b == nil {
		return nil
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	if now.Before(b.openUntil) {
		return exception.Unavailable(exception.MsgUnavailable, nil)
	}
	if !b.openUntil.IsZero() && !now.Before(b.openUntil) {
		b.openUntil = time.Time{}
		b.failures = 0
	}

	return nil
}

// Record updates breaker state after an outbound call.
func (b *Breaker) Record(err error) {
	if b == nil {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if err == nil {
		b.failures = 0
		b.openUntil = time.Time{}

		return
	}
	if !IsOutboundUnavailable(err) {
		return
	}

	b.failures++
	if b.failures >= b.threshold {
		b.openUntil = time.Now().Add(b.openFor)
		b.failures = 0
	}
}
