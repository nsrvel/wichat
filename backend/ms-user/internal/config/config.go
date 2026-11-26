package config

import (
	"fmt"
	"strings"

	sharedconfig "github.com/wichat/wichat/backend/wi-shared/config"
)

// Config is ms-user service configuration.
type Config struct {
	Env                   string `mapstructure:"env"`
	LogLevel              string `mapstructure:"log_level"`
	Port                  string `mapstructure:"port"`      // gRPC listen port
	GrpcBind              string `mapstructure:"grpc_bind"` // loopback by default in dev
	DatabaseURL           string `mapstructure:"database_url"`
	AuthJWTSecret         string `mapstructure:"auth_jwt_secret"`
	GoogleClientID        string `mapstructure:"google_client_id"`
	GoogleClientSecret    string `mapstructure:"google_client_secret"`
	GoogleRedirectURL     string `mapstructure:"google_redirect_url"`
	RedisURL              string `mapstructure:"redis_url"`
	NotificationHost      string `mapstructure:"ms_notification_host"`
	MinioEndpoint         string `mapstructure:"minio_endpoint"`
	MinioAccessKey        string `mapstructure:"minio_access_key"`
	MinioSecretKey        string `mapstructure:"minio_secret_key"`
	MinioBucket           string `mapstructure:"minio_bucket"`
	MinioPublicBaseURL    string `mapstructure:"minio_public_base_url"`
	MinioUseSSL           bool   `mapstructure:"minio_use_ssl"`
	MinioRegion           string `mapstructure:"minio_region"`
	WebOrigin             string `mapstructure:"web_origin"`
	MFAEncryptionKey      string `mapstructure:"mfa_encryption_key"`
	WebAuthnRPID          string `mapstructure:"webauthn_rp_id"`
	WebAuthnRPOrigins     string `mapstructure:"webauthn_rp_origins"`
	WebAuthnRPDisplayName string `mapstructure:"webauthn_rp_display_name"`
	InternalGrpcToken     string `mapstructure:"internal_grpc_token"`
	InviteOnly            bool   `mapstructure:"invite_only"`
	MFAEnabled            bool   `mapstructure:"mfa_enabled"`
}

// Defaults are applied when env / .env do not set a value.
var Defaults = Config{
	Env:                   "development",
	LogLevel:              "info",
	Port:                  "3001",
	GrpcBind:              "127.0.0.1",
	InternalGrpcToken:     "dev-internal-grpc-token",
	RedisURL:              "redis://localhost:6379/0",
	NotificationHost:      "localhost:3005",
	WebOrigin:             "http://localhost:8080",
	WebAuthnRPDisplayName: "WiChat",
	MFAEnabled:            true,
}

// Load reads configuration from defaults, optional envFile, and environment.
func Load(envFile string) (Config, error) {
	var cfg Config

	if err := sharedconfig.Load(envFile, &cfg, Defaults); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// ListenAddr returns the address for net.Listen (e.g. "127.0.0.1:3001").
func (c Config) ListenAddr() string {
	return sharedconfig.GrpcListenAddress(c.GrpcBind, c.Port)
}

// Validate enforces production safety invariants when Env is not development.
func (c Config) Validate() error {
	if err := sharedconfig.RequireInternalGrpcToken(c.Env, c.InternalGrpcToken); err != nil {
		return err
	}
	if strings.EqualFold(strings.TrimSpace(c.Env), "development") {
		return nil
	}
	origin := strings.ToLower(strings.TrimSpace(c.WebOrigin))
	if !strings.HasPrefix(origin, "https://") {
		return fmt.Errorf("WEB_ORIGIN must use https:// when ENV=%s", c.Env)
	}
	if c.MFAEnabled {
		mfaKey := strings.TrimSpace(c.MFAEncryptionKey)
		if len(mfaKey) < 32 {
			return fmt.Errorf("MFA_ENCRYPTION_KEY must be at least 32 characters when ENV=%s and MFA_ENABLED=true", c.Env)
		}
		if mfaKey == c.AuthJWTSecret {
			return fmt.Errorf("MFA_ENCRYPTION_KEY must be distinct from AUTH_JWT_SECRET when ENV=%s and MFA_ENABLED=true", c.Env)
		}
	}
	return nil
}
