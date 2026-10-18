// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

package grpc

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils/correlation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

// NewServer builds a gRPC server with request ID and panic recovery. Optional extra grpc.ServerOption values append after defaults.
func NewServer(log logger.Logger, extra ...grpc.ServerOption) *grpc.Server {
	opts := make([]grpc.ServerOption, 0, len(extra)+3)
	opts = append(opts, extra...)
	opts = append(opts,
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    30 * time.Second,
			Timeout: 5 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             10 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	opts = append(opts, grpc.ChainUnaryInterceptor(
		UnaryServerRequestID(),
		recoveryUnary(log),
	))

	return grpc.NewServer(opts...)
}

func recoveryUnary(log logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				if log != nil {
					log.Error("grpc panic recovered",
						"method", info.FullMethod,
						"request_id", correlation.RequestIDFromContext(ctx),
						"panic", r,
						"stack", string(debug.Stack()),
					)
				}
				err = status.Error(codes.Internal, "internal error")
			}
		}()

		resp, err = handler(ctx, req)
		if err != nil && log != nil {
			if st, ok := status.FromError(err); ok && st.Code() == codes.Internal {
				log.Error("grpc internal error",
					"method", info.FullMethod,
					"request_id", correlation.RequestIDFromContext(ctx),
					"err", err,
				)
			}
		}

		return resp, err
	}
}
