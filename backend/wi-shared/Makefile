GOLANGCI_VERSION := v2.1.6
GOLANGCI := go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)

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

check: fmt lint test
