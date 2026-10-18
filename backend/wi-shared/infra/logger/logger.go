// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

package logger

import (
	"log/slog"
	"os"
	"strings"
)

// Logger is the application logging interface (inject into service, repository, delivery).
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	With(args ...any) Logger
}

type slogLogger struct {
	log *slog.Logger
}

// New builds a stdout logger with service/env labels.
// Production uses JSON; other envs use text for local readability.
func New(opts Options) Logger {
	level := ParseLevel(opts.Level)
	handlerOpts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if stringsEqualFold(opts.Env, "production") {
		handler = slog.NewJSONHandler(os.Stdout, handlerOpts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, handlerOpts)
	}

	base := slog.New(handler).With(
		slog.String("service", opts.Service),
		slog.String("env", opts.Env),
	)

	return &slogLogger{log: base}
}

func (l *slogLogger) Debug(msg string, args ...any) { l.log.Debug(msg, args...) }
func (l *slogLogger) Info(msg string, args ...any)  { l.log.Info(msg, args...) }
func (l *slogLogger) Warn(msg string, args ...any)  { l.log.Warn(msg, args...) }
func (l *slogLogger) Error(msg string, args ...any) { l.log.Error(msg, args...) }

func (l *slogLogger) With(args ...any) Logger {
	return &slogLogger{log: l.log.With(args...)}
}

func stringsEqualFold(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}

// HTTPWriteErrorLogger adapts Logger for response/http.WithLogger.
func HTTPWriteErrorLogger(l Logger) func(error) {
	return func(err error) {
		if err == nil {
			return
		}
		l.Error("http error response", "err", err)
	}
}
