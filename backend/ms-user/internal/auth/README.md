# auth (ms-user)

Auth domain in the same process as user/workspace: `repository/` → `service/` → `delivery/grpc/` when implemented.

Covers login, OAuth, sessions, refresh rotation, first-run admin, guest tokens (per system design).

Domain types live in [`internal/models/auth_model.go`](../models/auth_model.go). Bun table mapping: [`internal/models/auth_repository_model.go`](../models/auth_repository_model.go).

See `wichat-go-layers.mdc` and `docs/architecture/wichat-system-design.md`.

## Service

Use cases live in a single [`service/service.go`](service/service.go) (`Service` interface + implementation). Crypto and OAuth helpers: [`pkg/password`](../../pkg/password), [`pkg/jwt`](../../pkg/jwt), [`pkg/token`](../../pkg/token), [`pkg/google`](../../pkg/google). Tests: `service/test/`.

## Delivery (gRPC)

[`delivery/grpc/server.go`](delivery/grpc/server.go): `Server` interface + thin handlers (Bind / Service / Response). Proto mapping: [`pkg/auth-pb`](../../pkg/auth-pb).

## Repository layout (template for other `ms-*` domains)

Under `internal/<domain>/repository/`:

| File | Role |
|------|------|
| `pg_repository.go` | Sentinel errors, `PgRepository` interface, `NewPgRepository`, Postgres I/O |
| `sql_queries.go` | Advanced SQL only (locks, `FOR UPDATE`, joins, bulk deletes) |
| `test/` | Black-box integration tests (`package repository_test`) |

**Hybrid:** [Bun](https://bun.uptrace.dev/) on `database/sql` + pgx for simple selects/inserts/updates; raw SQL via `tx.NewRaw` for the rest. Wire `bun.NewDB(sql.Open("pgx", url), pgdialect.New())` in `server_run.go`.

`type Repository = PgRepository` keeps service wiring stable. Add `redis_repository.go` only when the domain has a real Redis use case.

Repository tests use Docker (testcontainers Postgres 16 + goose on `internal/migrations`). Without Docker they skip.
