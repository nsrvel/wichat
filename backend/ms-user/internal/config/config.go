package config

import (
	sharedconfig "github.com/wichat/wichat/backend/wi-shared/config"
)

// Config is ms-user service configuration.
type Config struct {
	Env                string `mapstructure:"env"`
	LogLevel           string `mapstructure:"log_level"`
	Port               string `mapstructure:"port"` // gRPC listen port
	DatabaseURL        string `mapstructure:"database_url"`
	AuthJWTSecret      string `mapstructure:"auth_jwt_secret"`
	GoogleClientID     string `mapstructure:"google_client_id"`
	GoogleClientSecret string `mapstructure:"google_client_secret"`
	GoogleRedirectURL  string `mapstructure:"google_redirect_url"`
	RedisURL           string `mapstructure:"redis_url"`
	NotificationHost   string `mapstructure:"ms_notification_host"`
	MinioEndpoint      string `mapstructure:"minio_endpoint"`
	MinioAccessKey     string `mapstructure:"minio_access_key"`
	MinioSecretKey     string `mapstructure:"minio_secret_key"`
	MinioBucket        string `mapstructure:"minio_bucket"`
	MinioPublicBaseURL string `mapstructure:"minio_public_base_url"`
	MinioUseSSL        bool   `mapstructure:"minio_use_ssl"`
	MinioRegion        string `mapstructure:"minio_region"`
	WebOrigin          string `mapstructure:"web_origin"`
}

// Defaults are applied when env / .env do not set a value.
var Defaults = Config{
	Env:              "development",
	LogLevel:         "info",
	Port:             "3001",
	RedisURL:         "redis://localhost:6379/0",
	NotificationHost: "localhost:3005",
	WebOrigin:        "http://localhost:8080",
}

// Load reads configuration from defaults, optional envFile, and environment.
func Load(envFile string) (Config, error) {
	var cfg Config

	if err := sharedconfig.Load(envFile, &cfg, Defaults); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// ListenAddr returns the address for net.Listen (e.g. ":3001").
func (c Config) ListenAddr() string {
	return sharedconfig.ListenAddress(c.Port)
}
