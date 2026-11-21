package grpc

import (
	"context"
	"strings"

	"github.com/wichat/wichat/backend/wi-shared/microservices/utils/correlation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	metadataRequestIDKey     = "x-request-id"
	metadataInternalTokenKey = "x-wichat-internal-token"
)

// UnaryClientInternalToken attaches a shared secret for service-to-service gRPC when token is non-empty.
func UnaryClientInternalToken(token string) grpc.UnaryClientInterceptor {
	tok := strings.TrimSpace(token)
	if tok == "" {
		return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
	}
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = metadata.AppendToOutgoingContext(ctx, metadataInternalTokenKey, tok)

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// UnaryServerInternalToken rejects calls when want is set and metadata token does not match.
// General health RPCs are typically registered on a separate server without this interceptor.
func UnaryServerInternalToken(want string) grpc.UnaryServerInterceptor {
	expected := strings.TrimSpace(want)
	if expected == "" {
		return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			return handler(ctx, req)
		}
	}
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if strings.HasSuffix(info.FullMethod, "/Health") || strings.Contains(info.FullMethod, ".GeneralService/") {
			return handler(ctx, req)
		}
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing internal credentials")
		}
		vals := md.Get(metadataInternalTokenKey)
		if len(vals) == 0 || vals[0] != expected {
			return nil, status.Error(codes.Unauthenticated, "invalid internal credentials")
		}

		return handler(ctx, req)
	}
}

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
