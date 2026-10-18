// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils/correlation"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

// Chain wraps the handler with api-gateway middleware (order: request ID, recover).
func Chain(next http.Handler) http.Handler {
	return requestID(recoverPanic(next))
}

// Request ID
func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(correlation.RequestIDHeader)
		if id == "" {
			id = newRequestID()
		}

		w.Header().Set(correlation.RequestIDHeader, id)
		r = r.WithContext(correlation.WithRequestID(r.Context(), id))
		next.ServeHTTP(w, r)
	})
}

func newRequestID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])

	return hex.EncodeToString(b[:])
}

// Recover panic
func recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recover() != nil {
				responsehttp.WriteError(w, r, exception.Internal(nil))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
