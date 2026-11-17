package config

import sharedconfig "github.com/wichat/wichat/backend/wi-shared/config"

// Config is ms-core service configuration.
type Config struct {
	Env                        string `mapstructure:"env"`
	LogLevel                   string `mapstructure:"log_level"`
	Port                       string `mapstructure:"port"`
	DatabaseURL                string `mapstructure:"database_url"`
	RedisURL                   string `mapstructure:"redis_url"`
	MSUserHost                 string `mapstructure:"ms_user_host"`
	KafkaBrokers               string `mapstructure:"kafka_brokers"`
	KafkaChatEventsTopic       string `mapstructure:"kafka_chat_events_topic"`
	LiveKitURL                 string `mapstructure:"livekit_url"`
	LiveKitAPIKey              string `mapstructure:"livekit_api_key"`
	LiveKitAPISecret           string `mapstructure:"livekit_api_secret"`
	LiveKitMonthlyQuotaMinutes int64  `mapstructure:"livekit_monthly_quota_minutes"`
	LiveKitEnforceHardCap      bool   `mapstructure:"livekit_enforce_hard_cap"`
	MinioEndpoint              string `mapstructure:"minio_endpoint"`
	MinioAccessKey             string `mapstructure:"minio_access_key"`
	MinioSecretKey             string `mapstructure:"minio_secret_key"`
	MinioBucket                string `mapstructure:"minio_bucket"`
	MinioPublicBaseURL         string `mapstructure:"minio_public_base_url"`
	MinioUseSSL                bool   `mapstructure:"minio_use_ssl"`
	MinioRegion                string `mapstructure:"minio_region"`
}

// Defaults are applied when env / .env do not set a value.
var Defaults = Config{
	Env:                        "development",
	LogLevel:                   "info",
	Port:                       "3003",
	RedisURL:                   "redis://localhost:6379/0",
	MSUserHost:                 "localhost:3001",
	KafkaBrokers:               "localhost:9092",
	KafkaChatEventsTopic:       "chat-events",
	LiveKitMonthlyQuotaMinutes: 5000,
}

// Load reads configuration from defaults, optional envFile, and environment.
func Load(envFile string) (Config, error) {
	var cfg Config
	if err := sharedconfig.Load(envFile, &cfg, Defaults); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// ListenAddr returns the address for net.Listen (e.g. ":3003").
func (c Config) ListenAddr() string {
	return sharedconfig.ListenAddress(c.Port)
}
