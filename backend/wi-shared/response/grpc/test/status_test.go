// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

package grpc_test

import (
	"errors"
	"testing"

	"github.com/wichat/wichat/backend/wi-shared/exception"
	responsegrpc "github.com/wichat/wichat/backend/wi-shared/response/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestToStatusRoundTripWithParams(t *testing.T) {
	err := exception.NotFound(exception.MsgUserNotFound, map[string]any{"user_id": "abc"})
	stErr := responsegrpc.ToStatus(err)
	st, ok := status.FromError(stErr)
	if !ok {
		t.Fatal("not a status error")
	}
	if st.Code() != codes.NotFound {
		t.Fatalf("code %v", st.Code())
	}
	if st.Message() != string(exception.MsgUserNotFound) {
		t.Fatalf("message %q", st.Message())
	}

	parsed, ok := responsegrpc.ParseStatus(st)
	if !ok {
		t.Fatal("ParseStatus failed")
	}
	if parsed.Kind != exception.KindNotFound || parsed.Code != exception.MsgUserNotFound {
		t.Fatalf("parsed kind=%v code=%s", parsed.Kind, parsed.Code)
	}
	if parsed.Params["user_id"] != "abc" {
		t.Fatalf("params %v", parsed.Params)
	}
}

func TestToStatusRawErrorInternal(t *testing.T) {
	stErr := responsegrpc.ToStatus(errors.New("pq: secret"))
	st, ok := status.FromError(stErr)
	if !ok {
		t.Fatal("not a status error")
	}
	if st.Code() != codes.Internal {
		t.Fatalf("code %v", st.Code())
	}
	if st.Message() != string(exception.MsgInternal) {
		t.Fatalf("message leaked: %q", st.Message())
	}
}

func TestToStatusNil(t *testing.T) {
	if responsegrpc.ToStatus(nil) != nil {
		t.Fatal("expected nil")
	}
}

func TestParseStatusRejectsInvalidMessageCode(t *testing.T) {
	st := status.New(codes.NotFound, "raw db error")
	parsed, ok := responsegrpc.ParseStatus(st)
	if !ok {
		t.Fatal("ParseStatus failed")
	}
	if parsed.Kind != exception.KindInternal || parsed.Code != exception.MsgInternal {
		t.Fatalf("kind=%v code=%s", parsed.Kind, parsed.Code)
	}
}

func TestUnavailableRoundTrip(t *testing.T) {
	stErr := responsegrpc.ToStatus(exception.Unavailable(exception.MsgUnavailable, nil))
	st, ok := status.FromError(stErr)
	if !ok || st.Code() != codes.Unavailable {
		t.Fatalf("code %v", st.Code())
	}
	parsed, ok := responsegrpc.ParseStatus(st)
	if !ok || parsed.Code != exception.MsgUnavailable {
		t.Fatalf("code %s", parsed.Code)
	}
}
