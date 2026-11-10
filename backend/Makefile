# Backend workspace scripts (like package.json scripts). Run from backend/.
REPO_ROOT := ..
COMPOSE_FILE := $(REPO_ROOT)/deploy/docker-compose.yml
COMPOSE := docker compose -f $(COMPOSE_FILE)

.PHONY: help check run run-infra run-infra-minio-check run-down dev-ms-user dev-ms-chat dev-ms-notification dev-api-gateway run-ms-user run-ms-chat run-ms-notification run-api-gateway migrate-up migrate-down migrate-status

help:
	@echo "WiChat backend — common targets:"
	@echo "  make run                docker compose up + all Go services (go run, no Air)"
	@echo "                          env: each ms-*/.env (see .env.example)"
	@echo "  make run-infra          full compose up --wait (+ minio health curl)"
	@echo "  make run-down           docker compose down (infra only; does not stop a running make run)"
	@echo "  make migrate-up         apply Postgres migrations (goose)"
	@echo "  make migrate-status     show migration state"
	@echo "  make check              lint + test (wi-shared, ms-chat, ms-notification, ms-user, api-gateway)"
	@echo "  make dev-ms-user        ms-user with Air live reload"
	@echo "  make dev-ms-chat        ms-chat with Air live reload"
	@echo "  make dev-ms-notification ms-notification with Air live reload"
	@echo "  make dev-api-gateway    api-gateway with Air live reload"
	@echo "  make run-ms-user        ms-user once (go run, no reload)"
	@echo "  make run-ms-chat        ms-chat once (go run, no reload)"
	@echo "  make run-ms-notification ms-notification once (go run, no reload)"
	@echo "  make run-api-gateway    api-gateway once (go run, no reload)"
	@echo ""
	@echo "Per module: cd ms-user|api-gateway|wi-shared && make dev|run|check"

check:
	$(MAKE) -C wi-shared check
	$(MAKE) -C ms-chat check
	$(MAKE) -C ms-notification check
	$(MAKE) -C ms-user check
	$(MAKE) -C api-gateway check

run-infra:
	@echo "Starting full stack: postgres, redis, minio, kafka, elasticsearch..."
	$(COMPOSE) up -d --wait --wait-timeout 240
	@$(MAKE) run-infra-minio-check COMPOSE_CMD='$(COMPOSE)'

run-infra-minio-check:
	@bash -ec '\
		echo "Checking minio (no compose healthcheck)..."; \
		for i in $$(seq 1 30); do \
			curl -sf http://127.0.0.1:9000/minio/health/live >/dev/null && exit 0; \
			sleep 1; \
		done; \
		echo "MinIO not ready on :9000"; $(COMPOSE_CMD) ps; exit 1 \
	'

run-down:
	$(COMPOSE) down

# Infra + microservices (no Air). Ctrl+C stops Go children; infra stays up (use make run-down).
run: run-infra
	@bash -ec '\
		trap "kill 0" INT TERM; \
		echo "Starting ms-user (3001)..."; \
		$(MAKE) -C ms-user run & \
		sleep 1; \
		echo "Starting ms-chat (3002)..."; \
		$(MAKE) -C ms-chat run & \
		echo "Starting ms-notification (3005)..."; \
		$(MAKE) -C ms-notification run & \
		sleep 1; \
		echo "Starting api-gateway (3000)..."; \
		$(MAKE) -C api-gateway run & \
		wait \
	'

dev-ms-user:
	$(MAKE) -C ms-user dev

dev-ms-chat:
	$(MAKE) -C ms-chat dev

dev-ms-notification:
	$(MAKE) -C ms-notification dev

dev-api-gateway:
	$(MAKE) -C api-gateway dev

run-ms-user:
	$(MAKE) -C ms-user run

run-ms-chat:
	$(MAKE) -C ms-chat run

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
