GOLANGCI_VERSION := v2.1.6
GOLANGCI := go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)

BACKEND_ROOT := ..

.PHONY: test-cover
test-cover:
	@bash "$(BACKEND_ROOT)/scripts/coverage-gate.sh" "$(CURDIR)" 90

.PHONY: fmt lint test check proto buf-lint

fmt:
	$(GOLANGCI) fmt ./...

lint:
	$(GOLANGCI) run ./...

test:
	go test ./...

proto:
	buf generate

buf-lint:
	buf lint

check: fmt lint test test-cover
