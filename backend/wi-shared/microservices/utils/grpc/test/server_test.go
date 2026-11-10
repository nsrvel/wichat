package grpc_test

import (
	"context"
	"net"
	"testing"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	grpcutil "github.com/wichat/wichat/backend/wi-shared/microservices/utils/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestNewServerNonNil(t *testing.T) {
	t.Parallel()

	log := logger.New(logger.Options{Service: "grpc-test", Env: "test", Level: "error"})
	if grpcutil.NewServer(log) == nil {
		t.Fatal("expected server")
	}
}

type panicService interface {
	Go(context.Context, *struct{}) (*struct{}, error)
}

type panicServer struct {
	internal bool
}

func (p panicServer) Go(context.Context, *struct{}) (*struct{}, error) {
	if p.internal {
		return nil, status.Error(codes.Internal, "boom")
	}
	panic("boom")
}

func TestNewServerRecoveryInternalError(t *testing.T) {
	t.Parallel()

	lis := bufconn.Listen(1 << 20)
	log := logger.New(logger.Options{Service: "grpc-test", Env: "test", Level: "error"})
	srv := grpcutil.NewServer(log)
	srv.RegisterService(&grpc.ServiceDesc{
		ServiceName: "test.Internal",
		HandlerType: (*panicService)(nil),
		Methods: []grpc.MethodDesc{{
			MethodName: "Go",
			Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
				in := new(struct{})
				if err := dec(in); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return srv.(panicService).Go(ctx, in)
				}

				return interceptor(ctx, srv, &grpc.UnaryServerInfo{
					Server:     srv,
					FullMethod: "/test.Internal/Go",
				}, func(ctx context.Context, req any) (any, error) {
					return srv.(panicService).Go(ctx, req.(*struct{}))
				})
			},
		}},
	}, panicServer{internal: true})

	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(func() { srv.Stop() })

	dialer := func(context.Context, string) (net.Conn, error) { return lis.Dial() }
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	var resp struct{}
	err = conn.Invoke(context.Background(), "/test.Internal/Go", &struct{}{}, &resp)
	if status.Code(err) != codes.Internal {
		t.Fatalf("code %v", err)
	}
}

func TestNewServerRecoveryOnPanic(t *testing.T) {
	t.Parallel()

	lis := bufconn.Listen(1 << 20)
	log := logger.New(logger.Options{Service: "grpc-test", Env: "test", Level: "error"})
	srv := grpcutil.NewServer(log)
	srv.RegisterService(&grpc.ServiceDesc{
		ServiceName: "test.Panic",
		HandlerType: (*panicService)(nil),
		Methods: []grpc.MethodDesc{{
			MethodName: "Go",
			Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
				in := new(struct{})
				if err := dec(in); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return srv.(panicService).Go(ctx, in)
				}

				return interceptor(ctx, srv, &grpc.UnaryServerInfo{
					Server:     srv,
					FullMethod: "/test.Panic/Go",
				}, func(ctx context.Context, req any) (any, error) {
					return srv.(panicService).Go(ctx, req.(*struct{}))
				})
			},
		}},
	}, panicServer{internal: false})

	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(func() { srv.Stop() })

	dialer := func(context.Context, string) (net.Conn, error) { return lis.Dial() }
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	var resp struct{}
	err = conn.Invoke(context.Background(), "/test.Panic/Go", &struct{}{}, &resp)
	if err == nil {
		t.Fatal("expected error")
	}
	if status.Code(err) != codes.Internal {
		t.Fatalf("code %v", err)
	}
}
