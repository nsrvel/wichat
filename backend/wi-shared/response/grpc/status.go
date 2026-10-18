// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

package grpc

import (
	"encoding/json"

	"github.com/wichat/wichat/backend/wi-shared/exception"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

const metadataParamsKey = "params"

// ToStatus converts err into a gRPC status. Raw errors become Internal with error.internal.
func ToStatus(err error) error {
	if err == nil {
		return nil
	}
	ex, ok := exception.As(err)
	if !ok {
		return status.New(exception.GRPCCode(exception.KindInternal), string(exception.MsgInternal)).Err()
	}

	code := exception.GRPCCode(ex.Kind)
	msg := string(clientCode(ex))
	st := status.New(code, msg)

	if len(ex.Params) > 0 && ex.Kind != exception.KindInternal {
		paramsJSON, marshalErr := json.Marshal(ex.Params)
		if marshalErr != nil {
			return status.New(exception.GRPCCode(exception.KindInternal), string(exception.MsgInternal)).Err()
		}
		withDetails, detailErr := st.WithDetails(&errdetails.ErrorInfo{
			Reason:   msg,
			Metadata: map[string]string{metadataParamsKey: string(paramsJSON)},
		})
		if detailErr != nil {
			return status.New(exception.GRPCCode(exception.KindInternal), string(exception.MsgInternal)).Err()
		}
		return withDetails.Err()
	}
	return st.Err()
}

// ParsedStatus holds fields extracted from a gRPC status for HTTP translation.
type ParsedStatus struct {
	Kind   exception.Kind
	Code   exception.MessageCode
	Params map[string]any
}

// ParseStatus reads i18n code and params from st (message = code, ErrorInfo metadata).
func ParseStatus(st *status.Status) (ParsedStatus, bool) {
	if st == nil {
		return ParsedStatus{}, false
	}
	kind := exception.KindFromGRPCCode(st.Code())
	code := exception.MessageCode(st.Message())
	if code == "" {
		switch kind {
		case exception.KindInvalidInput:
			code = exception.MsgInvalidInput
		case exception.KindUnavailable:
			code = exception.MsgUnavailable
		case exception.KindDeadlineExceeded:
			code = exception.MsgTimeout
		default:
			code = exception.MsgInternal
		}
	}
	if !exception.ValidMessageCode(code) {
		kind = exception.KindInternal
		code = exception.MsgInternal
	}

	var params map[string]any
	for _, d := range st.Details() {
		info, ok := d.(*errdetails.ErrorInfo)
		if !ok || info.Metadata == nil {
			continue
		}
		raw, ok := info.Metadata[metadataParamsKey]
		if !ok || raw == "" {
			continue
		}
		if err := json.Unmarshal([]byte(raw), &params); err != nil {
			params = nil
		}
		break
	}

	if kind == exception.KindInternal {
		params = nil
		code = exception.MsgInternal
	}

	return ParsedStatus{Kind: kind, Code: code, Params: params}, true
}

// FromError parses a gRPC error returned from status.FromError.
func FromError(err error) (ParsedStatus, bool) {
	if err == nil {
		return ParsedStatus{}, false
	}
	st, ok := status.FromError(err)
	if !ok {
		return ParsedStatus{}, false
	}
	return ParseStatus(st)
}

func clientCode(ex *exception.Error) exception.MessageCode {
	if ex.Kind == exception.KindInternal {
		return exception.MsgInternal
	}
	return ex.Code
}
