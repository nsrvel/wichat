package utils

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// IsClientCanceled reports whether err is a client-aborted request (no response body).
func IsClientCanceled(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return true
	}
	st, ok := status.FromError(err)
	return ok && st.Code() == codes.Canceled
}

// IsOutboundUnavailable reports whether err should map to HTTP 503 at the gateway.
func IsOutboundUnavailable(err error) bool {
	if err == nil {
		return false
	}
	if isTransportError(err) {
		return true
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	switch st.Code() {
	case codes.Unavailable, codes.DeadlineExceeded:
		return true
	default:
		return false
	}
}
