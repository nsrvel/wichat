package middleware

import (
	"net/http"
	"net/url"
)

func corsOriginAllowed(configured, request string) bool {
	if configured == "" || request == "" {
		return false
	}
	if configured == request {
		return true
	}
	co, err1 := url.Parse(configured)
	ro, err2 := url.Parse(request)
	if err1 != nil || err2 != nil {
		return false
	}
	if co.Scheme != ro.Scheme || co.Port() != ro.Port() {
		return false
	}
	devLoopback := func(host string) bool {
		return host == "localhost" || host == "127.0.0.1"
	}

	return devLoopback(co.Hostname()) && devLoopback(ro.Hostname())
}

// CORS allows credentialed browser calls from one configured origin.
func CORS(origin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqOrigin := r.Header.Get("Origin")
		if corsOriginAllowed(origin, reqOrigin) {
			w.Header().Set("Access-Control-Allow-Origin", reqOrigin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)

			return
		}

		next.ServeHTTP(w, r)
	})
}
