// Package correlation holds cross-transport request ID helpers (HTTP + gRPC).
package correlation

import "context"

// RequestIDHeader is the HTTP header used for correlation across api-gateway and gRPC metadata.
const RequestIDHeader = "X-Request-ID"

type requestIDKey struct{}

// WithRequestID returns a child context carrying the request ID.
func WithRequestID(ctx context.Context, id string) context.Context {
	if id == "" {
		return ctx
	}

	return context.WithValue(ctx, requestIDKey{}, id)
}

// RequestIDFromContext returns the request ID when present.
func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	id, _ := ctx.Value(requestIDKey{}).(string)

	return id
}
