package config

import (
	"fmt"
	"os"
	"strings"

	sharedconfig "github.com/wichat/wichat/backend/wi-shared/config"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
)

// Config is api-gateway service configuration.
type Config struct {
	Env                        string         `mapstructure:"env"`
	LogLevel                   string         `mapstructure:"log_level"`
	Port                       string         `mapstructure:"port"`
	UserHost                   string         `mapstructure:"ms_user_host"`
	ChatHost                   string         `mapstructure:"ms_chat_host"`
	NotificationHost           string         `mapstructure:"ms_notification_host"`
	WebOrigin                  string         `mapstructure:"web_origin"`
	KafkaBrokers               string         `mapstructure:"kafka_brokers"`
	KafkaChatEventsTopic       string         `mapstructure:"kafka_chat_events_topic"`
	KafkaRealtimeConsumerGroup string         `mapstructure:"kafka_realtime_consumer_group"`
	InstanceID                 string         `mapstructure:"instance_id"`
	TrustedProxies             string         `mapstructure:"trusted_proxies"`
	TurnstileSiteKey           string         `mapstructure:"turnstile_site_key"`
	TurnstileSecretKey         string         `mapstructure:"turnstile_secret_key"`
	Microservices              utils.Settings `mapstructure:",squash"`
}

// Defaults are applied when env / .env do not set a value.
var Defaults = Config{
	Env:                        "development",
	LogLevel:                   "info",
	Port:                       "3000",
	UserHost:                   "localhost:3001",
	ChatHost:                   "localhost:3002",
	NotificationHost:           "localhost:3005",
	WebOrigin:                  "http://localhost:8080",
	KafkaBrokers:               "localhost:9092",
	KafkaChatEventsTopic:       "chat-events",
	KafkaRealtimeConsumerGroup: "api-gateway-realtime",
	InstanceID:                 "local",
	Microservices:              utils.DefaultSettings,
}

// Load reads configuration from defaults, optional envFile, and environment.
func Load(envFile string) (Config, error) {
	var cfg Config

	if err := sharedconfig.Load(envFile, &cfg, Defaults); err != nil {
		return Config{}, err
	}

	if strings.TrimSpace(cfg.InstanceID) == "" || cfg.InstanceID == "local" {
		if host, err := os.Hostname(); err == nil && strings.TrimSpace(host) != "" {
			cfg.InstanceID = host
		}
	}

	return cfg, nil
}

// HTTPListenAddr returns the address for net/http.Server (e.g. ":3000").
func (c Config) HTTPListenAddr() string {
	return sharedconfig.ListenAddress(c.Port)
}

// SecureCookies is false for local HTTP development.
func (c Config) SecureCookies() bool {
	return !strings.EqualFold(strings.TrimSpace(c.Env), "development")
}

// RealtimeKafkaGroupID is a per-instance consumer group for broadcast fanout.
func (c Config) RealtimeKafkaGroupID() string {
	prefix := strings.TrimSpace(c.KafkaRealtimeConsumerGroup)
	if prefix == "" {
		prefix = "api-gateway-realtime"
	}
	instance := strings.TrimSpace(c.InstanceID)
	if instance == "" {
		instance = "unknown"
	}
	return fmt.Sprintf("%s-%s", prefix, instance)
}
