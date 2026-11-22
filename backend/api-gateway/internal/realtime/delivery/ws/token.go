package ws

import (
	"net/http"
	"strings"
)

// BearerTokenFromRequest resolves an access token for the realtime upgrade (Sec-WebSocket-Protocol bearer.* or Authorization).
func BearerTokenFromRequest(r *http.Request) string {
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

// AllowsDeferredBearerAuth is true when the client offers the short "bearer" subprotocol (JWT in heartbeat body).
func AllowsDeferredBearerAuth(r *http.Request) bool {
	for _, part := range strings.Split(r.Header.Get("Sec-WebSocket-Protocol"), ",") {
		if strings.EqualFold(strings.TrimSpace(part), "bearer") {
			return true
		}
	}
	return false
}

// NegotiatedSubprotocols lists server-supported subprotocols that match the client offer (incl. legacy bearer.<jwt>).
func NegotiatedSubprotocols(r *http.Request) []string {
	out := []string{"bearer"}
	for _, part := range strings.Split(r.Header.Get("Sec-WebSocket-Protocol"), ",") {
		part = strings.TrimSpace(part)
		if len(part) > 7 && strings.EqualFold(part[:7], "bearer.") {
			out = append(out, part)
		}
	}
	return out
}
