package http_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wichat/wichat/backend/wi-shared/exception"
	responsegrpc "github.com/wichat/wichat/backend/wi-shared/response/grpc"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

func TestWriteErrorNil(t *testing.T) {
	rec := httptest.NewRecorder()
	responsehttp.WriteError(rec, nil, nil)
	if rec.Code != 200 || rec.Body.Len() != 0 {
		t.Fatalf("code=%d body=%q", rec.Code, rec.Body.String())
	}
}

func TestWriteOutboundErrorNil(t *testing.T) {
	rec := httptest.NewRecorder()
	responsehttp.WriteOutboundError(rec, nil, nil)
	if rec.Code != 200 || rec.Body.Len() != 0 {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestWriteErrorExceptionWithParams(t *testing.T) {
	rec := httptest.NewRecorder()
	err := exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "email"})
	responsehttp.WriteError(rec, nil, err)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rec.Code)
	}
	var body map[string]map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["error"]["code"] != string(exception.MsgInvalidInput) {
		t.Fatalf("code %v", body["error"]["code"])
	}
}

func TestWriteErrorRawDoesNotLeak(t *testing.T) {
	rec := httptest.NewRecorder()
	responsehttp.WriteError(rec, nil, errors.New("pq: secret"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "pq") {
		t.Fatalf("body leaked driver: %s", rec.Body.String())
	}
	var body map[string]map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["error"]["code"] != string(exception.MsgInternal) {
		t.Fatalf("code %v", body["error"]["code"])
	}
}

func TestWriteErrorInternalLogsCause(t *testing.T) {
	var logged error
	logFn := func(err error) { logged = err }
	cause := errors.New("db timeout")

	rec := httptest.NewRecorder()
	responsehttp.WriteError(rec, nil, exception.Internal(cause), responsehttp.WithLogger(logFn))

	if logged == nil || logged.Error() != "db timeout" {
		t.Fatalf("logged: %v", logged)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestWriteError4xxDoesNotLog(t *testing.T) {
	var logged bool
	logFn := func(err error) { logged = true }

	rec := httptest.NewRecorder()
	responsehttp.WriteError(rec, nil, exception.NotFound(exception.MsgUserNotFound, nil), responsehttp.WithLogger(logFn))

	if logged {
		t.Fatal("expected no log for 4xx")
	}
}

func TestWriteOutboundErrorTransport(t *testing.T) {
	rec := httptest.NewRecorder()
	responsehttp.WriteOutboundError(rec, nil, errors.New("connection refused"))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status %d", rec.Code)
	}
	var body map[string]map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["error"]["code"] != string(exception.MsgUnavailable) {
		t.Fatalf("code %v", body["error"]["code"])
	}
}

func TestWriteOutboundErrorExceptionPassthrough(t *testing.T) {
	rec := httptest.NewRecorder()
	err := exception.Unauthenticated(exception.MsgInvalidCredentials, nil)
	responsehttp.WriteOutboundError(rec, nil, err)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestWriteOutboundErrorGRPCStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	grpcErr := responsegrpc.ToStatus(exception.PermissionDenied(exception.MsgForbidden, nil))
	responsehttp.WriteOutboundError(rec, nil, grpcErr)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestWriteErrorFromGRPCRaw(t *testing.T) {
	var logged bool
	rec := httptest.NewRecorder()
	responsehttp.WriteErrorFromGRPC(rec, nil, errors.New("not a status"), responsehttp.WithLogger(func(error) { logged = true }))

	if !logged || rec.Code != http.StatusInternalServerError {
		t.Fatalf("logged=%v status=%d", logged, rec.Code)
	}
}

func TestWriteErrorFromGRPCInternalLogs(t *testing.T) {
	var logged bool
	rec := httptest.NewRecorder()
	grpcErr := responsegrpc.ToStatus(exception.Internal(errors.New("cause")))
	responsehttp.WriteErrorFromGRPC(rec, nil, grpcErr, responsehttp.WithLogger(func(error) { logged = true }))

	if !logged || rec.Code != http.StatusInternalServerError {
		t.Fatalf("logged=%v status=%d", logged, rec.Code)
	}
}

func TestWriteErrorFromGRPCNotFound(t *testing.T) {
	grpcErr := responsegrpc.ToStatus(exception.NotFound(exception.MsgUserNotFound, map[string]any{"user_id": "1"}))
	rec := httptest.NewRecorder()
	responsehttp.WriteErrorFromGRPC(rec, nil, grpcErr)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d", rec.Code)
	}
	var body map[string]map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["error"]["code"] != string(exception.MsgUserNotFound) {
		t.Fatalf("code %v", body["error"]["code"])
	}
}
