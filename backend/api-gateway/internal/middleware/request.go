// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

package middleware

import (
	"net"
	"net/http"
)

const maxUserAgentLen = 512

// ClientIP returns the remote host without a port.
func ClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}

// UserAgent returns the request user agent, capped for storage.
func UserAgent(r *http.Request) string {
	ua := r.UserAgent()
	if len(ua) <= maxUserAgentLen {
		return ua
	}

	return ua[:maxUserAgentLen]
}
