# Deployment

## Local infrastructure (all services)

From repo root:

```bash
docker compose -f deploy/docker-compose.yml up -d
```

Stop:

```bash
docker compose -f deploy/docker-compose.yml down
```

## Per-component compose (infra / partial install)

Files live in [`deploy/compose/`](compose/). Config trees: `compose/prometheus/`, `compose/grafana/`. Use project name **`wichat`** when starting stacks separately.

From repo root:

```bash
# Data stores
docker compose -p wichat -f deploy/compose/postgres.yml up -d
docker compose -p wichat -f deploy/compose/redis.yml up -d

# Observability (both files, or one Kuma service per file — start Prometheus first)
docker compose -p wichat -f deploy/compose/prometheus.yml -f deploy/compose/grafana.yml up -d
```

Kuma (Shell), working directory = repo root:

```bash
# Prometheus
docker compose -p wichat -f deploy/compose/prometheus.yml up

# Grafana (after Prometheus is up, same -p wichat)
docker compose -p wichat -f deploy/compose/grafana.yml up
```

Stop one component (example):

```bash
docker compose -p wichat -f deploy/compose/redis.yml down
docker compose -p wichat -f deploy/compose/prometheus.yml down
docker compose -p wichat -f deploy/compose/grafana.yml down
```

Postgres data volume: `wichat_wichat-postgres` (Compose prefixes volume names with the project name).

### Local ports (defaults)

| Service | Host port | Notes |
|---------|-----------|--------|
| Postgres | **5432** | `postgres://wichat:wichat@localhost:5432/wichat` |
| Redis | **6379** | `redis://localhost:6379/0` |
| Prometheus | **9090** | UI + API |
| Grafana | **9000** | Host maps to container **3000** |
| api-gateway | **3000** | Public HTTP |
| ms-user | **3001** | gRPC (auth + user/workspace) |
| ms-chat | **3002** | gRPC (messages, WS fanout, search indexer) — planned |
| ms-core | **3003** | gRPC (LiveKit tokens, call lifecycle) — planned |
| ms-plugin | **3004** | gRPC — planned |
| ms-notification | **3005** | gRPC — planned |
| ms-audit | **3006** | gRPC — planned |

Prometheus/Grafana containers may be present in Compose for future use; Go services do not expose `/metrics` yet.

Instance env: [`deploy/.env.example`](.env.example) (Compose overrides). Go services: each module’s `backend/<service>/.env.example`. See [docs/architecture/monorepo-layout.md](../docs/architecture/monorepo-layout.md).

See `docs/architecture/wichat-system-design.md` §7 and §11.

### Go services (local, not Docker)

From `backend/`: `make dev-ms-user` then `make dev-api-gateway` (Air live reload), or `make run-*` without reload. Kuma Shell: working directory `backend/ms-user` or `backend/api-gateway`, command `make dev`.
