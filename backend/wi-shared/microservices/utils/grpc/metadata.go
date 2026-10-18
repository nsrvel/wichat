// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

package grpc

import (
	"context"

	"github.com/wichat/wichat/backend/wi-shared/microservices/utils/correlation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const metadataRequestIDKey = "x-request-id"

// UnaryClientRequestID propagates the request ID from context into outgoing gRPC metadata.
func UnaryClientRequestID() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		id := correlation.RequestIDFromContext(ctx)
		if id != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, metadataRequestIDKey, id)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// UnaryServerRequestID reads x-request-id from incoming metadata into context.
func UnaryServerRequestID() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if vals := md.Get(metadataRequestIDKey); len(vals) > 0 && vals[0] != "" {
				ctx = correlation.WithRequestID(ctx, vals[0])
			}
		}

		return handler(ctx, req)
	}
}
