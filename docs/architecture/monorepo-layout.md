# Monorepo layout & ownership

One git repository; **clear boundaries** so env, build artifacts, and docs do not mix between deploy, backend modules, and frontend.

## Top-level map

| Path | Owns | Does not own |
|------|------|----------------|
| `deploy/` | Docker Compose, infra ports, optional compose env | Go service config, app secrets in repo root |
| `backend/wi-shared/` | Shared Go libs, protobuf contracts | Runnable `.env`, service ports |
| `backend/api-gateway/` | Public HTTP, gateway `.env` | Postgres credentials (unless documented for local dev) |
| `backend/ms-user/` | Domain service config & `.env` | Frontend API URLs |
| `frontend/` | pnpm workspace root (`package.json`, `pnpm-lock.yaml`, Husky via `prepare`) | Backend Go modules |
| `frontend/apps/web/` | Next.js app (`web-client`), `NEXT_PUBLIC_*`, Node artifacts | Backend gRPC, DB URLs (except public API base URL) |
| `frontend/packages/*` | `@wichat/contracts`, `@wichat/api-client`, `@wichat/i18n` | Next env, UI components |
| `docs/` | Product & architecture spec | Runtime config |
| `references/` | Local clones for patterns (gitignored) | Shipped product code |

## Backend module paths

Go workspace: [`backend/go.work`](../../backend/go.work). Each runnable or library module is a **top-level directory under `backend/`** (not `backend/services/`):

| Path | Module role |
|------|-------------|
| `backend/api-gateway/` | Public HTTP `/api/v1`; thin delivery — routes to gRPC backends |
| `backend/ms-user/` | Auth + users/workspaces: domains under `internal/auth`, `internal/user`, …; shared types in `internal/models/`; SQL in `internal/migrations/` |
| `backend/ms-chat/`, `ms-core/`, … | Planned services; empty placeholder dirs until the slice lands |
| `backend/wi-shared/` | Protobuf (`models/`), errors, response helpers, gRPC client/server utilities |

There is **no** separate `ms-auth` service — auth lives in **`ms-user`**. Gateway must not import another service’s Go packages; only **`wi-shared`** generated gRPC clients.

**ms-user internals (pattern for other `ms-*`):**

```text
internal/models/           # auth_model.go, auth_repository_model.go, …
internal/migrations/       # goose per service
internal/<domain>/
  repository/              # pg_repository.go, sql_queries.go, test/
  service/
  delivery/grpc/
internal/server/           # Run(), bun + pgx pool wiring
```

Domain packages do **not** own a `model/` subfolder — types live in `internal/models/`.

**Helper placement:** `ms-user/pkg/` for code reused within that service; `wi-shared/` when two or more backend modules need it (proto, errors, HTTP/gRPC writers).

## Environment files

**Rule:** no aggregate `.env.example` at repo root. Each **runnable** unit documents its own variables.

| Location | File | Purpose |
|----------|------|---------|
| `deploy/` | `.env.example` | Optional Compose overrides (infra); connection hints for local dev |
| `backend/api-gateway/` | `.env.example` | `PORT`, `MS_*_HOST`, log level |
| `backend/ms-user/` | `.env.example` | `PORT`, `DATABASE_URL`, auth/OAuth secrets |
| `frontend/apps/web/` | `.env.example` | `NEXT_PUBLIC_API_URL` (and other `NEXT_PUBLIC_*` only) |

**Local dev:** from the module directory, `cp .env.example .env` (Go) or `.env.local` (Next.js). Never commit `.env`.

**Frontend workspace:** [`frontend/package.json`](../../frontend/package.json) + `frontend/pnpm-lock.yaml` (no Node toolchain at repo root). Run `cd frontend && pnpm install` (installs Husky at git root). Quality: `pnpm --filter web-client check` from `frontend/`. Pre-commit: lint-staged on `apps/web/**` and Prettier on `packages/**`.

**Infra URLs** (Postgres/Redis defaults): documented in [deploy/README.md](../../deploy/README.md), not duplicated as a second root env file.

## `.gitignore` ladder

Git merges ignore rules from root downward.

1. **Root `.gitignore`** — repo-wide: `references/`, OS/editors, all `.env` files, `!.env.example`, generic Go/Node/mobile patterns.
2. **`backend/.gitignore`** — Air `tmp/`, workspace-wide backend noise.
3. **Per module** (`api-gateway/`, `ms-user/`, `wi-shared/`, `frontend/apps/web/`) — short file stating **this module’s** local overrides and build dirs (even when root already covers them, for on-call clarity).

Do not put service-specific secrets or `node_modules` policy only in root without a module stub when that module exists.

## CI & PR scope

- Backend change → `cd backend/<module> && make check` (see [CONTRIBUTING.md](../../CONTRIBUTING.md)).
- Frontend change → `cd frontend && pnpm --filter web-client check` when `frontend/apps/web` or `frontend/packages` change.
- Compose / infra → `deploy/`.

Keep PRs **one concern** to reduce merge noise (see `wichat-program.mdc`).

## Future split (optional)

Go modules are already separate under `backend/go.work`. Extracting to multiple git repos later — not required for MVP:

1. Publish **`wi-shared`** (semver tag or private module proxy).
2. Publish each **`ms-*`** / **`api-gateway`** repo; depend on a pinned `wi-shared` version (drop `replace => ../wi-shared` in `go.mod`).
3. Copy deploy/CI docs per repo.

Services must keep importing **`github.com/wichat/wichat/backend/wi-shared`** (or the published module path after split), never `../ms-user/internal/...`.

**Frontend packages** (`frontend/packages/*`) are each a candidate git repo. Split order:

1. Publish **`@wichat/contracts`** (semver tag or private registry).
2. Publish **`@wichat/api-client`** and **`@wichat/i18n`** (depend on contracts version).
3. Point **`apps/web`**, future **`apps/mobile`**, etc. at published versions instead of `workspace:*`.

Until then, apps depend on packages only via **`import … from "@wichat/…"`** and **`"workspace:*"`** in `package.json` — no `../../packages` paths. Add `dist/` builds when a consumer lives outside this workspace.

## See also

- [wichat-system-design.md](wichat-system-design.md) §5.1 (services)
- [frontend/README.md](../../frontend/README.md)
- `.cursor/rules/wichat-web-*.mdc` (frontend boundaries)
