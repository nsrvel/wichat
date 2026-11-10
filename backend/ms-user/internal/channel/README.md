# channel (ms-user)

Workspace **channel metadata** (list, default seed, sidebar section) — not messages (`ms-chat`).

## Layout

| Path | Role |
|------|------|
| `repository/pg_repository.go` | Postgres list + `SeedDefaultChannels`; tx helpers `InsertDefaultChannelsTx` for co-located transactions |
| `service/service.go` | `ListGroupChannels`, `SeedDefaultChannels`; tests [`service/test/service_test.go`](service/test/service_test.go) |
| `delivery/grpc/server.go` | `ListGroupChannels` RPC (mounted on `AuthService` until proto splits) |
| `repository/mock/` | Fast tests |

## Boundaries

- **group** creates rows in `groups` / `group_members` only; **channel** owns `channels` rows.
- **Create group:** `group/service.CreateGroup` calls `channel/service.SeedDefaultChannels` after insert (rollback delete group on seed failure).
- **Instance setup:** `user/repository.CompleteInstanceSetup` runs `InsertGroupWithOwnerTx` + `InsertDefaultChannelsTx` in one transaction (bootstrap only).

Default seed: **General** (`text_and_voice`, `is_private=false`), **Announcements** (`broadcast`, favorites). JSON field `type` is the channel **kind**; visibility uses `is_private` (always `false` for `dm`). Client default icons use kind when `image_url` is empty.

Models: [`internal/models/channel_model.go`](../models/channel_model.go), [`channel_repository_model.go`](../models/channel_repository_model.go).

See `wichat-go-layers.mdc`.
