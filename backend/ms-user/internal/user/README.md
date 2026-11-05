# user (ms-user)

User/workspace domain in the ms-user process: `repository/` → `service/` → `delivery/grpc/` when implemented.

Auth lives in sibling package `internal/auth/`. Postgres users, groups, and instance policy: [`repository/pg_repository.go`](repository/pg_repository.go). Domain types today: [`internal/models/auth_model.go`](../models/auth_model.go) and [`auth_repository_model.go`](../models/auth_repository_model.go) — see [`internal/models/README.md`](../models/README.md).

## Repository

| File | Role |
|------|------|
| `pg_repository.go` | `PgRepository` — users, identities (Google), groups, instance settings |
| `sql_queries.go` | Locks and identity join SQL |
| `mock/` | In-memory `PgRepository` (`pg_repository_mock.go`, `State`); shared with auth session mock |

**File layout:** interface + constructor → all interface methods (same order) → private helpers at bottom. Driver errors via [`pkg/db`](../../pkg/db) (`pgErr`); sentinels via [`pkg/constants`](../../pkg/constants) (no re-export shims under `internal/`).

## Service

[`service/service.go`](service/service.go) — instance setup, guest user creation, setup status. Tests: [`service/test/service_test.go`](service/test/service_test.go).

Auth service injects `user/service.Service` for guest registration and username policy; gRPC setup RPCs call user service directly.

See `wichat-go-layers.mdc` and `docs/architecture/wichat-system-design.md`.
