// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

package http

import (
	"encoding/json"
	"net/http"

	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	"github.com/wichat/wichat/backend/wi-shared/response"
	responsegrpc "github.com/wichat/wichat/backend/wi-shared/response/grpc"
)

// Logger logs errors that become internal responses (optional observability hook).
type Logger func(err error)

type writeOptions struct {
	log Logger
}

// WriteOption configures WriteError.
type WriteOption func(*writeOptions)

// WithLogger attaches a logger for non-exception and internal errors.
func WithLogger(log Logger) WriteOption {
	return func(o *writeOptions) {
		o.log = log
	}
}

// WriteJSON writes a JSON response with the given status.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// WriteError maps err to an HTTP JSON error envelope.
func WriteError(w http.ResponseWriter, _ *http.Request, err error, opts ...WriteOption) {
	if err == nil {
		return
	}

	var o writeOptions
	for _, opt := range opts {
		opt(&o)
	}

	ex, ok := exception.As(err)
	if ok {
		writeException(w, ex, o)
		return
	}

	if o.log != nil {
		o.log(err)
	}
	WriteJSON(w, http.StatusInternalServerError, response.NewErrorEnvelope(string(exception.MsgInternal), nil))
}

func writeException(w http.ResponseWriter, ex *exception.Error, o writeOptions) {
	if ex.Kind == exception.KindInternal && o.log != nil {
		if ex.Cause != nil {
			o.log(ex.Cause)
		} else {
			o.log(ex)
		}
	}

	status := exception.HTTPStatus(ex.Kind)
	code := string(clientCode(ex))

	var params map[string]any
	if ex.Kind != exception.KindInternal && len(ex.Params) > 0 {
		params = ex.Params
	}

	WriteJSON(w, status, response.NewErrorEnvelope(code, params))
}

func clientCode(ex *exception.Error) exception.MessageCode {
	if ex.Kind == exception.KindInternal {
		return exception.MsgInternal
	}

	return ex.Code
}

// WriteOutboundError maps downstream gRPC or transport errors to HTTP JSON.
func WriteOutboundError(w http.ResponseWriter, r *http.Request, err error, opts ...WriteOption) {
	if err == nil {
		return
	}

	if _, ok := exception.As(err); ok {
		WriteError(w, r, err, opts...)
		return
	}

	if _, ok := responsegrpc.FromError(err); ok {
		WriteErrorFromGRPC(w, r, err, opts...)
		return
	}

	if utils.IsOutboundUnavailable(err) {
		WriteJSON(w, http.StatusServiceUnavailable, response.NewErrorEnvelope(string(exception.MsgUnavailable), nil))
		return
	}

	WriteError(w, r, err, opts...)
}

// WriteErrorFromGRPC maps a downstream gRPC error to an HTTP JSON error response.
func WriteErrorFromGRPC(w http.ResponseWriter, _ *http.Request, grpcErr error, opts ...WriteOption) {
	if grpcErr == nil {
		return
	}

	var o writeOptions
	for _, opt := range opts {
		opt(&o)
	}

	parsed, ok := responsegrpc.FromError(grpcErr)
	if !ok {
		if o.log != nil {
			o.log(grpcErr)
		}
		WriteJSON(w, http.StatusInternalServerError, response.NewErrorEnvelope(string(exception.MsgInternal), nil))
		return
	}

	if parsed.Kind == exception.KindInternal && o.log != nil {
		o.log(grpcErr)
	}

	status := exception.HTTPStatus(parsed.Kind)
	code := string(parsed.Code)

	var params map[string]any
	if parsed.Kind != exception.KindInternal {
		params = parsed.Params
	}

	WriteJSON(w, status, response.NewErrorEnvelope(code, params))
}
