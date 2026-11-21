# Presence (live)

## Data (ms-user → Redis)

| Key | TTL |
|-----|-----|
| `presence:conn:{user}:{connId}` | 60s |
| `presence:conn_ids:{user}` | ~120s |
| `presence:manual:{user}` | optional |
| `presence:effective:{user}` | 60s |

`resolveUserEffectivePresence` on **GetMyPresence** and **BatchGetEffectivePresence** when effective key is missing but connections exist.

## Realtime (api-gateway)

`GET /api/v1/realtime?access_token=…` (WebSocket upgrade)

| Client → server | Server → client |
|-----------------|-----------------|
| `presence.heartbeat` | `presence.updated` |
| `subscribe_group` (24-hex public group id only) | `subscribe.denied` (never closes socket) |

## Web client (`frontend/apps/web/src/features/presence/`)

| File | Role |
|------|------|
| `live-connection.ts` | One WS/tab; query-string token auth |
| `live-cache.ts` | Patches `Q_MY_PRESENCE` + `Q_INVITE_MEMBERS` |
| `presence-gate.tsx` | Mounted in `AppProviders` when session exists |
| `use-live-presence.ts` | Avatar dot + manual status for picker |
| `presence-dot.ts` | `presenceDotFrom` (self), `memberPresenceDot` (lists) |
| `use-group-online-count.ts` | Live online tally from member cache |
| `live-cache.ts` | Patches `Q_INVITE_MEMBERS`, `Q_GROUP_ROLE_MEMBERS`, debounced `Q_INVITE_PREVIEW` on `/invite/*` |

No localStorage for dots. `presence-gate` **fetchQuery** bootstraps `Q_MY_PRESENCE` before WS can seed the cache. Member UIs use `memberPresenceDot`; self chrome uses `useLivePresence`.