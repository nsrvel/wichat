// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

package config_test

import (
	"testing"

	"github.com/wichat/wichat/backend/wi-shared/config"
)

func TestListenAddress(t *testing.T) {
	if config.ListenAddress("3000") != ":3000" {
		t.Fatal("port only")
	}
	if config.ListenAddress(":3000") != ":3000" {
		t.Fatal("with colon")
	}
	if config.ListenAddress("localhost:3001") != "localhost:3001" {
		t.Fatal("host:port")
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
