// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wichat/wichat/backend/wi-shared/config"
)

type testConfig struct {
	GRPCAddr string `mapstructure:"grpc_addr"`
	LogLevel string `mapstructure:"log_level"`
}

var testDefaults = testConfig{
	GRPCAddr: ":50051",
	LogLevel: "info",
}

func TestLoad_defaults(t *testing.T) {
	var cfg testConfig
	if err := config.Load("", &cfg, testDefaults); err != nil {
		t.Fatal(err)
	}
	if cfg.GRPCAddr != ":50051" || cfg.LogLevel != "info" {
		t.Fatalf("got %+v", cfg)
	}
}

func TestLoad_envOverridesDefault(t *testing.T) {
	t.Setenv("GRPC_ADDR", ":59999")

	var cfg testConfig
	if err := config.Load("", &cfg, testDefaults); err != nil {
		t.Fatal(err)
	}
	if cfg.GRPCAddr != ":59999" {
		t.Fatalf("grpc_addr: got %q", cfg.GRPCAddr)
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("log_level: got %q", cfg.LogLevel)
	}
}

func TestLoad_envFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("LOG_LEVEL=debug\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var cfg testConfig
	if err := config.Load(path, &cfg, testDefaults); err != nil {
		t.Fatal(err)
	}
	if cfg.LogLevel != "debug" {
		t.Fatalf("log_level: got %q", cfg.LogLevel)
	}
}

func TestLoad_envWinsOverEnvFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("GRPC_ADDR=:50052\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GRPC_ADDR", ":50053")

	var cfg testConfig
	if err := config.Load(path, &cfg, testDefaults); err != nil {
		t.Fatal(err)
	}
	if cfg.GRPCAddr != ":50053" {
		t.Fatalf("grpc_addr: got %q want :50053", cfg.GRPCAddr)
	}
}

func TestLoad_nilDst(t *testing.T) {
	err := config.Load("", nil, testDefaults)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoad_missingEnvFileOK(t *testing.T) {
	var cfg testConfig
	if err := config.Load(".env-does-not-exist", &cfg, testDefaults); err != nil {
		t.Fatal(err)
	}
	if cfg.GRPCAddr != ":50051" {
		t.Fatalf("grpc_addr: got %q", cfg.GRPCAddr)
	}
}
