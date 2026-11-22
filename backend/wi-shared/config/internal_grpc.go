package config

import (
	"fmt"
	"strings"
)

// RequireInternalGrpcToken enforces a shared secret for service-to-service gRPC (all envs).
func RequireInternalGrpcToken(env, token string) error {
	_ = env
	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("INTERNAL_GRPC_TOKEN is required")
	}
	return nil
}

// IsDevelopmentEnv reports whether env is local development.
func IsDevelopmentEnv(env string) bool {
	return strings.EqualFold(strings.TrimSpace(env), "development")
}
