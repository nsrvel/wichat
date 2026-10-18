# WiChat web client (Next.js)

From **`frontend/`** (pnpm workspace + Husky at repo root):

```bash
cd frontend
pnpm install
cp apps/web/.env.example apps/web/.env.local
pnpm --filter web-client dev
```

Or from this directory: `pnpm dev` (after `pnpm install` in `frontend/`).

Dev server: **http://localhost:8080**. **api-gateway** **:3000** (`NEXT_PUBLIC_API_URL`); **ms-user** **:3001** (gRPC, not for browser).

**Quality gate before PR:**

```bash
cd frontend && pnpm --filter web-client check
# or from this folder: make check
```

Pre-commit runs lint-staged from `frontend/` (ESLint + Prettier on staged app/package files).

Conventions: [frontend/README.md](../../README.md), `.cursor/rules/wichat-web-*.mdc`.
