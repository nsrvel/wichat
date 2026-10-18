// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

package logger_test

import (
	"log/slog"
	"testing"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

func TestParseLevel(t *testing.T) {
	if logger.ParseLevel("debug") != slog.LevelDebug {
		t.Fatal("debug")
	}
	if logger.ParseLevel("unknown") != slog.LevelInfo {
		t.Fatal("default info")
	}
}

func TestNewProductionJSON(t *testing.T) {
	l := logger.New(logger.Options{Service: "test", Env: "production", Level: "info"})
	l.Info("hello", "key", "value")
}

func TestHTTPWriteErrorLogger(t *testing.T) {
	l := logger.New(logger.Options{Service: "test", Env: "development", Level: "error"})
	fn := logger.HTTPWriteErrorLogger(l)
	fn(nil)
	fn(testingErr())
}

type testErr struct{}

func (testErr) Error() string { return "boom" }

func testingErr() error { return testErr{} }
