// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

package exception_test

import (
	"net/http"
	"testing"

	"github.com/wichat/wichat/backend/wi-shared/exception"
	"google.golang.org/grpc/codes"
)

func TestKindHTTPAndGRPCMapping(t *testing.T) {
	cases := []struct {
		kind       exception.Kind
		httpStatus int
		grpcCode   codes.Code
	}{
		{exception.KindInvalidInput, http.StatusBadRequest, codes.InvalidArgument},
		{exception.KindUnauthenticated, http.StatusUnauthorized, codes.Unauthenticated},
		{exception.KindPermissionDenied, http.StatusForbidden, codes.PermissionDenied},
		{exception.KindNotFound, http.StatusNotFound, codes.NotFound},
		{exception.KindConflict, http.StatusConflict, codes.AlreadyExists},
		{exception.KindStateConflict, http.StatusConflict, codes.Aborted},
		{exception.KindRateLimited, http.StatusTooManyRequests, codes.ResourceExhausted},
		{exception.KindUnavailable, http.StatusServiceUnavailable, codes.Unavailable},
		{exception.KindDeadlineExceeded, http.StatusGatewayTimeout, codes.DeadlineExceeded},
		{exception.KindInternal, http.StatusInternalServerError, codes.Internal},
	}
	for _, tc := range cases {
		if got := exception.HTTPStatus(tc.kind); got != tc.httpStatus {
			t.Errorf("HTTPStatus(%v) = %d, want %d", tc.kind, got, tc.httpStatus)
		}
		if got := exception.GRPCCode(tc.kind); got != tc.grpcCode {
			t.Errorf("GRPCCode(%v) = %v, want %v", tc.kind, got, tc.grpcCode)
		}
		if got := exception.KindFromGRPCCode(tc.grpcCode); got != tc.kind {
			t.Errorf("KindFromGRPCCode(%v) = %v, want %v", tc.grpcCode, got, tc.kind)
		}
	}
}
