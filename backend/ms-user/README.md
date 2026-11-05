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
| `internal/auth/` | Auth use cases + gRPC; `repository/` = sessions/refresh + Redis challenges |
| `internal/user/` | User/workspace repo + [`service/service.go`](internal/user/service/service.go) (setup, guest) |
| `internal/group/` | Group membership [`service/service.go`](internal/group/service/service.go) |
| `internal/general/` | Operational gRPC only (e.g. health liveness) — see [`internal/general/README.md`](internal/general/README.md) |
| `pkg/` | Grouped by role (`auth/`, `validator/`, `converter/`, `utils/`, …) — see [`pkg/README.md`](pkg/README.md) |
| `internal/server/` | gRPC listen, DB (`bun` + pgx), wiring |

Postgres: **Bun** for simple CRUD, raw SQL in each domain’s `repository/sql_queries.go`. Wire **two** repos from one DB in `server_run.go`: `user/repository.NewPgRepository(db)` + `auth/repository.NewPgRepository(db)`. See [`internal/auth/README.md`](internal/auth/README.md), [`internal/user/README.md`](internal/user/README.md), [`pkg/README.md`](pkg/README.md).

**Testing:** `repository/mock/pg_repository_mock.go` for fast service tests (no Postgres). Redis OTP stores — **miniredis** + `NewRedisEmailAuthStore` / `NewRedisPasswordResetStore`, not `repository/mock`.

Schema files: `internal/migrations/*.sql`. Goose tracks versions in `goose_db_version`.

Default listen: **`PORT=3001`** (gRPC).

From `backend/`: `make dev-ms-user`.
