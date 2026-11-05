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

Files live in [`deploy/compose/`](compose/). Use project name **`wichat`** when starting stacks separately.

From repo root:

```bash
# Data stores
docker compose -p wichat -f deploy/compose/postgres.yml up -d
docker compose -p wichat -f deploy/compose/redis.yml up -d
docker compose -p wichat -f deploy/compose/minio.yml up -d
```

Stop one component (example):

```bash
docker compose -p wichat -f deploy/compose/redis.yml down
```

Postgres data volume: `wichat_wichat-postgres` (Compose prefixes volume names with the project name).

### Local ports (defaults)

| Service | Host port | Notes |
|---------|-----------|--------|
| Postgres | **5432** | `postgres://wichat:wichat@localhost:5432/wichat` |
| Redis | **6379** | `redis://localhost:6379/0` |
| MinIO | **9000** (API), **9001** (console) | optional — group icons when `MINIO_*` set on ms-user |
| api-gateway | **3000** | Public HTTP |
| ms-user | **3001** | gRPC (auth + user/workspace) |
| ms-chat | **3002** | gRPC (messages, WS fanout, search indexer) — planned |
| ms-core | **3003** | gRPC (LiveKit tokens, call lifecycle) — planned |
| ms-plugin | **3004** | gRPC — planned |
| ms-notification | **3005** | gRPC — planned |
| ms-audit | **3006** | gRPC — planned |

Prometheus/Grafana are not in Compose until Go services expose metrics (see `docs/architecture/wichat-system-design.md` observability).

Instance env: [`deploy/.env.example`](.env.example) (Compose overrides). Go services: each module’s `backend/<service>/.env.example`. See [docs/architecture/monorepo-layout.md](../docs/architecture/monorepo-layout.md).

See `docs/architecture/wichat-system-design.md` §7 and §11.

### Go services (local, not Docker)

From `backend/`: `make dev-ms-user` then `make dev-api-gateway` (Air live reload), or `make run-*` without reload. Kuma Shell: working directory `backend/ms-user` or `backend/api-gateway`, command `make dev`.
