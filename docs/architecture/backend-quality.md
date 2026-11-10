# Backend quality — layout & coverage

## Layout

`make -C backend check-layout` runs [`backend/tools/check-layout`](../../backend/tools/check-layout) and enforces:

- `go.mod` module path `github.com/wichat/wichat/backend/<module>`
- `internal/<domain>/repository/` — only `pg_repository.go`, `minio_repository.go`, `es_repository.go`, `redis_*.go`, plus `mock/` and `test/`
- Every `pg_repository.go` has a matching `mock/pg_repository_mock.go`
- `pkg/` role dirs use **kebab-case** (`auth-pb`, `test-support`, …) with **camelCase** Go package names
- `*_test.go` only under a **`test/`** subdirectory (never beside production `.go`), folder name **`test/`** not `tests/`
- **Test 1:1:** each production `foo.go` has exactly `test/foo_test.go` in that package (no `auth_more_test.go`, `mock_test.go`, …)

| Bad | Good |
|-----|------|
| `service/test/service_test.go` + `service/test/auth_edges_test.go` | `service/test/service_test.go` only |
| `mock/test/mock_test.go` for `pg_repository_mock.go` | `mock/test/pg_repository_mock_test.go` |
| `recovery_test.go` beside `server.go` | `test/recovery_test.go` |

**No exception:** every `*_test.go` lives under `test/` (including no `export_test.go` beside production). Prefer public API or shared `pkg/` for logic that needs direct unit tests.

Authoritative rules: `.cursor/rules/wichat-go-layers.mdc`, `wichat-go-style.mdc`.

## Coverage (target 90%)

Scoped coverage counts **unit-testable** code:

| Module | Scope |
|--------|--------|
| `ms-*` | `internal/.../service`, `pkg/...`, `internal/.../repository/mock` |
| `wi-shared` | See [`wi-shared/coverage.scope`](../../backend/wi-shared/coverage.scope) |
| `api-gateway` | See [`api-gateway/coverage.scope`](../../backend/api-gateway/coverage.scope) |

Excluded from the gate (integration / wiring): `cmd/`, `internal/server`, `delivery/`, `repository/` (Postgres/ES), `migrations/`, `models/gen`.

Each module has **`.coverage-min`** — the CI/husky **floor** (ratchet upward). **Target** remains **90%**; bump `.coverage-min` whenever coverage improves so the floor never slips.

```bash
cd backend/<module> && make test-cover   # single module
cd backend && make check                 # layout + all modules
```

## Hooks

- **pre-commit** (frontend): lint-staged
- **pre-push** (repo): `backend` layout + `make -C backend check` when Go files changed

## Redis mocks

Do not add `repository/mock` for Redis — use **miniredis** + production `redis_*.go` (`wichat-go-layers.mdc`).
