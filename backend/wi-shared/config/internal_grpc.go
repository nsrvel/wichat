package config

import (
	"fmt"
	"strings"
)

// RequireInternalGrpcToken enforces a shared secret for service-to-service gRPC in non-development envs.
func RequireInternalGrpcToken(env, token string) error {
	if IsDevelopmentEnv(env) {
		return nil
	}
	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("INTERNAL_GRPC_TOKEN is required when ENV=%s", strings.TrimSpace(env))
	}
	return nil
}

// IsDevelopmentEnv reports whether env is local development.
func IsDevelopmentEnv(env string) bool {
	return strings.EqualFold(strings.TrimSpace(env), "development")
}
