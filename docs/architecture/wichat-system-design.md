# WiChat — System Design Document

**Version:** 1.0 (Draft)
**Status:** Pre-development / Architecture finalized
**Owner:** Garam

---

## 1. Executive Summary

WiChat is a self-hosted, plugin-based team communication platform positioned between casual tools (Discord, WhatsApp) and rigid enterprise tools (Microsoft Teams, Lark, Mattermost). It combines Discord-like flat channel simplicity with Teams-like professional UX, built around a plugin architecture where features beyond core chat (task boards, whiteboards, virtual office) are optional modules rather than baked-in bloat.

The project originated from a real gap: an internal team found existing tools either too rigid/corporate (Mattermost, Rocket.Chat) or too casual (Discord, WhatsApp) for a small, inconsistently-active team. WiChat is not the primary business — it is a side project / portfolio piece that the team will use regardless of commercial outcome, with self-hosting as its strongest differentiator (data sovereignty, cost control, full customization) against open-source incumbents.

## 2. Goals

- Deliver a **self-hosted**, single-tenant-per-instance, multi-workspace communication platform
- Support **2–500 people per workspace** without requiring clustering or sharding
- Ship a **plugin architecture** from day one, validated with dummy/stub plugins before real plugins are built
- Keep the **core lightweight** — spatial/heavy features (virtual office) are optional, not load-bearing
- Reach a working **internal MVP** usable by the founding team within a short timeframe

## 3. Non-Goals (for now)

- Multi-tenant SaaS model (à la Discord's public server model) — explicitly out of scope
- Public plugin marketplace / third-party plugin developers — in-house only for v1
- End-to-end encryption — explicitly deferred; TLS + at-rest encryption is the baseline
- Native iOS app — PWA only, to avoid Apple Developer Program costs
- 2FA — deferred to a later phase
- Bot/webhook integration ecosystem — deferred, separate concern from the plugin system
- Custom role builder UI — the 4 preset roles (Owner/Admin/Member/Guest) cover the 2–500-person target; the underlying RBAC schema (§6.2) is designed to support this later without a breaking migration, but the UI itself is not built for v1
- **Cross-instance federation** (Matrix/ActivityPub-style) — explicitly rejected, not just deferred: it contradicts WiChat's core value proposition of per-organization data sovereignty and isolation. No mainstream closed team-chat product (Slack, Teams, Mattermost, Rocket.Chat) federates, for the same reason.

## 4. Product Positioning

| Axis | WiChat | Discord/WhatsApp | Mattermost/Rocket.Chat | MS Teams/Lark |
|---|---|---|---|---|
| Tone | Semi-professional | Casual | Corporate | Corporate |
| Self-hosted | Yes (core value) | No | Yes | No |
| Plugin model | Modular, opt-in | N/A | Plugin system exists, heavier | App integrations |
| Target user | Small teams/communities wanting structure without corporate rigidity | Communities | Enterprises wanting compliance | Enterprises |

## 5. High-Level Architecture

See [`architecture-context.mermaid`](architecture-context.mermaid) for the C4 Context-level view (users, WiChat, external systems) and [`architecture-container.mermaid`](architecture-container.mermaid) for the Container-level view (internal services, data stores, plugins). Summary of major building blocks:

- **Core Backend**: **monorepo microservices** — multiple Go binaries under `backend/` (`api-gateway`, `ms-*`, `wi-shared`; see §5.1)
- **Media Layer**: LiveKit (self-hosted SFU) as a genuinely separate process
- **Plugin Layer**: separate processes per plugin, communicating with platform/plugin services via gRPC; frontend plugin UI via iframe/WebView + postMessage
- **Data Layer**: PostgreSQL (primary store), Redis (ephemeral/real-time), Elasticsearch (search), MinIO (file storage)
- **Event Backbone**: Kafka for durable, replayable events (audit log, notifications, async search indexing, future analytics)
- **Clients**: Next.js web app (also serves as iOS PWA), Electron desktop (wraps web core), React Native Android app

### 5.1 Monorepo microservices (core backend)

**Decision:** Core runs as **separate deployable Go modules** in `backend/` (workspace [`backend/go.work`](../../backend/go.work)): **`api-gateway/`** (public HTTP) and **`ms-<domain>/`** gRPC services (e.g. `ms-user`, `ms-chat`), plus **`wi-shared/`** for protobuf and cross-cutting libs. Services talk **gRPC** internally and **Kafka** for async work.

**Why not one binary:** Keeps each runtime small and lets teams evolve/chat scale/realtime paths independently without one growing “god binary.”

**Tradeoff (explicit):** Self-host runs **more containers** than a monolith — mitigated by a single `docker compose` stack and shared infra (one Postgres cluster, one Redis, one Kafka).

**Service set (v1 — seven deployable core services; repo paths match module names under `backend/`):**

| Service | Responsibility |
|---------|----------------|
| **api-gateway** | Public HTTP `/api/v1`, rate limits, JWT validation (via ms-user), routes to gRPC backends, presigned URL issuance, export job trigger |
| **ms-user** | Auth (OAuth, email/password, sessions, refresh rotation, first-run admin, guest tokens when enabled) **and** workspaces, membership, RBAC, channels/categories/DM metadata, workspace settings, lobby/onboarding, user profiles — one gRPC process, separate domain packages (`internal/auth`, `internal/user`) |
| **ms-chat** | Messages, threads, reactions, pins, typing, read receipts, WebSocket fanout, presence/activity status, chat-side moderation signals, **and** message search (query API + Elasticsearch indexing via Kafka `chat-events` consumer, RBAC-scoped per §6.10) — one gRPC process, separate domain package `internal/search` when implemented |
| **ms-core** | LiveKit room tokens, webhooks, call lifecycle (realtime voice/video/screen share control — not auth, chat messages, or plugins) |
| **ms-plugin** | Plugin registry, install lifecycle, gRPC to plugin processes in `backend/plugins/` |
| **ms-notification** | @mention and notification dispatch (Kafka `notif-events` in → FCM/Web Push out) |
| **ms-audit** | Audit log writer (Kafka consumer → Postgres) |

**Default local listen ports** (gRPC unless noted): api-gateway **3000** (HTTP), ms-user **3001**, ms-chat **3002**, ms-core **3003**, ms-plugin **3004**, ms-notification **3005**, ms-audit **3006**. See [deploy/README.md](../../deploy/README.md).

**Data ownership:** One **PostgreSQL cluster** at launch; each service owns its **tables** (no cross-service SQL joins — use IDs + gRPC/events). Shared DB is a pragmatic compromise; split databases only if ops requires it later.

**Shared code:** `backend/wi-shared/` for protobuf contracts (`models/`), shared error types, and small utilities only — **not** shared domain logic (avoid distributed monolith).

**Clients:** live under `frontend/` (web first). Talk only to **api-gateway** (REST) and **ms-chat** (WebSocket via api-gateway or documented WS URL).

## 6. Core Domains

### 6.1 Auth & Identity
- Registration mode is workspace-configurable: `open` (community) / `invite_only` / `approval_required` (corporate)
- Login: Google OAuth + email/password (bcrypt/argon2 hashing). Generic enterprise SSO (SAML/OIDC against IdPs like Azure AD/Okta) is explicitly deferred — added only if/when a concrete enterprise client requests it, not built preemptively
- Guest access: Google Meet-style — join via link, no account, short-lived session (JWT + Redis TTL), scoped to a single channel, cannot see message history from before joining, upload quota capped (e.g. 5MB / 3 files)
- Multi-device login supported from the start; incoming calls ring all active devices, answering on one dismisses the others
- **Session management**: an `active_sessions` record (device info, refresh token ID, last-active time) per login, allowing an Admin to force-revoke a specific device/session (e.g. lost laptop, sudden employee departure) without affecting the user's other active sessions
- First-run: a "Create First Admin Account" form (not a full setup wizard) prevents reliance on a hardcoded default password
- 2FA and full E2E encryption are deferred
- **Account deletion**: same pattern as Discord/Slack — the account is anonymized ("Deleted User", generic avatar; email/password hash scrubbed permanently), but their past messages remain intact for conversation context rather than leaving gaps in channel history

### 6.2 Roles & Permissions (RBAC)
`Owner` (1, non-demotable safety role) → `Admin` (manage members/plugins/settings) → `Member` (standard user) → `Guest` (session-based, restricted) are the four preset roles, workspace-scoped.

Only Admins can create new channels. Private channels are supported for ad-hoc meetings.

**Data model (finalized as a foundation-level decision, not a v1 feature):** a custom role builder UI is explicitly deferred (out of scope for now, per §3 Non-Goals) — the 4 presets are enough for the 2–500-person target. However, the underlying schema is designed to support custom roles and plugin-defined permissions from day one, so this never requires a breaking migration later:

- `roles` table: workspace-scoped, `permissions TEXT[]`, `is_preset`/`is_system` flags protect the 4 built-in roles from deletion/edit, `position` for future ordering
- `permissions` table acts as a catalog/registry (`key`, `label`, `category`, optional `source_plugin_id`) — covers both core permissions and permissions a plugin registers for itself (e.g. `taskboard.delete_card`). Role-to-permission mapping stays as the `TEXT[]` array (validated against this catalog in the Go application layer, not a DB-level FK, since Postgres doesn't support per-element array foreign keys) rather than a normalized many-to-many join table, to keep the common "does this role have permission X" check a single indexed array lookup (GIN index) instead of an extra join
- `user_workspace_roles` uses a composite primary key (`user_id, workspace_id, role_id`) rather than a single `role_id` column on `users`, so multi-role-per-user is possible later without a schema change

Some rules layered on top of this RBAC base are attribute-based rather than pure role checks (e.g. guests can't see message history from before they joined; message hard-delete has a grace window tied to authorship/recency) — these are handled as conditional logic in the relevant module, not part of the role/permission table itself.

### 6.3 Workspace Model
- Single-tenant per instance, but **multi-workspace within an instance** — a user must be explicitly added to a workspace to access it
- Single domain with a workspace picker (Slack-style), not per-workspace subdomains
- Config is split: instance-level (env/file, set before run — DB, secrets, domain) vs. workspace-level (DB-stored, admin-editable — upload limits, registration mode, retention, branding), with workspace settings overriding instance defaults

### 6.4 Channels & Messaging
- Flat channel structure (Discord-style), not Teams' Team→Channel hierarchy, but with Teams-like professional UI
- **Unified channel**: no separate text/voice channel type — any channel can start an inline voice/video call
- Chat model: inline-first replies (quote banner) with expandable threads; schema supports both via `reply_to_message_id`, `root_message_id`, `reply_count`
- Rich text via TipTap/ProseMirror — Markdown shortcuts + floating toolbar
- Deletion: soft-delete when a message has replies (preserves thread context), hard-delete otherwise within a grace window; soft-deleted items sit in a "trash" state governed by the workspace's **trash grace period** (see §7.1), separate from the workspace's overall data retention window
- Offline reconnect: client tracks `last_synced_message_id` (UUIDv7), requests a `sync:missed` catch-up on reconnect
- Features: edit (visible "(edited)" label), reactions, pin, typing indicator, workspace custom emoji (core feature, tightly integrated with editor/search)
- Read receipts: simplified aggregate indicator (all-but-one read), not per-person tracking
- 1-on-1 DM supported

### 6.5 Voice & Video
- LiveKit (self-hosted SFU) for voice/video/screen share, started inline per channel
- Controls: Mute/Deafen (+ server-mute/deafen for admins), Push-to-Talk vs Voice Activity Detection
- Audio enhancement: basic AEC/noise suppression via built-in libwebrtc (free); optional self-hosted DTLN/RNNoise plugin later for higher quality. **Krisp is excluded** — it requires LiveKit Cloud or a paid license, not viable for self-hosted OSS.
- Soundboard: short clips (<5s) with rate limiting and per-user mute
- A single well-provisioned LiveKit node (4–8 vCPU) is expected to comfortably serve the 2–500-people-per-workspace target; multi-node clustering is deferred until load requires it

### 6.6 Presence
Status: available / busy / DND / BRB / appear-away / offline — automatic (activity-based) or manually set, with optional duration.

**Activity Status**: a general-purpose (not casual-only) auto-status layer on top of presence — e.g. joining a voice/video call automatically sets status to "In a call" / "In a meeting", overridable manually. Kept branding-neutral so it fits both community and corporate contexts. App-specific rich presence (e.g. "listening to X") is explicitly a **Client-Only Personal Plugin**, not core.

### 6.7 Channel Organization
- **Categories**: collapsible, non-hierarchical grouping in the sidebar (Slack "Sections" / Discord "Categories" style, not a Teams-style Team→Channel hierarchy). Data model: nullable `category_id` on channel; uncategorized channels surface at the top.
- **Archiving**: channels are archived, not hard-deleted — consistent with the message soft-delete philosophy. An archived channel disappears from the active sidebar, history remains intact and restorable by an admin, but no new messages can be sent to it.

### 6.8 Plugin System
Two-tier taxonomy:
1. **Workspace/Server Plugins** (admin-installed, shared): backend via gRPC service, frontend via iframe/WebView + postMessage, surfaced as a channel tab or workspace sidebar item. Examples: Task Board, Whiteboard, Virtual Office (nice-to-have), Interactive Activities.
2. **Client-Only Personal Plugins** (per-user, opt-in, no server involvement): themes/CSS, sound packs, keybindings, the MediaPipe gesture-camera toggle.

In-house only for v1; manifest/gRPC contracts are designed cleanly so a public marketplace (Raycast/Obsidian-style) can be added later without rearchitecting.

**Permission manifest**: every plugin declares upfront, in its manifest, which scopes it needs (e.g. "read channel messages", "access user profile", "manage its own channel tab data") — validated against the `permissions` catalog described in §6.2. Even though all plugins are in-house for now (not third-party), this manifest + admin-approval-at-install pattern is built from day one rather than defaulting to full trust, since it's the same mechanism a future public marketplace will need — no rearchitecting required when that day comes.

**Virtual Office Plugin (optional)**: 2D spatial map with proximity-based audio attenuation, proximity-scoped track subscription (to avoid client overload at 50+ concurrent users), and soundproofed private zones. Explicitly a nice-to-have — the core product must remain usable without it, particularly on lower-spec devices.

### 6.9 Camera/Gesture AI
MediaPipe embedded client-side (not a third-party app/virtual camera) — processes video locally before publishing to LiveKit. Opt-in, default OFF on mobile/PWA to protect battery/thermals. A separate native macOS project ("GestureCam", Vision framework + CMIOExtension) may later serve as an advanced virtual-camera input option.

### 6.10 Search
Message search is implemented in **`ms-chat`** (not a separate microservice). Elasticsearch is populated asynchronously via a Kafka consumer in the same process (decoupled from the chat write path to avoid dual-write failure modes).

- **Scope**: global across all channels the user has access to (Teams/Discord-style), with filter operators (`in:channel`, `from:user`, `before:`/`after:date`, `has:file`) rather than search being scoped to only the currently-open channel
- **Attachment search**: v1 covers filename/metadata search; full-text search of file contents (extracted via a tool like `pdftotext`/Apache Tika in the upload pipeline, then indexed alongside message metadata) is a planned follow-up, not required for v1

### 6.11 Notifications
Push notifications are not an MVP priority but the architecture accounts for them from the start (@mention triggers routed through Kafka to a notification dispatcher). @mention itself is a day-one requirement.

- Android: Firebase Cloud Messaging (free, unlimited)
- iOS: Web Push API via VAPID (free, supported since iOS 16.4, no Apple Developer account needed) — trade-off is a standard push banner instead of a full-screen CallKit takeover

### 6.12 Moderation & Abuse Handling
Needed from day one, especially given the `open` registration mode allows unknown users to join. Baseline v1 scope (covered by the Admin role's permission set, no dedicated Moderator role yet):
- Report message/user
- Block user (personal, client-side filtering — doesn't affect what others see)
- Kick (remove from workspace) vs. Ban (kick + prevent rejoining): for registered users, enforced by blocking the same identity/email; for **Guests, who have no account/email**, enforced differently — banning a guest revokes their specific invite link/session token (invalidated immediately) and can optionally add an IP-based block for that channel, since there is no persistent identity to block otherwise
- Admin can delete any message, server-mute/deafen in voice

A dedicated "Moderator" role (kick/ban/delete but not manage plugins/settings) is a natural first real-world use case for the custom-role-ready RBAC schema (§6.2), but is not required for v1 — the Admin role covers moderation until that need is concrete.

**Automated abuse detection**: lives in **Core** (Chat module), not a plugin — this is treated as foundational safety infrastructure, the same category as rate limiting, not an optional feature. Behavior-pattern detection (e.g. flood: N messages within a short window) triggers a **temporary auto-mute/rate-limit + a flagged item in an admin review queue** — deliberately not an automatic permanent ban, since a false positive shouldn't be able to unilaterally ban someone, especially in a corporate context. Final punitive action (ban) stays a human admin decision.

### 6.13 Onboarding & Lobby
New members land in a default **Lobby/waiting-room channel** on join:
- `approval_required` mode: user sees a "waiting for admin approval" state, no workspace access yet
- Once approved (or under `open`/`invite_only` mode): user auto-joins a default channel (e.g. `#general` or `#lobby`), and can browse/join other public channels manually — no complex channel-discovery UI needed for v1

### 6.14 File Preview & Content Handling
- Core handles generic previews only: images, basic PDF, text/code
- Specialized file types (e.g. design files, CAD) require downloading a **Preview Plugin** first — same "core stays light, plugins extend" philosophy as the rest of the system, modeled after VS Code's file-type extensions
- Code blocks in chat get syntax highlighting via `highlight.js` (chosen over Shiki for lower weight, integrated with the TipTap editor)

## 7. Data Layer

| Store | Purpose |
|---|---|
| PostgreSQL (+ JSONB) | Primary store: users, workspaces, channels, messages, roles, audit logs, custom emoji metadata. Chosen for ACID + relational RBAC, and because the team is strongest in Go + PostgreSQL. |
| Redis | Ephemeral: pub/sub, cache, presence, guest session TTL, typing indicators, WebSocket fanout |
| Elasticsearch | Full-text message search index |
| MinIO (S3-compatible) | File storage — public bucket (avatars, emoji, soundboard) via cached public URLs; private bucket (DM/private-channel attachments) via short-lived presigned GET URLs. Uploads go directly client → MinIO via presigned PUT, bypassing backend bandwidth. |
| Kafka | Durable event backbone: chat events (search indexing), audit events, notification triggers, future analytics feed |

### 7.1 Data Retention
Two distinct, independently configurable settings — kept separate to avoid the two purge policies conflicting:
- **Data retention** (`retention_days`, default `0` = keep forever): governs how long *active* (non-deleted) chat history is kept before a daily cron job purges it. Admin-configurable per workspace (e.g. 30/90/365 days).
- **Trash grace period** (`trash_retention_days`, default `30`): governs how long *already soft-deleted* messages (e.g. deleted-but-replied-to messages, archived channel content) sit recoverable before the same daily cron job hard-purges them. Also admin-configurable, but independent of the data retention setting above — a workspace can keep full history forever while still cleaning up its trash after 30 days, or vice versa.

### 7.2 Backup & Disaster Recovery
- PostgreSQL: daily `pg_dump` (portable full snapshot) + continuous WAL archiving to a separate backup bucket/target
- **RPO note**: the daily `pg_dump` alone would imply a 24-hour worst-case data loss window, but continuous WAL archiving enables **Point-in-Time Recovery (PITR)** — restoring to within minutes of a failure, not just the last daily snapshot. This combination is already best-practice-tier and requires no further tightening at the current scale.
- MinIO: built-in versioning + replication to a separate instance
- Redis: not backed up (purely ephemeral)

### 7.3 Data Portability
Two separate mechanisms, serving different needs:
- **Self-service export** (Owner/Admin): "Export Workspace Data" generates an archive (JSON + file attachments) of the workspace's data — a concrete proof point for the "data sovereignty" positioning, not just a marketing claim
- **Raw database backup**: documented as a technical fallback (manual `pg_dump`) for sysadmins who want full control outside the UI

## 8. Security

- TLS/WSS in transit (mandatory baseline)
- Password hashing via bcrypt/argon2
- At-rest encryption via PostgreSQL/MinIO built-in server-side encryption
- Rate limiting at the API gateway from day one (chat, upload, channel creation)
- End-to-end encryption is explicitly out of scope — high implementation cost, low incremental value here since self-hosting already gives the deploying organization full server control
- Guest sessions are scoped and time-limited; guest uploads are quota-capped

## 9. Legal & Licensing

- **Source code license: AGPL-3.0** — permissive enough to stay free/open, while its network-use clause requires anyone who modifies WiChat and runs it as a service (even without distributing binaries) to open-source their changes back. This follows the same pattern as Mattermost, Rocket.Chat, and n8n, and protects against a third party forking WiChat into a closed-source paid SaaS without contributing back. An open-core model (proprietary premium plugins on top of an AGPL core) remains compatible with this.
- **Privacy Policy / Terms of Service**: not WiChat's responsibility as the software vendor — the self-hosting organization is the data controller. A generic `PRIVACY_TEMPLATE.md` (adaptable via a free generator like Termly/TermsFeed) is included in the repo for deploying organizations to customize with their own name and contact details.

## 10. Observability & Testing

- **Observability** (target): Prometheus metrics + Grafana; application logs via slog → Loki (see [logging.md](logging.md)). Metrics/tracing/health endpoints are **not implemented** in the Go services yet.
- **Testing**:
  - Unit tests: Go `testing` + `testify`, ~90% coverage target
  - Integration tests: testcontainers (real Postgres/Redis in CI)
  - E2E tests: Playwright, scoped to critical paths only (login, send message, join channel, start call) — not held to the same 90% bar

## 11. Deployment & Operations

- Target: local/self-hosted, single `docker-compose.yml` orchestrating Postgres, Redis, MinIO, Kafka, Elasticsearch, LiveKit, and the WiChat core binary (`docker compose up -d`)
- No setup wizard beyond env/config-file setup + the first-run admin account form
- Updates:
  - Server: new Docker image
  - Desktop: `electron-updater` auto-update (same pattern as Discord/Slack/VS Code)
  - Android: user-driven APK re-download, or in-app update check
  - iOS/Web: always latest on load, no special mechanism
- Kubernetes/clustered scaling is deferred to when an infra lead engages, not required for the 2–500-person target
- **Environments**: dev, staging, and production, differentiated via `docker-compose.override.yml` layers + per-environment `.env` files
- **Documentation**: a self-hosting install guide (README-level) is planned but deferred to closer to release, not an MVP blocker
- **No centralized demo instance**: no publicly-hosted WiChat instance is planned for demoing to prospects — fully self-host, every deployment is the adopter's own instance

## 12. Client Strategy

| Platform | Approach | Notes |
|---|---|---|
| Web | Next.js | Primary client, also doubles as the iOS experience |
| iOS | PWA (no native app) | Avoids the $99/year Apple Developer Program requirement and App Store review; Web Push supported since iOS 16.4. Trade-off: no CallKit full-screen incoming call UI. |
| Android | React Native | Sideloadable APK from GitHub/landing page — no store fees required; FCM push; native module for `ConnectionService` background call handling |
| Desktop | Electron | Wraps the web client; auto-updates |

### 12.1 Accessibility
Handled primarily through tooling choice rather than a dedicated a11y workstream: UI built on **shadcn/ui** (on top of Radix UI primitives), which provides keyboard navigation and ARIA labeling by default, plus baseline discipline (semantic HTML, image alt text, visible focus states).

### 12.2 Offline Strategy
No offline-first support — mobile/web require an active connection. This is a deliberate simplification: it avoids local database sync, conflict resolution, and cache-invalidation complexity. On reconnect, the client relies on the existing catch-up sync mechanism (`last_synced_message_id`, §6.4) rather than a local offline cache.

## 13. API & Data Conventions

- REST for CRUD + existing WebSocket channel for realtime — GraphQL is deferred unless a concrete need for complex nested queries emerges
- Versioning: URI-based (`/api/v1/`); old versions stay live until all clients (including slower-to-update mobile) migrate
- Migrations: [goose](https://github.com/pressly/goose) per service under `backend/<service>/internal/migrations/` (not applied on process startup)
- i18n: backend emits message codes/keys only, never raw text; frontend resolves via translation files (i18next for web/Electron, go-i18n for backend-generated text like emails); ICU-style interpolation for dynamic values
- Dark mode is a day-one requirement; a shared design-token system underlies web/desktop/mobile for visual consistency
- **Error handling**: standard HTTP status codes + JSON body `{ "error": { "code": "<i18n_key>", "params": { ... } } }` — backend emits translation keys only; see [error-handling.md](error-handling.md)
- **Sessions**: JWT access token (15 min expiry) + refresh token (30 days, httpOnly cookie, rotated on every use — the old refresh token is invalidated as soon as a new one is issued)
- **Message & upload limits** (workspace-configurable, defaults below):
  - Message length: 4,000 characters
  - Image upload: 10 MB
  - Document upload: 25 MB
  - Video/large file upload: 100 MB
- **Timezone**: all timestamps stored as UTC (`TIMESTAMPTZ` in Postgres); converted to the viewer's local timezone client-side at render time

## 14. Scale Assumptions

Target: **2–500 people per workspace**. At this scale, a single instance each of PostgreSQL, Redis (optionally with Sentinel), Elasticsearch, and a single well-provisioned LiveKit node is sufficient — no sharding or clustering is required at launch. Core **microservices** scale horizontally only when metrics justify it (typically **chat** and **api-gateway** first); other services stay single-instance at this tier. Polyglot extraction (Python ML, etc.) remains possible via new services without redesigning domain boundaries.

**Workspaces per instance**: no fixed cap is set for v1. The real constraint is total concurrent load — particularly voice/video (LiveKit) and active WebSocket connections — not raw workspace count; Postgres/Redis/Elasticsearch handle many workspaces comfortably before that becomes a bottleneck. Grafana observability (already planned) is the intended signal for when scaling action is needed, rather than a number decided upfront.

**Workspace branding**: default light/dark design tokens only for v1 (no custom logo/color per workspace yet) — deferred until the product's visual identity is settled; adding it later is a small addition (`logo_url`, `primary_color` fields on the existing `workspaces` table), not a redesign.

## 15. Build Roadmap (Phase Summary)

| Phase | Scope |
|---|---|
| 0 | Foundation: repo, docker-compose, Go **service** skeletons + gRPC contracts, CI |
| 1 | Auth & Workspace: registration modes, login, roles, multi-workspace |
| 2 | Core Chat: channels, messaging, reactions, read receipts, sync |
| 3 | Voice/Video: LiveKit integration, call controls |
| 4 | Plugin Manager: gRPC contract + dummy plugin validation |
| 5 | Guest Access: link-based join, restricted sessions |
| 6 | Polish: search, notifications, i18n, dark mode |
| 7 | Real Plugins: task board, whiteboard, virtual office |
| 8 | Hardening: Kafka backbone, audit log, observability, backup |

Each phase will be detailed to implementation-level (exact schemas, endpoints, configs) when work on it begins, rather than all at once up front.

## 16. Open Questions

- None currently blocking — data retention defaults and encryption scope have been resolved as documented above.

---
*This document reflects all architecture and product decisions made through the current planning session and supersedes earlier informal notes.*
