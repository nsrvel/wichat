# AGENTS.md — WiChat

Guidance for humans and coding agents working in this repository.

## Product

WiChat is a **self-hosted**, **AGPL-3.0** team communication platform with a **plugin architecture** — a **major product** built for real daily use. Operability, security, and deployability are part of the application, not optional polish.

**Canonical spec:** [docs/architecture/wichat-system-design.md](docs/architecture/wichat-system-design.md)

## Architecture constraints (do not drift)

- Core backend: **Go monorepo microservices** under `backend/` — `api-gateway/`, `ms-<domain>/`, `wi-shared/` (gRPC + Kafka between services).
- Separate processes: **LiveKit**, **plugins (gRPC)**, not core domains.
- Data: PostgreSQL, Redis, Elasticsearch, MinIO, Kafka.
- Clients: Next.js (web + iOS PWA), Electron, React Native (Android).
- **Non-goals:** multi-tenant public SaaS, federation, E2E encryption v1, public plugin marketplace v1.

## Repository layout (target)

```
backend/       # Go: api-gateway/, ms-*/, wi-shared/, plugins/ (plugin servers, later)
frontend/      # apps/web (Next.js) + packages/@wichat/*
deploy/        # docker compose, env templates
docs/          # architecture & ops
```

Monorepo boundaries: [docs/architecture/monorepo-layout.md](docs/architecture/monorepo-layout.md).

**Go services (`ms-*`):** repository → service → delivery; models and migrations per `wichat-go-layers.mdc`. Shared backend libs: [backend/wi-shared/README.md](backend/wi-shared/README.md). **ms-user:** split Postgres repos (`internal/user/repository` vs `internal/auth/repository`); domain models in `internal/models`, constants in `pkg/constants` (not in repository/service/delivery); `pkg/db` (`pgErr`); no cross-package re-exports (`wichat-go.mdc`). **Tests:** Postgres in-memory mocks in `repository/mock/` only; Redis stores tested with miniredis + real repo code, not `repository/mock`. **Service:** one `service.go` + `service/test/service_test.go` per domain package; shared test helpers in `pkg/test-support` (`testSupport`).

Phases in system design §15 — implement in order unless explicitly reprioritized.

## Quality bar (production-ready)

- Security: TLS, rate limits, no secrets in repo, bcrypt/argon2, guest/session scoping per spec.
- Tests: Go unit (~90% goal on core), testcontainers for integration, Playwright for critical E2E paths.
- Observability: Grafana + Loki (when services exist).
- API: REST `/api/v1/`, WebSocket for realtime; UTC timestamps; i18n keys from backend.
- Keep core **light** — heavy features belong in plugins.

## Legal

License: [LICENSE](LICENSE) (AGPL-3.0-or-later at repo root; no per-file SPDX/copyright headers). Do not add proprietary-licensed dependencies without explicit approval.

## Cursor rules

**Always-on:** program + core. **File-scoped:** Go and web when touching those paths.

| Rule                           | Scope                                                                                               |
| ------------------------------ | --------------------------------------------------------------------------------------------------- |
| `wichat-program.mdc`           | How we work: `main` (prod) / `release` (staging), micro-steps, vertical slices                      |
| `wichat-core.mdc`              | Product + architecture guardrails                                                                   |
| `wichat-go.mdc`                | Go backend: tests under `<pkg>/test/`, per-module `make check`, wi-shared patterns                  |
| `wichat-go-layers.mdc`         | Go backend: repository / service / delivery layout                                                  |
| `wichat-go-style.mdc`          | Go backend: `//` step comments, wiring spacing                                                      |
| `wichat-web-layout.mdc`        | `frontend/apps/web/**` — app vs features, components, no micro-FE                                   |
| `wichat-web-data.mdc`          | `frontend/apps/web/**` + `frontend/packages/**` — `@wichat/*`, TanStack Query services, WS vs Query |
| `wichat-no-export-barrels.mdc` | `frontend/**` — no re-export-only `index.ts`; import defining modules directly                      |

**Skill:** `.cursor/skills/wichat-web-client/` — scaffold and feature slices for web-client.

Detail standards: [system design](docs/architecture/wichat-system-design.md) and this file.
