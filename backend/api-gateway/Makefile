GOLANGCI_VERSION := v2.1.6
GOLANGCI := go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)
AIR_VERSION := v1.62.0
AIR := go run github.com/air-verse/air@$(AIR_VERSION)

ifneq (,$(wildcard .env))
include .env
export
endif

BACKEND_ROOT := ..

.PHONY: test-cover
test-cover:
	@bash "$(BACKEND_ROOT)/scripts/coverage-gate.sh" "$(CURDIR)" 90

.PHONY: fmt lint test check dev run tunnel

fmt:
	$(GOLANGCI) fmt ./...

lint:
	$(GOLANGCI) run ./...

test:
	go test ./...

check: fmt lint test test-cover

dev:
	$(AIR)

run:
	go run ./cmd

TUNNEL_ORIGIN ?= https://wichat.wistudio.my.id
TUNNEL_TRUSTED_PROXIES ?= 127.0.0.1,::1

tunnel:
	WEB_ORIGIN=$(TUNNEL_ORIGIN) TRUSTED_PROXIES=$(TUNNEL_TRUSTED_PROXIES) go run ./cmd
