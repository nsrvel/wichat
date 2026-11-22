# Presence (live)

## Data (ms-user → Redis)

| Key | TTL |
|-----|-----|
| `presence:conn:{user}:{connId}` | 60s |
| `presence:conn_ids:{user}` | ~120s |
| `presence:manual:{user}` | optional |
| `presence:effective:{user}` | 60s |

`resolveUserEffectivePresence` on **GetMyPresence** and **BatchGetEffectivePresence** when effective key is missing but connections exist.

**Degrade (no design change to TTL/heartbeat):** Redis or ms-user read failures must not 500 member lists or `GET /users/me/presence`. ms-user returns offline snapshots / degraded self view; api-gateway maps gRPC `Unavailable` on GetMyPresence to HTTP 200 with `DegradedMyPresenceBody` (manual `online`, effective `offline`).

## Realtime (api-gateway)

`GET /api/v1/realtime` — WebSocket upgrade with `Sec-WebSocket-Protocol: bearer` (short marker) or legacy `bearer.<jwt>`. First `presence.heartbeat` carries `access_token` when the upgrade used deferred auth. Query `access_token` is rejected.

| Client → server | Server → client |
|-----------------|-----------------|
| `presence.heartbeat` | `presence.updated` |
| `subscribe_group` (24-hex public group id only) | `subscribe.denied` (never closes socket) |

## Web client (`frontend/apps/web/src/features/presence/`)

| File | Role |
|------|------|
| `gateway-socket.ts` / `live-connection.ts` | One WS/tab via unified gateway socket |
| `live-cache.ts` | Patches `Q_MY_PRESENCE` + `Q_INVITE_MEMBERS` |
| `presence-gate.tsx` | Mounted in `AppProviders` when session exists |
| `use-live-presence.ts` | Avatar dot + manual status for picker |
| `presence-dot.ts` | `presenceDotFrom` (self), `memberPresenceDot` (lists) |
| `use-group-online-count.ts` | Live online tally from member cache |
| `live-cache.ts` | Patches `Q_INVITE_MEMBERS`, `Q_GROUP_ROLE_MEMBERS`, debounced `Q_INVITE_PREVIEW` on `/invite/*` |

No localStorage for dots. `presence-gate` **fetchQuery** bootstraps `Q_MY_PRESENCE` before WS can seed the cache. Member UIs use `memberPresenceDot`; self chrome uses `useLivePresence`.