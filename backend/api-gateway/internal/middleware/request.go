package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/wichat/wichat/backend/api-gateway/pkg/constants"
)

var (
	trustedMu   sync.RWMutex
	trustedNets []*net.IPNet
)

// ConfigureTrustedProxies parses a comma-separated list of IPs or CIDRs whose forwarding headers are trusted.
func ConfigureTrustedProxies(raw string) error {
	var parsed []*net.IPNet
	for _, entry := range strings.Split(raw, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if !strings.Contains(entry, "/") {
			ip := net.ParseIP(entry)
			if ip == nil {
				return &net.ParseError{Type: "CIDR/IP", Text: entry}
			}
			bits := 32
			if ip.To4() == nil {
				bits = 128
			}
			parsed = append(parsed, &net.IPNet{IP: ip, Mask: net.CIDRMask(bits, bits)})
			continue
		}
		_, cidr, err := net.ParseCIDR(entry)
		if err != nil {
			return err
		}
		parsed = append(parsed, cidr)
	}

	trustedMu.Lock()
	trustedNets = parsed
	trustedMu.Unlock()
	return nil
}

func isTrustedProxy(ip net.IP) bool {
	if ip == nil {
		return false
	}
	trustedMu.RLock()
	nets := trustedNets
	trustedMu.RUnlock()

	for _, n := range nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

func remoteHost(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return strings.TrimSpace(remoteAddr)
	}
	return strings.TrimSpace(host)
}

// ClientIP returns the client IP, honoring forwarding headers only from configured trusted proxies.
func ClientIP(r *http.Request) string {
	host := remoteHost(r.RemoteAddr)
	remoteIP := net.ParseIP(host)
	if !isTrustedProxy(remoteIP) {
		return host
	}

	if cf := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); cf != "" {
		if ip := net.ParseIP(cf); ip != nil {
			return ip.String()
		}
	}

	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		for i := len(parts) - 1; i >= 0; i-- {
			candidate := strings.TrimSpace(parts[i])
			ip := net.ParseIP(candidate)
			if ip == nil {
				continue
			}
			if !isTrustedProxy(ip) {
				return ip.String()
			}
		}
	}

	if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
		if ip := net.ParseIP(xri); ip != nil {
			return ip.String()
		}
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
