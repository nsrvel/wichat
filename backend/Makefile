# Backend workspace scripts (like package.json scripts). Run from backend/.
.PHONY: help check dev-ms-user dev-ms-notification dev-api-gateway run-ms-user run-ms-notification run-api-gateway migrate-up migrate-down migrate-status

help:
	@echo "WiChat backend — common targets:"
	@echo "  make migrate-up         apply Postgres migrations (goose)"
	@echo "  make migrate-status     show migration state"
	@echo "  make check              lint + test (wi-shared, ms-notification, ms-user, api-gateway)"
	@echo "  make dev-ms-user        ms-user with Air live reload"
	@echo "  make dev-ms-notification ms-notification with Air live reload"
	@echo "  make dev-api-gateway    api-gateway with Air live reload"
	@echo "  make run-ms-user        ms-user once (go run, no reload)"
	@echo "  make run-ms-notification ms-notification once (go run, no reload)"
	@echo "  make run-api-gateway    api-gateway once (go run, no reload)"
	@echo ""
	@echo "Per module: cd ms-user|api-gateway|wi-shared && make dev|run|check"

check:
	$(MAKE) -C wi-shared check
	$(MAKE) -C ms-notification check
	$(MAKE) -C ms-user check
	$(MAKE) -C api-gateway check

dev-ms-user:
	$(MAKE) -C ms-user dev

dev-ms-notification:
	$(MAKE) -C ms-notification dev

dev-api-gateway:
	$(MAKE) -C api-gateway dev

run-ms-user:
	$(MAKE) -C ms-user run

run-ms-notification:
	$(MAKE) -C ms-notification run

run-api-gateway:
	$(MAKE) -C api-gateway run

migrate-up:
	$(MAKE) -C ms-user migrate-up

migrate-down:
	$(MAKE) -C ms-user migrate-down

migrate-status:
	$(MAKE) -C ms-user migrate-status
