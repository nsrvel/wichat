# ms-user

Identity and platform (gRPC): **auth** (login, sessions, OAuth, guest) and **user/workspace** (profiles, membership, RBAC, channel metadata).

```bash
cp .env.example .env   # DATABASE_URL, AUTH_JWT_SECRET, …
make migrate-up        # goose — apply internal/migrations (before first run)
make dev    # Air live reload
make run    # single run
make check  # fmt, lint, test (repository integration tests need Docker)
```

## Layout

| Path | Role |
|------|------|
| `internal/models/` | Service-wide types: `<domain>_model.go` + `<domain>_repository_model.go` (Bun rows) |
| `internal/migrations/` | Goose SQL (not run on startup) |
| `internal/auth/`, `internal/user/` | Per domain: `repository/`, `service/` (single `service.go`), `delivery/grpc/` |
| `pkg/` | Shared helpers (`password`, `jwt`, `token`, `google`, `auth-pb`, …) — not business use cases |
| `internal/server/` | gRPC listen, DB (`bun` + pgx), wiring |

Postgres: **Bun** for simple CRUD, raw SQL in `repository/sql_queries.go` for locks / `FOR UPDATE` / joins. See [`internal/auth/README.md`](internal/auth/README.md) and [`internal/models/README.md`](internal/models/README.md).

Schema files: `internal/migrations/*.sql`. Goose tracks versions in `goose_db_version`.

Default listen: **`PORT=3001`** (gRPC).

From `backend/`: `make dev-ms-user`.
