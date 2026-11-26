package config

import (
	sharedconfig "github.com/wichat/wichat/backend/wi-shared/config"
)

// Config is ms-notification service configuration.
type Config struct {
	Env          string `mapstructure:"env"`
	LogLevel     string `mapstructure:"log_level"`
	Port         string `mapstructure:"port"`
	GrpcBind     string `mapstructure:"grpc_bind"`
	SMTPHost     string `mapstructure:"smtp_host"`
	SMTPPort     int    `mapstructure:"smtp_port"`
	SMTPUser     string `mapstructure:"smtp_user"`
	SMTPPassword string `mapstructure:"smtp_password"`
	SMTPFrom     string `mapstructure:"smtp_from"`
	// SMTPSecure: implicit TLS (typical port 465). false = plain connect + STARTTLS (typical port 587).
	SMTPSecure        bool   `mapstructure:"smtp_secure"`
	InternalGrpcToken string `mapstructure:"internal_grpc_token"`
}

// Defaults are applied when env / .env do not set a value.
var Defaults = Config{
	Env:               "development",
	LogLevel:          "info",
	Port:              "3005",
	GrpcBind:          "127.0.0.1",
	InternalGrpcToken: "dev-internal-grpc-token",
	SMTPHost:          "localhost",
	SMTPPort:          587,
	SMTPFrom:          "WiChat <noreply@localhost>",
}

// Load reads configuration from defaults, optional envFile, and environment.
func Load(envFile string) (Config, error) {
	var cfg Config

	if err := sharedconfig.Load(envFile, &cfg, Defaults); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// ListenAddr returns the address for net.Listen (e.g. "127.0.0.1:3005").
func (c Config) ListenAddr() string {
	return sharedconfig.GrpcListenAddress(c.GrpcBind, c.Port)
}

// Validate enforces production safety invariants when Env is not development.
func (c Config) Validate() error {
	return sharedconfig.RequireInternalGrpcToken(c.Env, c.InternalGrpcToken)
}
