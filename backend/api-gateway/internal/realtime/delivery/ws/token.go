package ws

import (
	"net/http"
	"strings"
)

// BearerTokenFromRequest resolves an access token for the realtime upgrade (query deprecated; prefer subprotocol).
func BearerTokenFromRequest(r *http.Request) string {
	if t := strings.TrimSpace(r.URL.Query().Get("access_token")); t != "" {
		return t
	}
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	const prefix = "Bearer "
	if strings.HasPrefix(authHeader, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
	}
	for _, part := range strings.Split(r.Header.Get("Sec-WebSocket-Protocol"), ",") {
		part = strings.TrimSpace(part)
		if len(part) > 7 && strings.EqualFold(part[:7], "bearer.") {
			return part[7:]
		}
	}
	return ""
}
