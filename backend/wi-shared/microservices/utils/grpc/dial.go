package grpc

import (
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// DialOptions configures a shared outbound gRPC connection.
type DialOptions struct {
	// Insecure uses plaintext gRPC (default for local dev).
	Insecure bool
}

// DefaultDialOptions enables insecure transport for development compose setups.
var DefaultDialOptions = DialOptions{Insecure: true}

// Dial opens a client connection with request ID propagation. Optional extra grpc.DialOption values append after defaults.
func Dial(target string, opts DialOptions, extra ...grpc.DialOption) (*grpc.ClientConn, error) {
	if target == "" {
		return nil, fmt.Errorf("microservices/utils/grpc: dial target is empty")
	}

	dialOpts := make([]grpc.DialOption, 0, len(extra)+4)
	dialOpts = append(dialOpts, extra...)
	dialOpts = append(dialOpts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:                30 * time.Second,
		Timeout:             5 * time.Second,
		PermitWithoutStream: true,
	}))
	dialOpts = append(dialOpts, grpc.WithChainUnaryInterceptor(UnaryClientRequestID()))
	if opts.Insecure {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(target, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("microservices/utils/grpc: dial %s: %w", target, err)
	}

	return conn, nil
}
