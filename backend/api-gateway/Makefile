GOLANGCI_VERSION := v2.1.6
GOLANGCI := go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)
AIR_VERSION := v1.62.0
AIR := go run github.com/air-verse/air@$(AIR_VERSION)

.PHONY: fmt lint test check dev run

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
