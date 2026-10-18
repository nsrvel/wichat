// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/wichat/wichat/backend/wi-shared/exception"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

// Limiter is a fixed-window counter kept in this process.
// ponytail: one gateway process only; move the buckets to Redis when the gateway is scaled out.
type Limiter struct {
	limit  int
	window time.Duration
	mu     sync.Mutex
	hits   map[string][]time.Time
	now    func() time.Time
}

// NewLimiter allows limit events per key inside window.
func NewLimiter(limit int, window time.Duration) *Limiter {
	return &Limiter{
		limit:  limit,
		window: window,
		hits:   map[string][]time.Time{},
		now:    time.Now,
	}
}

// Allow reports whether key may proceed and records the hit when it may.
func (l *Limiter) Allow(key string) bool {
	if l == nil || l.limit <= 0 {
		return true
	}

	now := l.now()
	cutoff := now.Add(-l.window)

	l.mu.Lock()
	defer l.mu.Unlock()

	l.prune(cutoff)

	kept := l.hits[key][:0]
	for _, ts := range l.hits[key] {
		if ts.After(cutoff) {
			kept = append(kept, ts)
		}
	}
	if len(kept) >= l.limit {
		if len(kept) == 0 {
			delete(l.hits, key)
		} else {
			l.hits[key] = kept
		}

		return false
	}

	l.hits[key] = append(kept, now)

	return true
}

func (l *Limiter) prune(cutoff time.Time) {
	for key, series := range l.hits {
		fresh := false
		for _, ts := range series {
			if ts.After(cutoff) {
				fresh = true
				break
			}
		}
		if !fresh {
			delete(l.hits, key)
		}
	}
}

// Wrap rejects requests over the limit with error.rate_limited.
func (l *Limiter) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.Allow(ClientIP(r)) {
			responsehttp.WriteError(w, r, exception.RateLimited(exception.MsgRateLimited, nil))

			return
		}

		next.ServeHTTP(w, r)
	})
}
