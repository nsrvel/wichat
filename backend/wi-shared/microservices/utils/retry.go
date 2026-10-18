package utils

import (
	"context"
	"errors"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func shouldRetry(policy RetryPolicy, err error, attempt, maxRetries int) bool {
	if policy == RetryNone || err == nil {
		return false
	}
	if attempt >= maxRetries {
		return false
	}
	if isClientGRPCCode(err) {
		return false
	}
	if isRetriableGRPCCode(err) || isTransportError(err) {
		return true
	}

	return false
}

func retryDelay(attempt int, backoff time.Duration, err error) time.Duration {
	if isTransportError(err) {
		return 0
	}

	return time.Duration(attempt+1) * backoff
}

func isClientGRPCCode(err error) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	switch st.Code() {
	case codes.InvalidArgument,
		codes.NotFound,
		codes.AlreadyExists,
		codes.PermissionDenied,
		codes.Unauthenticated,
		codes.FailedPrecondition,
		codes.OutOfRange,
		codes.Unimplemented:
		return true
	default:
		return false
	}
}

func isRetriableGRPCCode(err error) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	switch st.Code() {
	case codes.Unavailable, codes.ResourceExhausted, codes.Aborted, codes.DeadlineExceeded:
		if errors.Is(err, context.Canceled) {
			return false
		}

		return true
	default:
		return false
	}
}

func isTransportError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	st, ok := status.FromError(err)
	if ok && st.Code() == codes.Unavailable {
		return true
	}
	msg := strings.ToLower(err.Error())
	transportHints := []string{
		"connection refused",
		"connection reset",
		"connection closed",
		"eof",
		"broken pipe",
		"transport is closing",
	}
	for _, hint := range transportHints {
		if strings.Contains(msg, hint) {
			return true
		}
	}

	return false
}
