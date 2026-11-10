package logger_test

import (
	"testing"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

func TestNewProductionJSON(t *testing.T) {
	l := logger.New(logger.Options{Service: "test", Env: "production", Level: "info"})
	l.Info("hello", "key", "value")
	l.Debug("debug", "k", "v")
	l.Warn("warn", "k", "v")
	child := l.With("trace", "1")
	child.Error("child", "k", "v")
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
