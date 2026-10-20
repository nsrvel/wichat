package middleware

import (
	"net"
	"net/http"

	"github.com/wichat/wichat/backend/api-gateway/pkg/constants"
)

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
	if len(ua) <= constants.MaxUserAgentLen {
		return ua
	}

	return ua[:constants.MaxUserAgentLen]
}
