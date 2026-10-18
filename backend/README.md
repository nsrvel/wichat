# Backend (Go workspace)

Monorepo **microservices** under one `go.work` — each module is self-contained (future split-friendly).

## Modules

| Directory | Role |
|-----------|------|
| `wi-shared/` | Config, errors, response, gRPC models, infra helpers |
| `ms-user/` | Auth + users / workspaces (gRPC) |
| `api-gateway/` | Public HTTP `/api/v1` |

## Commands (from `backend/`)

```bash
make migrate-up         # Postgres schema (goose, ms-user) — run before ms-user
make migrate-status
make check              # all modules
make dev-ms-user        # Air reload
make dev-api-gateway
```

Per module: `cd ms-user && make dev` — copy `.env.example` → `.env` in **that** directory.

## Database migrations

Each service owns SQL under `internal/migrations/` (e.g. [`ms-user/internal/migrations`](ms-user/internal/migrations)). Services do **not** apply migrations on startup. Use [goose](https://github.com/pressly/goose) via `make migrate-up` from `backend/` or `cd ms-user && make migrate-up` (`DATABASE_URL` from `ms-user/.env`).

## Default gRPC/HTTP ports (local dev)

| Port | Service |
|------|---------|
| 3000 | api-gateway (HTTP) |
| 3001 | ms-user |
| 3002–3006 | ms-chat, ms-core, ms-plugin, ms-notification, ms-audit (planned) |

Full table: [deploy/README.md](../deploy/README.md#local-ports-defaults).

## Config

- **No** shared backend `.env` at repo root.
- Env templates: each module’s `.env.example`.
- Infra defaults: [deploy/README.md](../deploy/README.md).

Layout: [docs/architecture/monorepo-layout.md](../docs/architecture/monorepo-layout.md).
