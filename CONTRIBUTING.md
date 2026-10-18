# Contributing to WiChat

Thank you for helping make WiChat production-grade. This project is pre-implementation in places; when in doubt, follow [docs/architecture/wichat-system-design.md](docs/architecture/wichat-system-design.md).

## Branches

- **`release`** — active development (default branch; open PRs here).
- **`main`** — stable line; updated from `release` when we cut a shipped version (not day-to-day commits).

## Before you start

1. Open an issue for non-trivial work (feature, breaking change, new dependency).
2. Keep PRs focused — one concern per change when possible.
3. AGPL-3.0 applies to contributions; you agree your contributions are licensed under the same terms.

## Development setup

- **Go 1.26** (see `backend/go.work`). Each deployable module (`wi-shared`, `ms-user`, `api-gateway`) is self-contained for a future split into separate repos.
- Checks (monorepo orchestrator or per module):

```bash
cd backend/ms-user && make check      # fmt + lint + test (same for api-gateway, wi-shared)
cd backend/wi-shared && make proto    # regenerate protobuf (requires buf)
```

- Services depend on **wi-shared** (errors, config, and `models/gen` gRPC types). When repos split, use a tagged wi-shared module instead of `replace` in `go.mod`.

- **Environment:** no root `.env.example`. Use `deploy/.env.example` for Compose overrides; copy `backend/<service>/.env.example` → `.env` inside that service; `frontend/apps/web/.env.example` → `.env.local` for Next.js. Details: [docs/architecture/monorepo-layout.md](docs/architecture/monorepo-layout.md).

- Optional editor: enable format on save with **gofumpt** (gopls `go.formatTool`).

**Frontend** (`frontend/` pnpm workspace — `apps/web`, `packages/@wichat/*`): Node **22+**, pnpm **10**.

```bash
cd frontend && pnpm install    # Husky hooks at repo root
cp apps/web/.env.example apps/web/.env.local
pnpm --filter web-client check
```

Pre-commit runs lint-staged from `frontend/` (ESLint/Prettier on staged app and package files).

**Local reference clones:** optional pattern repos under `references/` are gitignored and not shipped; delete that folder anytime to reclaim disk (~1GB+ if you cloned examples).

Infra: `docker compose -f deploy/docker-compose.yml up -d`. Architecture docs remain the source of truth for product behavior.

## Code standards

Follow [system design](docs/architecture/wichat-system-design.md) and `.cursor/rules/wichat-*.mdc`. Narrower Cursor rules are added only when matching code lands (Go services, web app, etc.).

## Pull requests

- Describe **why**, not only what.
- Link issues (`Fixes #123`) when applicable.
- Note test evidence (unit, integration, or manual steps).
- Update docs if behavior or ops change.

## Commits

Use clear, imperative subjects (`Add workspace registration modes`). Squash or merge per maintainer preference on the PR.

## Plugin authors (in-house v1)

Plugins run as separate processes with declared permission manifests; see system design §6.8 and §6.2.
