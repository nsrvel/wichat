# auth (ms-user)

Auth domain in the same process as user/workspace: `repository/` → `service/` → `delivery/grpc/` when implemented.

Covers login, OAuth, sessions, refresh rotation, first-run admin, guest tokens (per system design).

Domain types live in [`internal/models/auth_model.go`](../models/auth_model.go). Bun table mapping: [`internal/models/auth_repository_model.go`](../models/auth_repository_model.go).

See `wichat-go-layers.mdc` and `docs/architecture/wichat-system-design.md`.

## Service

Use cases in [`service/service.go`](service/service.go) (`Service` interface + orchestration). **Dependencies:**

- `user/repository.PgRepository` → `s.users` (accounts, groups, instance policy)
- `auth/repository.PgRepository` → `s.sessions` (session + refresh rows)
- Redis: `PasswordResetStore`, `EmailAuthStore` from `auth/repository`

```go
userSvc := userservice.NewService(userPg, authPg, secret, log)
groupSvc := groupservice.NewService(userPg, authPg, secret, log)
authservice.NewService(userPg, authPg, secret, log, googleCfg, resetStore, emailAuthStore, notify, userSvc)
```

Persistence errors: `errors.Is(err, constants.ErrNotFound)` from [`pkg/constants`](../../pkg/constants). Helpers: [`pkg/password`](../../pkg/password), [`pkg/jwt`](../../pkg/jwt), [`pkg/token`](../../pkg/token), [`pkg/google`](../../pkg/google).

**Tests:** [`service/test/service_test.go`](service/test/service_test.go) only; wiring via [`pkg/test-support`](../../pkg/test-support). Postgres mocks: [`user/repository/mock`](../user/repository/mock) + [`auth/repository/mock`](repository/mock). Redis: miniredis + real `redis_*` stores.

**Related services:** [`user/service`](../user/service/service.go) (setup, guest), [`group/service`](../group/service/service.go) (membership) — same gRPC `AuthService` delivery.

## Delivery (gRPC)

[`delivery/grpc/server.go`](delivery/grpc/server.go): `Server` interface + thin handlers (Bind / Service / Response). Proto mapping: [`pkg/converter`](../../pkg/converter) (`authPb`).

## Repository layout (template for other `ms-*` domains)

Under `internal/<domain>/repository/`:

| File | Role |
|------|------|
| `pg_repository.go` | `PgRepository` — sessions and refresh tokens only |
| `sql_queries.go` | Session/refresh SQL (`FOR UPDATE`, bulk delete expired) |
| `redis_email_auth.go`, `redis_password_reset.go` | Redis OTP stores (`EmailAuthStore`, `PasswordResetStore`); JSON shapes in `internal/models`; shared crypto/backoff in `pkg/auth-redis` |
| `mock/` | In-memory `PgRepository` only (`pg_repository_mock.go`) — sessions delegate to shared state in `user/repository/mock` |
| `test/` | Integration tests (`repository_test`); Postgres testcontainers; Redis miniredis |

**Hybrid:** [Bun](https://bun.uptrace.dev/) on `database/sql` + pgx for simple selects/inserts/updates; raw SQL via `tx.NewRaw` for the rest. Wire `bun.NewDB(sql.Open("pgx", url), pgdialect.New())` in `server_run.go`.

User/group/instance Postgres I/O lives in `internal/user/repository/`. Auth service injects `user/repository.PgRepository` and `auth/repository.PgRepository` separately.

Repository tests (`repository/test/`): Postgres integration needs Docker (testcontainers + goose); skipped when unavailable. Redis stores use miniredis (no Docker).
