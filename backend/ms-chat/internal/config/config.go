package config

import (
	sharedconfig "github.com/wichat/wichat/backend/wi-shared/config"
)

// Config is ms-chat service configuration.
type Config struct {
	Env                      string `mapstructure:"env"`
	LogLevel                 string `mapstructure:"log_level"`
	Port                     string `mapstructure:"port"`
	GrpcBind                 string `mapstructure:"grpc_bind"`
	DatabaseURL              string `mapstructure:"database_url"`
	MSUserHost               string `mapstructure:"ms_user_host"`
	KafkaBrokers             string `mapstructure:"kafka_brokers"`
	KafkaChatEventsTopic     string `mapstructure:"kafka_chat_events_topic"`
	KafkaConsumerGroup       string `mapstructure:"kafka_consumer_group"`
	ElasticsearchURL         string `mapstructure:"elasticsearch_url"`
	ElasticsearchMessagesIdx string `mapstructure:"elasticsearch_messages_index"`
	MinioEndpoint            string `mapstructure:"minio_endpoint"`
	MinioAccessKey           string `mapstructure:"minio_access_key"`
	MinioSecretKey           string `mapstructure:"minio_secret_key"`
	MinioBucket              string `mapstructure:"minio_bucket"`
	MinioPublicBaseURL       string `mapstructure:"minio_public_base_url"`
	MinioUseSSL              bool   `mapstructure:"minio_use_ssl"`
	MinioRegion              string `mapstructure:"minio_region"`
	InternalGrpcToken        string `mapstructure:"internal_grpc_token"`
}

// Defaults are applied when env / .env do not set a value.
var Defaults = Config{
	Env:                      "development",
	LogLevel:                 "info",
	Port:                     "3002",
	GrpcBind:                 "127.0.0.1",
	InternalGrpcToken:        "dev-internal-grpc-token",
	DatabaseURL:              "postgres://wichat:wichat@localhost:5432/wichat?sslmode=disable",
	MSUserHost:               "localhost:3001",
	KafkaBrokers:             "localhost:9092",
	KafkaChatEventsTopic:     "chat-events",
	KafkaConsumerGroup:       "ms-chat-search-indexer",
	ElasticsearchURL:         "http://localhost:9200",
	ElasticsearchMessagesIdx: "wichat-messages",
}

// Load reads configuration from defaults, optional envFile, and environment.
func Load(envFile string) (Config, error) {
	var cfg Config

	if err := sharedconfig.Load(envFile, &cfg, Defaults); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// ListenAddr returns the address for net.Listen (e.g. "127.0.0.1:3002").
func (c Config) ListenAddr() string {
	return sharedconfig.GrpcListenAddress(c.GrpcBind, c.Port)
}

// Validate enforces production safety invariants when Env is not development.
func (c Config) Validate() error {
	return sharedconfig.RequireInternalGrpcToken(c.Env, c.InternalGrpcToken)
}
