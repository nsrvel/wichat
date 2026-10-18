// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

package exception_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/wichat/wichat/backend/wi-shared/exception"
)

func TestConstructorsSetKindAndCode(t *testing.T) {
	err := exception.NotFound(exception.MsgUserNotFound, map[string]any{"user_id": "x"})
	ex, ok := exception.As(err)
	if !ok {
		t.Fatal("As failed")
	}
	if ex.Kind != exception.KindNotFound || ex.Code != exception.MsgUserNotFound {
		t.Fatalf("got kind=%v code=%s", ex.Kind, ex.Code)
	}
	if ex.Params["user_id"] != "x" {
		t.Fatalf("params: %v", ex.Params)
	}
}

func TestAsThroughWrap(t *testing.T) {
	inner := exception.InvalidInput(exception.MsgInvalidInput, nil)
	wrapped := fmt.Errorf("ctx: %w", inner)
	ex, ok := exception.As(wrapped)
	if !ok || ex.Kind != exception.KindInvalidInput {
		t.Fatalf("As through wrap: ok=%v kind=%v", ok, ex.Kind)
	}
}

func TestParamsCopiedFromCallerMap(t *testing.T) {
	params := map[string]any{"user_id": "x"}
	err := exception.NotFound(exception.MsgUserNotFound, params)
	params["user_id"] = "mutated"
	ex, ok := exception.As(err)
	if !ok {
		t.Fatal("As failed")
	}
	if ex.Params["user_id"] != "x" {
		t.Fatalf("params mutated: %v", ex.Params)
	}
}

func TestValidMessageCode(t *testing.T) {
	if !exception.ValidMessageCode(exception.MsgUserNotFound) {
		t.Fatal("user code")
	}
	if exception.ValidMessageCode(exception.MessageCode("raw db error")) {
		t.Fatal("unexpected valid")
	}
}

func TestInternalFixedCodeNoParams(t *testing.T) {
	cause := errors.New("pq: connection refused")
	err := exception.Internal(cause)
	ex, ok := exception.As(err)
	if !ok {
		t.Fatal("As failed")
	}
	if ex.Code != exception.MsgInternal || ex.Params != nil {
		t.Fatalf("code=%s params=%v", ex.Code, ex.Params)
	}
	if !errors.Is(err, cause) {
		t.Fatal("expected Cause in chain")
	}
	if ex.Error() == cause.Error() {
		t.Fatal("Error() should not expose raw cause to string consumers")
	}
}
