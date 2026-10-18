# Frontend

pnpm workspace under **`frontend/`**:

| Path | Role |
|------|------|
| `apps/web/` | Next.js App Router (package name `web-client`, PWA-capable) |
| `packages/contracts/` | `@wichat/contracts` — Zod + types (API shapes) |
| `packages/api-client/` | `@wichat/api-client` — HTTP client factory + `ApiError` |
| `packages/i18n/` | `@wichat/i18n` — translation resources + `I18nProvider` |

Planned later (not in repo yet): `apps/desktop` (Electron shell), `apps/mobile` (Expo), `packages/tokens` when NativeWind needs shared numbers with web CSS.

Cross-app imports use **`@wichat/*`** and `workspace:*` only — no relative paths into `packages/`. See [monorepo-layout.md](../docs/architecture/monorepo-layout.md) for optional multi-repo split.

Copy-out starter: [`templates/personal-dashboard/`](../templates/personal-dashboard/).

Work **one vertical slice at a time**. UI-only slices with MSW/fixtures are OK if Zod shapes match **api-gateway** `/api/v1/`.

## Layout (`apps/web/src/`)

```
app/              # routing + layouts only (thin pages)
features/         # domain screens & feature-local components
components/ui/    # shadcn
components/layout/# shared shell
services/         # *.key | *.service | *.query | *.mutation
integrations/     # app wiring (api singleton, tanstack-query, locale cookie)
hooks/            # generic UI hooks
lib/
```

Shared API/i18n contracts live in **`frontend/packages/`**, not under `apps/web/src/schemas/`.

Agent rules: `.cursor/rules/wichat-web-layout.mdc`, `.cursor/rules/wichat-web-data.mdc`. Skill: `.cursor/skills/wichat-web-client/`.

Env: `cp apps/web/.env.example` → `apps/web/.env.local` (from `frontend/`). Install and dev:

```bash
cd frontend && pnpm install && pnpm --filter web-client dev
```

UI **8080**, API **3000**.

Pattern reference (not runtime): `references/react-tanstack/`.
