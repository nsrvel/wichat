---
name: wichat-web-client
description: >-
  Scaffold or extend WiChat Next.js app at frontend/apps/web — thin app/
  routes, features/, services query/mutation pattern, shadcn, api-gateway
  client. Use when adding frontend slices, auth/chat UI, or initializing
  web-client.
---

# WiChat web client

Read before implementing under `frontend/apps/web/` (shared packages: `frontend/packages/@wichat/*`):

1. **Rules:** `.cursor/rules/wichat-web-layout.mdc`, `.cursor/rules/wichat-web-data.mdc`
2. **Spec:** `docs/architecture/wichat-system-design.md` (clients, REST, WS, plugins iframe)
3. **Pattern reference:** `references/react-tanstack/src/services/` (key/service/query/mutation only)
4. **Templates:** [examples.md](examples.md) in this skill folder

## Scaffold checklist (first slice)

1. Next.js App Router + TypeScript + Tailwind under `frontend/apps/web/`
2. shadcn → `src/components/ui/`; `next-themes` in `app/layout.tsx`
3. `@/` alias → `src/`
4. `integrations/api-client`, `integrations/tanstack-query` (`get-query-client.ts`, `providers.tsx`)
5. `env.ts` (`NEXT_PUBLIC_API_URL`)
6. One domain end-to-end in UI: e.g. `@wichat/contracts`, `services/auth/*`, `features/auth/login/login-page.tsx` + `login-form.tsx`, thin `app/(auth)/sign-in/page.tsx`
7. Optional: `integrations/msw` for mock until gateway routes exist

## New feature slice (repeatable)

1. Zod in `schemas/<domain>/`
2. `services/<domain>/*.service.ts` → api-client
3. `*.query.ts` / `*.mutation.ts` + `*.key.ts` (no re-export `index.ts` — see `wichat-no-export-barrels.mdc`)
4. Screen under `features/<domain>/<slice>/`
5. Wire route in `app/` only
6. Mutations invalidate the correct query keys

## Verify

- No fetch/SQL in `components/ui`
- Errors surfaced via i18n keys from `error.code`
- Commit-sized diff; update `frontend/README.md` if layout changes
