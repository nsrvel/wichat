# GEMINI.md — WiChat

Workspace-level guidance and constraints for Google Antigravity / Gemini CLI.

Canonical source of truth and architecture constraints: [AGENTS.md](AGENTS.md) and [docs/architecture/wichat-system-design.md](docs/architecture/wichat-system-design.md).

## Quick Overview

WiChat is a self-hosted, AGPL-3.0 team communication platform with a plugin architecture. Core chat stays lightweight; task boards, whiteboards, and spatial office features ship as plugins.

- Core backend: Go monorepo microservices (`backend/` with `api-gateway/`, `ms-*/`, `wi-shared/` using gRPC + Kafka).
- Core frontend: Next.js (`frontend/apps/web`) with modular packages (`frontend/packages/@wichat/*`).
- Quality gates: Strict Go layering (`repository` -> `service` -> `delivery`), `test/` package separation, scoped 90% test coverage floor (`make check`), no re-export barrel files.

## Active Rules & Skills

Native Antigravity rules and skills are located in `.agents/`:
- **Workspace Rules:** `.agents/rules/*.md` (Core, Program sequencing, Folder grouping, Go layering, Backend quality, Web layout/data, Barrels, UI taste).
- **Workspace Skills:** `.agents/skills/wichat-web-client/` and `.agents/skills/design-taste-frontend/`.

Follow all constraints specified in [AGENTS.md](AGENTS.md) at all times.

