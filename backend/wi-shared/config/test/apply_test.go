package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wichat/wichat/backend/wi-shared/config"
)

func TestListenAddress(t *testing.T) {
	if config.ListenAddress("3000") != ":3000" {
		t.Fatal("port only")
	}
	if config.ListenAddress("127.0.0.1:8080") != "127.0.0.1:8080" {
		t.Fatal("host:port")
	}
}

func TestListenAddressEmpty(t *testing.T) {
	if config.ListenAddress("  ") != "" {
		t.Fatal("empty")
	}
}

func TestLoad_emptyEnvStringUsesDefault(t *testing.T) {
	t.Setenv("GRPC_ADDR", "")

	var cfg testConfig
	if err := config.Load("", &cfg, testDefaults); err != nil {
		t.Fatal(err)
	}
	if cfg.GRPCAddr != ":50051" {
		t.Fatalf("grpc_addr: got %q", cfg.GRPCAddr)
	}
}

func TestLoad_envFileEmptyKeyUsesDefault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("GRPC_ADDR=\nLOG_LEVEL=debug\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var cfg testConfig
	if err := config.Load(path, &cfg, testDefaults); err != nil {
		t.Fatal(err)
	}
	if cfg.GRPCAddr != ":50051" {
		t.Fatalf("grpc_addr: got %q", cfg.GRPCAddr)
	}
	if cfg.LogLevel != "debug" {
		t.Fatalf("log_level: got %q", cfg.LogLevel)
	}
}

func TestLoad_defaultsTypeMismatch(t *testing.T) {
	type other struct {
		GRPCAddr string `mapstructure:"grpc_addr"`
	}
	var cfg testConfig
	err := config.Load("", &cfg, other{GRPCAddr: ":1"})
	if err == nil {
		t.Fatal("expected type mismatch")
	}
}

func TestLoad_defaultsStructValueForm(t *testing.T) {
	var cfg testConfig
	if err := config.Load("", &cfg, testDefaults); err != nil {
		t.Fatal(err)
	}
}

func TestLoad_defaultsBadPointer(t *testing.T) {
	var cfg testConfig
	var bad *testConfig
	err := config.Load("", &cfg, bad)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoad_defaultsNonStruct(t *testing.T) {
	var cfg testConfig
	err := config.Load("", &cfg, 42)
	if err == nil {
		t.Fatal("expected error")
	}
}

type squashedInner struct {
	Host string `mapstructure:"host"`
}

type squashedConfig struct {
	squashedInner `mapstructure:",squash"`
	GRPCAddr      string `mapstructure:"grpc_addr"`
	LogLevel      string `mapstructure:"log_level"`
}

func TestLoad_ignoredMapstructureField(t *testing.T) {
	type cfg struct {
		Secret   string `mapstructure:"-"`
		LogLevel string `mapstructure:"log_level"`
	}
	def := cfg{Secret: "x", LogLevel: "info"}
	var c cfg
	t.Setenv("LOG_LEVEL", "debug")
	if err := config.Load("", &c, def); err != nil {
		t.Fatal(err)
	}
	if c.LogLevel != "debug" || c.Secret != "x" {
		t.Fatalf("got %+v", c)
	}
}

func TestLoad_envFileOnlyEmptyLogLevelUsesDefault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("LOG_LEVEL=\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_ = os.Unsetenv("LOG_LEVEL")

	var cfg testConfig
	if err := config.Load(path, &cfg, testDefaults); err != nil {
		t.Fatal(err)
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("log_level: %q", cfg.LogLevel)
	}
}

func TestLoad_envFileOverridesWithoutOsEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("LOG_LEVEL=trace\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_ = os.Unsetenv("LOG_LEVEL")

	var cfg testConfig
	if err := config.Load(path, &cfg, testDefaults); err != nil {
		t.Fatal(err)
	}
	if cfg.LogLevel != "trace" {
		t.Fatalf("log_level: %q", cfg.LogLevel)
	}
}

func TestLoad_squashedStructEmptyEnvUsesDefault(t *testing.T) {
	t.Setenv("GRPC_ADDR", "")

	var cfg squashedConfig
	defLoad := squashedConfig{
		squashedInner: squashedInner{Host: "localhost"},
		GRPCAddr:      ":50051",
		LogLevel:      "info",
	}
	if err := config.Load("", &cfg, defLoad); err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "localhost" {
		t.Fatalf("host: %q", cfg.Host)
	}
	if cfg.GRPCAddr != ":50051" {
		t.Fatalf("grpc_addr: %q", cfg.GRPCAddr)
	}
}
