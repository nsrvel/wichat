# Shell UX — workspace switching & rail

Status: **draft (product R&D)** — complements [wichat-system-design.md](../architecture/wichat-system-design.md) § workspace model (single domain, multi-workspace, picker).

## Problem

Users span two patterns:

- **Many communities** (game servers, OSS) — prefer quick context switch (Discord-style rail).
- **Few work workspaces** (freelance, small org) — prefer focus; a full rail of every membership feels noisy.

Spec requires a **workspace picker** on one domain; it does **not** forbid an optional rail — interpret picker as the **canonical list** (search, join, all memberships), not “dropdown only”.

## Principles

1. **One shell layout** — sidebar (channels/DMs) + main + optional member panel; rail is an optional **left strip**, not four different products.
2. **Picker is source of truth** — every workspace reachable without rail; rail = shortcuts.
3. **User preference** (synced per account, not only localStorage) — avoid shipping four separate playground layouts in production.

## Workspace switcher modes (v1 target)

| Mode | UI | Default for |
|------|-----|-------------|
| `sidebar_only` | Dropdown / list in sidebar header | 1 workspace; corporate-first |
| `pinned_rail` | Rail shows **pinned** workspaces only; overflow via `+N` + picker | **Recommended default** when user has 2+ workspaces |
| `full_rail` | Every membership on rail (Discord-like) | Opt-in; “many communities” |

Deferred: **Teams-style app nav rail** (Home / Chat / Activity) — product navigation, not workspace switching; add when those modules exist.

### Smart default (automatic)

- 1 workspace → `sidebar_only`.
- 2+ workspaces → start `sidebar_only`; optional one-time hint to enable `pinned_rail` after joining a 3rd workspace.
- User overrides in **Preferences → Appearance → Workspace switcher**.

## Rail content (when enabled)

- Top: Home (DM hub — when implemented), divider.
- Middle: workspace avatars (per mode above), unread badges.
- Bottom: Browse workspaces, Add/join (opens picker flow). **User profile** spans **rail + sidebar** in one footer panel.
- **Not** on rail v1: public server discovery (out of scope for self-hosted instance).

## Implementation order (frontend)

1. **App shell** — `app/(app)/layout.tsx` + `components/layout/app-shell.tsx` (rail, channel sidebar, user bar, main slot). Playground removed.
2. **Full rail data** — workspace list from API; switch workspace context.
3. **Picker + sidebar-only path** — `railMode=sidebar_only` hides strip.
4. **Pinned rail + overflow** — filter rail by pin list; pins stored user-side when API exists.
5. **Preferences enum** — wire the three modes; default automatic rules.

Rationale: **full rail is the easiest rail to build**; **pinned + picker is the best default to ship** once pin API and settings exist. Do not maintain four divergent layouts from R&D playgrounds.

## References (local mockups)

`references/dribble/` — Conceptzilla/Snazzy (3-col + panel), Discord web (rail + categories), Teams Mac (module rail). Not committed; optional inspiration only.

### Rail visual (v1 R&D)

- **Chrome:** Conceptzilla layout — rail `shell-rail` (`#F0F0F0`), tiles `shell-rail-tile` (white), channel column `sidebar` (white), user panel `sidebar-accent`.
- **Behavior:** full membership list on rail; user foot spans rail + sidebar width.
- **Theme:** shell chrome uses **design tokens only** (`globals.css` / Tailwind semantic colors). Presence dots use `--presence-*`. Toggle light/dark via theme control to review both.

### Sidebar user bar (planned)

- **Placement:** Bottom of **workspace rail + channel sidebar** — one `#232428` panel spanning the full nav column (~292px), not sidebar-only.
- **Contents (shell R&D / early chat):** avatar + presence ring, display name, optional status line, settings/user menu — **no mic / deafen / disconnect**.
- **Voice controls:** separate `VoiceCallControls` (or meet strip) when LiveKit session is active — **Phase 3 voice/video** per system design; do not add mute/deafen to the profile bar before then.

## Out of scope (this doc)

- Cross-instance workspace list (separate login per instance).
- Unified inbox across workspaces.
- Server folders on rail (consider post-v1).

**App URL scheme** (e.g. `/channels/@me`, `/channels/{workspaceId}/{channelId}`) is deferred until the workspace/channel API slice lands with BE; the web-client demo is **`/` only** until then.

