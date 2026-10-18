// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

package exception

import (
	"errors"
)

// Error is a semantic application error with an i18n code. Cause is for logging only.
// Treat returned *Error as read-only (do not mutate fields after construction).
type Error struct {
	Kind   Kind
	Code   MessageCode
	Params map[string]any
	Cause  error
}

func (e *Error) Error() string {
	return string(e.Code)
}

func (e *Error) Unwrap() error {
	return e.Cause
}

func cloneParams(params map[string]any) map[string]any {
	if params == nil {
		return nil
	}

	out := make(map[string]any, len(params))
	for k, v := range params {
		out[k] = v
	}

	return out
}

func newErr(kind Kind, code MessageCode, params map[string]any) *Error {
	return &Error{Kind: kind, Code: code, Params: cloneParams(params)}
}

func InvalidInput(code MessageCode, params map[string]any) error {
	return newErr(KindInvalidInput, code, params)
}

func Unauthenticated(code MessageCode, params map[string]any) error {
	return newErr(KindUnauthenticated, code, params)
}

func PermissionDenied(code MessageCode, params map[string]any) error {
	return newErr(KindPermissionDenied, code, params)
}

func NotFound(code MessageCode, params map[string]any) error {
	return newErr(KindNotFound, code, params)
}

func Conflict(code MessageCode, params map[string]any) error {
	return newErr(KindConflict, code, params)
}

func StateConflict(code MessageCode, params map[string]any) error {
	return newErr(KindStateConflict, code, params)
}

func RateLimited(code MessageCode, params map[string]any) error {
	return newErr(KindRateLimited, code, params)
}

func Unavailable(code MessageCode, params map[string]any) error {
	return newErr(KindUnavailable, code, params)
}

func DeadlineExceeded(code MessageCode, params map[string]any) error {
	return newErr(KindDeadlineExceeded, code, params)
}

// Internal wraps cause for logs; clients always receive MsgInternal with no params.
func Internal(cause error) error {
	return &Error{Kind: KindInternal, Code: MsgInternal, Cause: cause}
}

// As returns *Error if err is or wraps an exception.Error.
func As(err error) (*Error, bool) {
	var ex *Error
	if errors.As(err, &ex) {
		return ex, true
	}

	return nil, false
}
