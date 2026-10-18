// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

package exception

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

// Kind classifies errors for HTTP and gRPC mapping.
type Kind int

const (
	KindInvalidInput Kind = iota
	KindUnauthenticated
	KindPermissionDenied
	KindNotFound
	KindConflict
	KindStateConflict
	KindRateLimited
	KindUnavailable
	KindDeadlineExceeded
	KindInternal
)

// HTTPStatus returns the HTTP status for k.
func HTTPStatus(k Kind) int {
	switch k {
	case KindInvalidInput:
		return http.StatusBadRequest
	case KindUnauthenticated:
		return http.StatusUnauthorized
	case KindPermissionDenied:
		return http.StatusForbidden
	case KindNotFound:
		return http.StatusNotFound
	case KindConflict, KindStateConflict:
		return http.StatusConflict
	case KindRateLimited:
		return http.StatusTooManyRequests
	case KindUnavailable:
		return http.StatusServiceUnavailable
	case KindDeadlineExceeded:
		return http.StatusGatewayTimeout
	case KindInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// GRPCCode returns the gRPC status code for k.
func GRPCCode(k Kind) codes.Code {
	switch k {
	case KindInvalidInput:
		return codes.InvalidArgument
	case KindUnauthenticated:
		return codes.Unauthenticated
	case KindPermissionDenied:
		return codes.PermissionDenied
	case KindNotFound:
		return codes.NotFound
	case KindConflict:
		return codes.AlreadyExists
	case KindStateConflict:
		return codes.Aborted
	case KindRateLimited:
		return codes.ResourceExhausted
	case KindUnavailable:
		return codes.Unavailable
	case KindDeadlineExceeded:
		return codes.DeadlineExceeded
	case KindInternal:
		return codes.Internal
	default:
		return codes.Internal
	}
}

// KindFromGRPCCode maps a gRPC code to Kind for api-gateway HTTP translation.
func KindFromGRPCCode(c codes.Code) Kind {
	switch c {
	case codes.InvalidArgument:
		return KindInvalidInput
	case codes.Unauthenticated:
		return KindUnauthenticated
	case codes.PermissionDenied:
		return KindPermissionDenied
	case codes.NotFound:
		return KindNotFound
	case codes.AlreadyExists:
		return KindConflict
	case codes.Aborted:
		return KindStateConflict
	case codes.ResourceExhausted:
		return KindRateLimited
	case codes.Unavailable:
		return KindUnavailable
	case codes.DeadlineExceeded:
		return KindDeadlineExceeded
	default:
		return KindInternal
	}
}
