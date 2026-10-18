package config

import (
	"strings"

	sharedconfig "github.com/wichat/wichat/backend/wi-shared/config"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
)

// Config is api-gateway service configuration.
type Config struct {
	Env           string         `mapstructure:"env"`
	LogLevel      string         `mapstructure:"log_level"`
	Port          string         `mapstructure:"port"`
	UserHost      string         `mapstructure:"ms_user_host"`
	WebOrigin     string         `mapstructure:"web_origin"`
	Microservices utils.Settings `mapstructure:",squash"`
}

// Defaults are applied when env / .env do not set a value.
var Defaults = Config{
	Env:           "development",
	LogLevel:      "info",
	Port:          "3000",
	UserHost:      "localhost:3001",
	WebOrigin:     "http://localhost:8080",
	Microservices: utils.DefaultSettings,
}

// Load reads configuration from defaults, optional envFile, and environment.
func Load(envFile string) (Config, error) {
	var cfg Config

	if err := sharedconfig.Load(envFile, &cfg, Defaults); err != nil {
		return Config{}, err
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
