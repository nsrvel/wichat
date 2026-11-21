package grpc

import (
	"context"
	"runtime/debug"
	"strings"
	"time"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils/correlation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

// NewServer builds a gRPC server with request ID and panic recovery. Optional extra grpc.ServerOption values append after defaults.
// internalToken when non-empty requires matching x-wichat-internal-token on RPCs except General/Health.
func NewServer(log logger.Logger, internalToken string, extra ...grpc.ServerOption) *grpc.Server {
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
	interceptors := []grpc.UnaryServerInterceptor{UnaryServerRequestID(), recoveryUnary(log)}
	if strings.TrimSpace(internalToken) != "" {
		interceptors = append([]grpc.UnaryServerInterceptor{UnaryServerInternalToken(internalToken)}, interceptors...)
	}
	opts = append(opts, grpc.ChainUnaryInterceptor(interceptors...))

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
