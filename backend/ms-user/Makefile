GOLANGCI_VERSION := v2.1.6
GOLANGCI := go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)
AIR_VERSION := v1.62.0
AIR := go run github.com/air-verse/air@$(AIR_VERSION)
GOOSE := go tool goose
MIGRATIONS_DIR := internal/migrations

ifneq (,$(wildcard .env))
include .env
export
endif

DATABASE_URL ?= postgres://wichat:wichat@localhost:5432/wichat?sslmode=disable

.PHONY: fmt lint test check dev run migrate-up migrate-down migrate-status migrate-create

fmt:
	$(GOLANGCI) fmt ./...

lint:
	$(GOLANGCI) run ./...

test:
	go test ./...

check: fmt lint test

dev:
	$(AIR)

run:
	go run ./cmd

migrate-up:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

migrate-down:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

migrate-status:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" status

migrate-create:
	@test -n "$(NAME)" || (echo 'usage: make migrate-create NAME=00003_example' && exit 1)
	$(GOOSE) -dir $(MIGRATIONS_DIR) create $(NAME) sql
