# WiChat Phase 3 — Voice & Video Implementation Plan

> **Status:** Production-Ready Implementation Plan (Planning Only)
> **Source of Truth:** `~/workspace/your_files/wichat-voice-video-spec.en.md` (with 2026-11-17 Locked Decisions in §13)
> **Grounded against Repo:** `~/workspace/wichat-repo` (`nsrvel/wichat`)

---

## 1. Overview

### 1.1 Scope Recap
WiChat Phase 3 delivers drop-in voice, video, and screen sharing across unified channels (`voice_only`, `text_and_voice`) and `dm` conversations, backed by **LiveKit Cloud (Build/free plan)** as the initial SFU target with a zero-friction migration path to self-hosted LiveKit (`livekit-server` + `livekit-egress`).

**Locked Product Decisions (Spec §13 — Non-Negotiable):**
1. **Deployment Target & Ownership:** LiveKit Cloud (Build plan: 5,000 participant-minutes/month, 100 concurrent connections, 50 GB downstream/month) is the initial target; self-host is the scale path. **`ms-core` (`:3003` gRPC)** owns LiveKit JWT issuance, webhook processing, moderation enforcement, and call lifecycle.
2. **Drop-In Model & Room Caps:** Both group channels (`voice_only`, `text_and_voice`) and `dm` conversations use drop-in join/leave (no ringing in v1). Per-room caps are enforced at **50 total participants** and **25 active video/screenshare publishers** (plus per-channel `user_limit` when set below 50).
3. **Call Recording (First Release):** Required in v1 via **LiveKit Egress → MinIO** (S3-compatible), gated by a new `record_call` permission (default: Owner/Admin), with a mandatory visible recording indicator (red dot + `"Recording"`) for all participants and a pre-join notice if recording is already active.
4. **Idle-Kick:** 30 minutes of audio/video/UI inactivity triggers automatic removal, with a 2-minute prior warning prompt where any interaction resets the timer.

### 1.2 Architecture Recap & Grounding in Existing Repo
 inspection of `wichat-repo` reveals several important realities that shape the implementation:

```mermaid
flowchart LR
    subgraph Clients["Frontend (apps/web + Electron)"]
        UI["Voice Stage / Tile Grid / VoiceCallControls"]
        LKClient["livekit-client SDK"]
        WSClient["Realtime WS (/api/v1/realtime)"]
    end

    subgraph Gateway["backend/api-gateway (:3000 HTTP/WS)"]
        HTTPVoice["internal/voice/delivery/http"]
        Fanout["internal/kafka/consumers/realtime_fanout.go"]
        RTHub["internal/realtime/hub"]
    end

    subgraph Core["backend/ms-core (:3003 gRPC)"]
        VoiceSvc["internal/voice/service"]
        VoiceRepo["pg_repository + redis_occupancy + minio"]
        LKSDK["livekit/server-sdk-go/v2"]
    end

    subgraph User["backend/ms-user (:3001 gRPC)"]
        ChanSvc["ChannelService (AssertGroupChannelAccess)"]
        RoleSvc["GroupRoleService (RBAC + record_call)"]
        PresSvc["Presence / Status Note (Activity Status)"]
    end

    subgraph Infra["Shared Infrastructure"]
        PG[("PostgreSQL (public schema)")]
        Redis[("Redis (:6379)")]
        Kafka[["Kafka (chat-events)"]]
        MinIO[("MinIO (:9000 S3)")]
        LKCloud(("LiveKit Cloud SFU + Egress"))
    end

    UI -->|POST /voice/token, /moderate, /recording| HTTPVoice
    HTTPVoice -->|gRPC core.v1.VoiceService| VoiceSvc
    VoiceSvc -->|gRPC AssertGroupChannelAccess| ChanSvc
    VoiceSvc -->|RoomService / EgressClient| LKCloud
    LKClient <-->|WSS WebRTC Media| LKCloud
    LKCloud -->|POST /api/v1/webhooks/livekit| HTTPVoice
    VoiceSvc <-->|Live Occupancy| Redis
    VoiceSvc <-->|call_sessions, call_participants, call_recordings| PG
    LKCloud -->|S3 Egress Output| MinIO
    VoiceSvc -->|Produce voice.* events| Kafka
    Kafka --> Fanout --> RTHub --> WSClient
    VoiceSvc -->|Auto Activity Status ("In a call")| PresSvc
```

**Key Repo Grounding Notes (Spec vs. Repo Alignment):**
- **Database & Migrations:** Although `wichat-system-design.md` §13 mentions per-service migrations, in the actual repository **all 48 goose SQL migrations live in `backend/ms-user/internal/migrations/`** against the shared `wichat` PostgreSQL database (`backend/Makefile` runs `$(MAKE) -C ms-user migrate-up`). `ms-chat` has no `internal/migrations/` folder; its `00021_messages.sql` and `00047_chat_production_features.sql` are in `ms-user/internal/migrations/`. Our plan places new migrations (`00049`–`00050`) in `backend/ms-user/internal/migrations/` (or wires `ms-core` to the same shared Postgres pool) to stay 100% consistent with `make migrate-up`.
- **`check-layout` Enforcement:** `backend/tools/check-layout/main.go` strictly enforces:
  1. Repository files in `internal/<domain>/repository/` must match `^(pg_repository|minio_repository|es_repository|redis_[a-z0-9_]+)\.go$`, with `mock/pg_repository_mock.go` required whenever `pg_repository.go` exists (and **no** mock for Redis — `miniredis` is used directly).
  2. Every production `.go` file in a tested package must have a **1:1** `test/<file>_test.go` (no extra test files, no tests beside production `.go`).
  3. `ms-*` modules must maintain **≥90% scoped unit test coverage** (`scripts/coverage-gate.sh`).
- **Presence & Activity Status Service Boundary:** Spec §3.1 and §5.6 attribute presence/activity status to `ms-chat`. In the actual codebase, **`PresenceService` (`user.v1.PresenceService`, `internal/presence/`) and `status_note_*` (`users` table) live in `ms-user`**, not `ms-chat`. Therefore, `ms-core` will coordinate automatic call activity status (`"In a call"` / `"In a meeting"`) via **`ms-user`** (and broadcast via the existing group realtime hub), while preserving manual user overrides.
- **Channel Types (`text_and_voice`):** Migration `00042_channel_categories_and_types.sql` dropped `text_and_voice` from `channels_type_check`, leaving `('broadcast', 'text_only', 'voice_only', 'forum', 'dm')`. To satisfy Spec §1 and §11 (*"Two users in a `text_and_voice` channel..."*), M0 restores `text_and_voice` in the DB constraint, `ms-user` channel validation, and `@wichat/contracts`.
- **Frontend `ShellUserDock` Cleanup:** `frontend/apps/web/src/components/layout/shell-user-dock.tsx` (lines 183–202) currently renders local dummy `Mic` and `Headphones` buttons directly in the user profile dock. Per `docs/design/shell-ux.md` §Sidebar user bar (*"no mic / deafen / disconnect... Voice controls: separate `VoiceCallControls`"*), we will remove those dummy buttons from `ShellUserDock` and mount the dedicated `VoiceCallControls` meet strip above the dock when a LiveKit session is active.

### 1.3 Critical Path
`M0-1 (Egress/Quota Spike)` + `M0-2 (DB Migrations & Channel Kind)` → `M0-3 (Proto & wi-shared)` → `M0-4 (ms-core Skeleton)` → `M1-1 (Token Issuance & Caps)` → `M1-2 (Webhooks, Redis Occupancy & Kafka Fanout)` → `M1-3 (Gateway Endpoints & WS Fanout)` → `M1-5 (Frontend LiveKit Provider, Sidebar Occupancy & Meet Strip)` → `M2 (Video, Screen Share & Stage Grid/Focus)` → `M3 (Moderation, Presence & Idle-Kick)` → `M4 (LiveKit Egress Recording to MinIO)` → `M5 (Quota Telemetry & E2E Verification)`.

---

## 2. Work Breakdown

All tasks are sized to respect the protocol's diff budgets (**Small (S)**: ≤2 files, ≤60 lines; **Medium (M)**: ≤8 files, ≤500 lines) so Hydrogen can execute each task cleanly within a single turn/session without hitting hard-stop triggers.

### Milestone 0 — Foundations (Verification Spike, Schema, Contracts, `ms-core` Skeleton)

#### `M0-1`: LiveKit Cloud Build Plan Egress & Quota Verification Spike
- **Owning area:** `infra` / `docs`
- **What gets built:** Early verification task (per Spec §5.8 & §8). Verify against a live LiveKit Cloud Build project: (1) whether `StartRoomCompositeEgress` / `StartTrackCompositeEgress` to external S3 (MinIO) is enabled on the free Build tier and how egress minutes/participant-minutes are metered (whether egress counts as +1 participant in the room or requires Ship plan / self-hosted `livekit-egress` container in `deploy/compose/livekit.yml`), and (2) document exact `LIVEKIT_URL`, `LIVEKIT_API_KEY`, `LIVEKIT_API_SECRET`, and S3 endpoint reachability requirements in `deploy/README.md` and `deploy/.env.example`. If Cloud Build restricts egress on free tier, add `deploy/compose/livekit.yml` (local `livekit/livekit-server` + `livekit/egress` profile) so recording works out-of-the-box in dev/self-host and on Cloud Ship.
- **Files/areas touched:**
  - `deploy/README.md`
  - `deploy/.env.example`
  - `deploy/compose/livekit.yml` (optional compose profile for local/self-host LiveKit + Egress)
- **Dependencies:** None (runs first).
- **Size:** **S** (≤3 files, ~55 lines)
- **Verification step:** Run `livekit-cli` or curl test against LiveKit Cloud `Twirp` `Egress/StartRoomCompositeEgress` endpoint; verify `docker compose -f deploy/docker-compose.yml config` succeeds.

#### `M0-2`: Database Migrations for Voice Sessions, Recordings, `record_call`, and `text_and_voice`
- **Owning area:** `ms-user` (migrations & constants)
- **What gets built:**
  1. Goose migration `00049_voice_schema_and_permissions.sql`:
     - Restore `'text_and_voice'` to `channels_type_check` constraint (`broadcast`, `text_only`, `voice_only`, `text_and_voice`, `forum`, `dm`).
     - Create `call_sessions` table (`id`, `public_id`, `group_id`, `channel_id`, `room_name`, `started_at`, `ended_at`, `peak_participants`, `total_participant_seconds`).
     - Create `call_participants` table (`id`, `session_id`, `user_id`, `joined_at`, `left_at`, `duration_seconds`, `server_muted`, `server_deafened`, `disconnect_reason`).
     - Create `call_recordings` table (`id`, `public_id`, `session_id`, `channel_id`, `egress_id`, `status`, `file_key`, `file_size_bytes`, `duration_seconds`, `started_by`, `stopped_by`, `started_at`, `ended_at`).
  2. Add `PermRecordCall = "record_call"` to `backend/ms-user/pkg/constants/permissions.go` (`AllKnownPermissions`) and `ChannelKindTextAndVoice = "text_and_voice"` to `backend/ms-user/pkg/constants/channels.go` and `backend/ms-user/internal/channel/service/service.go`.
  3. Update `docs/architecture/wichat-permissions-curated.md` §4 with `record_call`.
- **Files/areas touched:**
  - `backend/ms-user/internal/migrations/00049_voice_schema_and_permissions.sql`
  - `backend/ms-user/pkg/constants/permissions.go`
  - `backend/ms-user/pkg/constants/channels.go`
  - `backend/ms-user/internal/channel/service/service.go`
  - `backend/ms-user/internal/channel/service/test/service_test.go`
  - `docs/architecture/wichat-permissions-curated.md`
- **Dependencies:** None.
- **Size:** **M** (6 files, ~240 lines)
- **Verification step:** `make -C backend migrate-up && make -C backend/ms-user check`

#### `M0-3`: Protobuf Contracts (`core/v1`) & Shared Error/Event Definitions in `wi-shared`
- **Owning area:** `wi-shared`
- **What gets built:**
  1. Define `backend/wi-shared/models/proto/core/v1/general.proto` and `backend/wi-shared/models/proto/core/v1/voice.proto` (`VoiceService`: `IssueVoiceToken`, `ModerateParticipant`, `HandleWebhook`, `GetChannelOccupancy`, `ListGroupOccupancy`, `StartRecording`, `StopRecording`, `ListChannelRecordings`, `GetQuotaUsage`, `AcknowledgeIdleWarning`).
  2. Run `make -C backend/wi-shared proto` to generate `backend/wi-shared/models/gen/core/v1/*`.
  3. Add voice `MessageCode` constants in `backend/wi-shared/exception/code.go` (`voice.room_full`, `voice.video_limit_reached`, `voice.channel_not_voice_capable`, `voice.quota_exceeded`, `voice.recording_already_active`, `voice.recording_not_active`, `voice.webhook_invalid_signature`, `voice.guest_not_allowed`) and allow `"voice."` prefix in `ValidMessageCode`.
  4. Add voice Kafka event structs in `backend/wi-shared/chat-events/events.go` and WebSocket envelope marshalers in `backend/wi-shared/realtime/chat_event.go` & `envelope_from_chat_events.go` (`voice.occupancy.updated`, `voice.recording.updated`, `voice.moderation.applied`, `voice.idle.warning`).
- **Files/areas touched:**
  - `backend/wi-shared/models/proto/core/v1/general.proto`
  - `backend/wi-shared/models/proto/core/v1/voice.proto`
  - `backend/wi-shared/exception/code.go`
  - `backend/wi-shared/exception/test/code_test.go`
  - `backend/wi-shared/chat-events/events.go`
  - `backend/wi-shared/chat-events/test/events_test.go`
  - `backend/wi-shared/realtime/chat_event.go`
  - `backend/wi-shared/realtime/test/chat_event_test.go`
- **Dependencies:** `M0-2`
- **Size:** **M** (8 source files + generated proto output, ~380 lines)
- **Verification step:** `make -C backend/wi-shared proto buf-lint check`

#### `M0-4`: Outbound `ms-core` gRPC Client Facade in `wi-shared/microservices`
- **Owning area:** `wi-shared`
- **What gets built:**
  - Create `backend/wi-shared/microservices/services/ms-core/ms_core.go`, `grpc/general/client.go`, and `grpc/voice/client.go` mirroring `services/ms-chat/`.
  - Wire `CoreAddr` and `Core msCore.Service` into `backend/wi-shared/microservices/clients.go` and health monitor targets.
- **Files/areas touched:**
  - `backend/wi-shared/microservices/services/ms-core/ms_core.go`
  - `backend/wi-shared/microservices/services/ms-core/grpc/general/client.go`
  - `backend/wi-shared/microservices/services/ms-core/grpc/voice/client.go`
  - `backend/wi-shared/microservices/clients.go`
  - `backend/wi-shared/microservices/test/clients_test.go`
- **Dependencies:** `M0-3`
- **Size:** **M** (5 files, ~260 lines)
- **Verification step:** `make -C backend/wi-shared check`

#### `M0-5`: `ms-core` Service Skeleton (`:3003` gRPC), Config, Health & Layout Wiring
- **Owning area:** `ms-core` / `infra`
- **What gets built:**
  - Create module `backend/ms-core/` (`go.mod`, `Makefile`, `coverage.scope`, `.env.example`, `cmd/main.go`, `internal/config/config.go`, `internal/general/delivery/grpc/server.go`, `internal/server/{server.go,server_grpc.go,server_run.go}`).
  - Register `./ms-core` in `backend/go.work`, `backend/Makefile` (`check`, `run`, `dev-ms-core`, `run-ms-core`), and `backend/tools/check-layout/main.go` module list.
- **Files/areas touched:**
  - `backend/ms-core/go.mod`, `backend/ms-core/Makefile`, `backend/ms-core/.env.example`, `backend/ms-core/coverage.scope`
  - `backend/ms-core/cmd/main.go`
  - `backend/ms-core/internal/config/config.go` + `test/config_test.go`
  - `backend/ms-core/internal/general/delivery/grpc/server.go`
  - `backend/ms-core/internal/server/{server.go,server_grpc.go,server_run.go}`
  - `backend/go.work`, `backend/Makefile`, `backend/tools/check-layout/main.go`
- **Dependencies:** `M0-4`
- **Size:** **M** (8 core files + build manifests, ~390 lines)
- **Verification step:** `make -C backend check-layout && make -C backend/ms-core check`

---

### Milestone 1 — Voice-Only Core (Tokens, Webhooks, Occupancy, Sidebar & Meet Strip)

#### `M1-1`: `ms-core` Postgres & Redis Repositories (`call_sessions`, `call_participants`, Live Occupancy)
- **Owning area:** `ms-core`
- **What gets built:**
  - Domain models in `backend/ms-core/internal/models/voice_model.go` and `voice_repository_model.go`.
  - Postgres repository `backend/ms-core/internal/voice/repository/pg_repository.go` + `mock/pg_repository_mock.go` (+ `test/pg_repository_mock_test.go`) for `call_sessions`, `call_participants`, and monthly participant-seconds aggregation.
  - Redis repository `backend/ms-core/internal/voice/repository/redis_occupancy.go` (+ `test/redis_occupancy_test.go` using `miniredis`) storing per-channel occupant hashes, mute/deafen/video/screenshare flags, last-active timestamps for idle-kick, and per-user token issuance rate-limit counters.
- **Files/areas touched:**
  - `backend/ms-core/internal/models/voice_model.go`
  - `backend/ms-core/internal/models/voice_repository_model.go`
  - `backend/ms-core/internal/voice/repository/pg_repository.go`
  - `backend/ms-core/internal/voice/repository/mock/pg_repository_mock.go`
  - `backend/ms-core/internal/voice/repository/mock/test/pg_repository_mock_test.go`
  - `backend/ms-core/internal/voice/repository/redis_occupancy.go`
  - `backend/ms-core/internal/voice/repository/test/redis_occupancy_test.go`
- **Dependencies:** `M0-5`
- **Size:** **M** (7 files, ~460 lines)
- **Verification step:** `make -C backend check-layout && make -C backend/ms-core check`

#### `M1-2`: `ms-core` LiveKit Token Issuance, Permission Mapping & Room/Video Cap Enforcement
- **Owning area:** `ms-core` / `ms-user`
- **What gets built:**
  - LiveKit SDK wrapper in `backend/ms-core/pkg/livekit-client/client.go` (+ `test/client_test.go`) encapsulating JWT generation (`auth.AccessToken` with 15m TTL, `identity = user_uuid`, `room = channel:{channel_uuid}` or `dm:{channel_uuid}`, metadata JSON `{display_name, username, avatar_url}`), `RoomServiceClient` (`UpdateParticipant`, `RemoveParticipant`, `MutePublishedTrack`), and `WebhookReceiver` signature validation.
  - Extend `AssertGroupChannelAccessResponse` in `ms-user` (`channel.proto` + `ms-user/internal/channel`) to also return `bitrate_kbps`, `video_quality`, `user_limit`, `rtc_region`, and handle `dm` channel membership verification.
  - Implement `IssueVoiceToken` in `backend/ms-core/internal/voice/service/service.go`:
    - Reject guests (`voice.guest_not_allowed`), non-voice channel types (`broadcast`, `text_only`, `forum` → `voice.channel_not_voice_capable`), and missing `connect` permission (`auth.forbidden`).
    - Rate-limit token issuance per user in Redis (max 10 token requests / minute → `error.rate_limited`) to prevent room-hopping abuse (Spec §10).
    - Enforce room cap: reject 51st participant (or `channel.user_limit + 1` when `0 < user_limit < 50`) with `voice.room_full`.
    - Map permissions to LiveKit grants: `canSubscribe = true`; `canPublish = speak || video`; `canPublishSources = [microphone (if speak), camera & screen_share & screen_share_audio (if video)]`. If `!speak`, mark participant initial state as server-muted.
- **Files/areas touched:**
  - `backend/ms-core/pkg/livekit-client/client.go`
  - `backend/ms-core/pkg/livekit-client/test/client_test.go`
  - `backend/ms-core/internal/voice/service/service.go`
  - `backend/ms-core/internal/voice/service/test/service_test.go`
  - `backend/ms-core/internal/voice/delivery/grpc/server.go`
  - `backend/wi-shared/models/proto/user/v1/channel.proto`
  - `backend/ms-user/internal/channel/delivery/grpc/server.go`
- **Dependencies:** `M1-1`
- **Size:** **M** (7 files, ~450 lines)
- **Verification step:** `make -C backend check` (unit tests verify grant mapping for `connect`, `speak`, `video`, 51st join rejection, guest rejection, and rate limit).

#### `M1-3`: `ms-core` LiveKit Webhook Processor & Kafka Occupancy Fanout
- **Owning area:** `ms-core`
- **What gets built:**
  - Implement `HandleWebhook` in `ms-core`: verify LiveKit webhook signature (`Authorization` SHA256 JWT header signed by `LIVEKIT_API_KEY`/`SECRET`); reject invalid signatures with `Unauthenticated` (`voice.webhook_invalid_signature`).
  - Handle webhook lifecycle events idempotently (deduplicated by `event.id` in Redis):
    - `room_started`: create `call_sessions` row in Postgres + initialize Redis room state.
    - `participant_joined`: add participant to Redis occupancy hash, insert `call_participants` row, update `call_sessions.peak_participants`, emit `voice.occupancy.updated` to Kafka `chat-events`.
    - `participant_left`: remove from Redis occupancy hash, set `call_participants.left_at` & `duration_seconds`, accumulate `call_sessions.total_participant_seconds`, emit `voice.occupancy.updated`.
    - `room_finished`: close `call_sessions.ended_at`, close any open `call_participants` rows, clear Redis room keys, emit empty `voice.occupancy.updated`.
    - `track_published` / `track_unpublished`: update participant `is_camera_on`, `is_screen_sharing`, `is_muted` in Redis; enforce **max 25 concurrent video publishers** per room (if 26th video track is published, mute/unpublish track via `RoomServiceClient.MutePublishedTrack` and notify client); emit `voice.occupancy.updated`.
  - Kafka producer `backend/ms-core/internal/kafka/producers/voice_events.go` publishing to `chat-events`.
- **Files/areas touched:**
  - `backend/ms-core/internal/kafka/producers/voice_events.go`
  - `backend/ms-core/internal/kafka/producers/test/voice_events_test.go`
  - `backend/ms-core/internal/voice/service/webhook.go`
  - `backend/ms-core/internal/voice/service/test/webhook_test.go`
  - `backend/ms-core/internal/server/server_run.go`
- **Dependencies:** `M1-2`
- **Size:** **M** (5 files, ~380 lines)
- **Verification step:** `make -C backend/ms-core check` (unit tests covering signed vs forged webhooks, out-of-order/duplicate events, 26th video publisher rejection, and session duration calculation).

#### `M1-4`: `api-gateway` Voice REST Routes, Webhook Forwarding & Realtime WebSocket Fanout
- **Owning area:** `api-gateway`
- **What gets built:**
  - Add `CoreHost` (`MS_CORE_HOST`, default `localhost:3003`) to `backend/api-gateway/internal/config/config.go` and `.env.example`.
  - Create `backend/api-gateway/internal/voice/delivery/http/{handler.go,routes.go}` and `pkg/voice-http/voice_http.go`:
    - `POST /api/v1/channels/{channelId}/voice/token` (and group-scoped alias `POST /api/v1/groups/{groupId}/channels/{channelId}/voice/token`)
    - `GET /api/v1/groups/{groupId}/voice/occupancy` (initial hydration of occupied voice channels for sidebar on load)
    - `POST /api/v1/webhooks/livekit` (forwards raw body + `Authorization` header to `ms-core.HandleWebhook`)
  - Update `backend/api-gateway/internal/kafka/consumers/realtime_fanout.go` to fan out `voice.occupancy.updated`, `voice.recording.updated`, `voice.moderation.applied`, and `voice.idle.warning` to both `hub.Deliver(groupID, channelID, payload)` and `hub.DeliverGroup(groupID, payload)` so **all group members see live sidebar occupancy without joining the room** (Spec §2.3 & §6.1).
- **Files/areas touched:**
  - `backend/api-gateway/internal/config/config.go` + `test/config_test.go`
  - `backend/api-gateway/pkg/voice-http/voice_http.go`
  - `backend/api-gateway/internal/voice/delivery/http/handler.go`
  - `backend/api-gateway/internal/voice/delivery/http/routes.go`
  - `backend/api-gateway/internal/kafka/consumers/realtime_fanout.go`
  - `backend/api-gateway/internal/server/server_handler.go`
  - `backend/api-gateway/internal/server/server_run.go`
  - `backend/tools/check-layout/main.go` (allow `voice-http` in `camelPkgDir`)
- **Dependencies:** `M1-3`
- **Size:** **M** (8 files, ~360 lines)
- **Verification step:** `make -C backend/api-gateway check && make -C backend check-layout`

#### `M1-5`: Frontend Contracts, API Service & LiveKit Room Context (`apps/web`)
- **Owning area:** `frontend` (`packages/contracts`, `packages/i18n`, `apps/web`)
- **What gets built:**
  - Add `@wichat/contracts/chat/voice.ts` (Zod schemas for token response, occupant state, moderation request, recording state, and realtime `voice.*` WebSocket events) and update `channel.ts` (`creatableChannelKindSchema` to include `"text_and_voice"`).
  - Add i18n keys in `frontend/packages/i18n/src/locales/en.json` and `id.json` for all voice UI states and error codes.
  - Install `livekit-client` and `@livekit/components-react` in `frontend/apps/web/package.json`.
  - Create `frontend/apps/web/src/services/voice/{voice.key.ts,voice.query.ts,voice.mutation.ts,voice.service.ts}`.
  - Create `frontend/apps/web/src/features/voice/voice-session-provider.tsx`: manages active `Room` instance from `livekit-client`, automatic reconnect + token refresh before 15m TTL expiry, local mute/deafen state persistence across reconnects, active speaker tracking (`RoomEvent.ActiveSpeakersChanged`), and WebAudio sound cues (join/leave/mute/deafen).
- **Files/areas touched:**
  - `frontend/packages/contracts/src/chat/voice.ts`
  - `frontend/packages/contracts/src/chat/channel.ts`
  - `frontend/packages/i18n/src/locales/{en.json,id.json}`
  - `frontend/apps/web/package.json`
  - `frontend/apps/web/src/services/voice/{voice.key.ts,voice.query.ts,voice.mutation.ts,voice.service.ts}`
  - `frontend/apps/web/src/features/voice/voice-session-provider.tsx`
- **Dependencies:** `M1-4`
- **Size:** **M** (8 files, ~440 lines)
- **Verification step:** `cd frontend && pnpm --filter web-client check`

#### `M1-6`: Sidebar Live Occupant List, `VoiceCallControls` Meet Strip & `ShellUserDock` Cleanup
- **Owning area:** `frontend` (`apps/web`)
- **What gets built:**
  - Remove the placeholder mic/headphone buttons from `frontend/apps/web/src/components/layout/shell-user-dock.tsx` so the profile bar strictly follows `docs/design/shell-ux.md`.
  - Build `frontend/apps/web/src/features/voice/components/voice-call-controls.tsx` (`VoiceCallControls` meet strip): mounted above `ShellUserDock` when a voice session is active; displays connection indicator (`connecting` / `connected` / `reconnecting` / `failed` with manual retry button), channel name, call duration timer, and buttons for Mute, Deafen (deafen implies mute), Camera, Screen Share, Record (when permitted), and Disconnect (red).
  - Build `frontend/apps/web/src/features/voice/components/sidebar-voice-occupants.tsx` and wire into `production-channel-sidebar.tsx`: renders live occupants under `voice_only` and `text_and_voice` channels (avatar with green speaking ring, bold name when speaking, mute/deafen/server-mute icons, plus room capacity badge e.g. `3/50`).
  - Subscribe to `voice.occupancy.updated` frames in `presence-client.ts` / `use-voice-occupancy-realtime.ts` so sidebar updates live for users who have not joined.
- **Files/areas touched:**
  - `frontend/apps/web/src/components/layout/shell-user-dock.tsx`
  - `frontend/apps/web/src/components/layout/workspace-panel.tsx`
  - `frontend/apps/web/src/features/voice/components/voice-call-controls.tsx`
  - `frontend/apps/web/src/features/voice/components/sidebar-voice-occupants.tsx`
  - `frontend/apps/web/src/features/voice/hooks/use-voice-occupancy-realtime.ts`
  - `frontend/apps/web/src/features/channel/components/production-channel-sidebar.tsx`
- **Dependencies:** `M1-5`
- **Size:** **M** (6 files, ~420 lines)
- **Verification step:** `cd frontend && pnpm --filter web-client check`

---

### Milestone 2 — Video, Screen Share, Voice Stage Grid/Focus & Settings

#### `M2-1`: Main Voice Stage (Tile Grid, Spotlight/Focus Layout, Split Text+Voice & DM Drop-in)
- **Owning area:** `frontend` (`apps/web`)
- **What gets built:**
  - Create `frontend/apps/web/src/features/voice/components/voice-stage-view.tsx` and `voice-participant-tile.tsx`:
    - **Grid View:** responsive participant tiles showing video stream (when camera or screen share is active) or large avatar with speaking ring (audio-only), mute/deafen/server-mute badges, and connection quality indicator. Screen shares and active speakers are prioritized in ordering.
    - **Focus / Spotlight View:** clicking any tile switches to a large primary stage tile + horizontal/vertical filmstrip of other participants; incoming screen shares auto-focus by default; supports multiple simultaneous screen shares where each viewer independently picks which share to focus (Spec §5.4, §6.4, §6.5).
    - **Empty & Error States:** empty room state (*"Nobody here yet — you're the first"*), locked channel (`!connect`), room full (`50/50`), quota reached, and browser mic/camera permission denied with clear step-by-step fix instructions (Spec §5.7, §6.7).
  - Integrate into `channel-chat-view.tsx`:
    - For `voice_only`: full-height `VoiceStageView` (with collapsible side text chat drawer if desired).
    - For `text_and_voice` and `dm`: collapsible top/split `VoiceStageView` header + drop-in "Join Voice" button in `ConversationHeader`, allowing simultaneous voice/video and text chat.
- **Files/areas touched:**
  - `frontend/apps/web/src/features/voice/components/voice-stage-view.tsx`
  - `frontend/apps/web/src/features/voice/components/voice-participant-tile.tsx`
  - `frontend/apps/web/src/features/voice/components/voice-empty-error-state.tsx`
  - `frontend/apps/web/src/features/chat/components/channel-chat-view.tsx`
  - `frontend/apps/web/src/features/chat/components/conversation-header.tsx`
- **Dependencies:** `M1-6`
- **Size:** **M** (5 files, ~470 lines)
- **Verification step:** `cd frontend && pnpm --filter web-client check`

#### `M2-2`: Adaptive Video/Screen Share Publishing (Simulcast, Dynacast, 720p/1080p & Audio Hint)
- **Owning area:** `frontend` (`apps/web`)
- **What gets built:**
  - Configure LiveKit `RoomOptions` in `voice-session-provider.tsx`: `dynacast: true`, `adaptiveStream: true`, camera default target `VideoPresets.h720` (respecting channel `video_quality` and `bitrate_kbps`), and screen share publish target `1080p @ 15–30fps` with `contentHint = "detail"` so IDE code is crisp on typical home connections (Spec §5.4, §11).
  - Build `frontend/apps/web/src/features/voice/components/screen-share-picker-hint.tsx`: surfaces a clear browser capability note when starting screen share (`system audio is supported on Chrome/Electron tab/screen share; other browsers may share video only`, Spec §5.4).
  - Enforce one screen share per participant while allowing multiple participants to share simultaneously.
- **Files/areas touched:**
  - `frontend/apps/web/src/features/voice/lib/livekit-media-options.ts`
  - `frontend/apps/web/src/features/voice/components/screen-share-picker-hint.tsx`
  - `frontend/apps/web/src/features/voice/voice-session-provider.tsx`
- **Dependencies:** `M2-1`
- **Size:** **S** (3 files, ~160 lines — or M within budget)
- **Verification step:** `cd frontend && pnpm --filter web-client check`

#### `M2-3`: User Settings → Voice & Video Section, Per-User Volume, VAD/PTT & Keybinds
- **Owning area:** `frontend` (`apps/web`)
- **What gets built:**
  - Replace the `"voice"` placeholder in `UserSettingsDialog` (`user-settings-dialog.tsx`) with `UserSettingsVoiceSection` (`frontend/apps/web/src/features/user-settings/components/sections/user-settings-voice-section.tsx`):
    - Input/output audio device pickers + camera picker (mid-call hot-switching via `room.switchActiveDevice` without reconnecting).
    - Mic test with live WebAudio level meter + local listen-back toggle.
    - Input mode selector: **Voice Activity (VAD)** with adjustable sensitivity threshold slider (`-80 dB` to `0 dB` WebAudio analyser gate) vs **Push-to-Talk (PTT)** with configurable keybind recorder (works in web & Electron).
    - Built-in libwebrtc audio processing toggles: Echo Cancellation (`echoCancellation`), Noise Suppression (`noiseSuppression`), Auto Gain Control (`autoGainControl`) (Spec §5.3, §6.6).
    - Camera live preview card.
  - Build participant right-click context menu (`voice-participant-context-menu.tsx`) for both sidebar occupants and stage tiles: client-side **Per-User Volume slider (0–200%)** (via WebAudio `GainNode` boost for >100%) and **Local Mute** toggle persisted per target user in `localStorage` (Spec §5.2, §6.1).
- **Files/areas touched:**
  - `frontend/apps/web/src/features/user-settings/components/sections/user-settings-voice-section.tsx`
  - `frontend/apps/web/src/features/user-settings/components/user-settings-dialog.tsx`
  - `frontend/apps/web/src/features/voice/components/voice-participant-context-menu.tsx`
  - `frontend/apps/web/src/features/voice/lib/voice-preferences-storage.ts`
  - `frontend/apps/web/src/features/voice/hooks/use-voice-input-processor.ts`
- **Dependencies:** `M2-2`
- **Size:** **M** (5 files, ~460 lines)
- **Verification step:** `cd frontend && pnpm --filter web-client check`

---

### Milestone 3 — Moderation, Presence Activity Status & Idle-Kick

#### `M3-1`: Server-Authoritative Moderation (`mute`, `deafen`, `disconnect`)
- **Owning area:** `ms-core` / `api-gateway` / `frontend`
- **What gets built:**
  - Implement `ModerateParticipant` RPC in `ms-core`:
    - Re-verify caller's effective permissions on the channel via `ms-user.AssertGroupChannelAccess`:
      - `mute` / `unmute` requires `mute_members` or `administrator`.
      - `deafen` / `undeafen` requires `deafen_members` or `administrator`.
      - `disconnect` requires `move_members`, `kick_members`, or `administrator`.
    - Execute server-side SFU enforcement via LiveKit `RoomServiceClient`:
      - Server-mute: `UpdateParticipant` (revoke audio `CanPublish` grant / `MutePublishedTrack`) so the client cannot bypass it; update `server_muted` in Redis & `call_participants`.
      - Server-deafen: `UpdateParticipant` (revoke `CanSubscribe` + audio `CanPublish`); update `server_deafened` in Redis & `call_participants`.
      - Disconnect: `RemoveParticipant(room, targetUserUUID)`.
    - Broadcast `voice.occupancy.updated` and `voice.moderation.applied` over Kafka → Gateway WS so all clients reflect the change in **< 2 seconds** (Spec §11).
  - Add gateway endpoint `POST /api/v1/channels/{channelId}/voice/moderate`.
  - Update frontend context menu (`voice-participant-context-menu.tsx`) and `VoiceCallControls`: show server-mute/server-deafen controls to permitted moderators; if the local user is server-muted/deafened, lock the self-unmute button with a clear red icon + tooltip (*"Server muted by a moderator"*).
- **Files/areas touched:**
  - `backend/ms-core/internal/voice/service/moderation.go`
  - `backend/ms-core/internal/voice/service/test/moderation_test.go`
  - `backend/api-gateway/internal/voice/delivery/http/handler.go`
  - `backend/api-gateway/internal/voice/delivery/http/routes.go`
  - `frontend/apps/web/src/features/voice/components/voice-participant-context-menu.tsx`
  - `frontend/apps/web/src/features/voice/components/voice-call-controls.tsx`
- **Dependencies:** `M1-4`, `M2-3`
- **Size:** **M** (6 files, ~380 lines)
- **Verification step:** `make -C backend check && cd frontend && pnpm --filter web-client check`

#### `M3-2`: Automatic Voice/Video Activity Status ("In a call" / "In a meeting")
- **Owning area:** `ms-user` / `ms-core` / `frontend`
- **What gets built:**
  - Per Spec §5.6 and System Design §6.6, joining a voice room sets automatic activity status to `"In a call"` (or `"In a meeting"` when camera/screenshare is active), without overwriting a user's manual `status_note_text`, and restores prior state on leave.
  - Add `SetVoiceActivityStatus` RPC to `user.v1.PresenceService` in `ms-user` (stores ephemeral `presence:voice_activity:{user_uuid}` in Redis with session TTL and includes `activity_status` in `presence.updated` broadcasts and member snapshots when no manual `status_note_text` overrides it).
  - Call `SetVoiceActivityStatus` from `ms-core` on `participant_joined`, `track_published`/`track_unpublished` (video transition), and `participant_left`/`room_finished`.
  - Render the automatic activity status badge (`"In a call"` / `"In a meeting"`) in `ShellUserDock` and `MemberPanel` whenever the user has no manual status note set.
- **Files/areas touched:**
  - `backend/wi-shared/models/proto/user/v1/presence.proto`
  - `backend/ms-user/internal/presence/repository/redis_presence.go` + `test/redis_presence_test.go`
  - `backend/ms-user/internal/presence/service/service.go` + `test/service_test.go`
  - `backend/ms-user/internal/presence/delivery/grpc/server.go`
  - `backend/ms-core/internal/voice/service/webhook.go`
  - `frontend/apps/web/src/components/layout/shell-user-dock.tsx`
- **Dependencies:** `M1-3`
- **Size:** **M** (8 files, ~340 lines)
- **Verification step:** `make -C backend check && cd frontend && pnpm --filter web-client check`

#### `M3-3`: Idle-Kick Worker (30-Min Inactivity, 2-Min Warning, Activity Reset)
- **Owning area:** `ms-core` / `api-gateway` / `frontend`
- **What gets built:**
  - Per Spec §8 and Locked Decision #4: protect Cloud Build participant-minutes from forgotten tabs.
  - Background ticker in `ms-core` (`internal/voice/service/idle_monitor.go`, running every 30s):
    - Tracks `last_active_at` per occupant in Redis (updated on join, track publish/unpublish, active speaker activity, or client heartbeat/ack).
    - At **28 minutes** of inactivity (no audio/video publishing and no interaction), emits `voice.idle.warning` (`expires_in_seconds: 120`) to Kafka → Gateway WS.
    - At **30 minutes** of inactivity, calls `LiveKit.RemoveParticipant` and records `disconnect_reason = "idle_timeout"` on `call_participants`.
  - Gateway endpoint `POST /api/v1/channels/{channelId}/voice/keepalive` (`AcknowledgeIdleWarning` RPC) resets `last_active_at` in Redis.
  - Frontend modal/banner (`voice-idle-warning-banner.tsx`): displays a 2-minute countdown when `voice.idle.warning` arrives; clicking *"Stay in Call"*, unmuting, or speaking calls `/voice/keepalive` and dismisses the warning.
- **Files/areas touched:**
  - `backend/ms-core/internal/voice/service/idle_monitor.go`
  - `backend/ms-core/internal/voice/service/test/idle_monitor_test.go`
  - `backend/api-gateway/internal/voice/delivery/http/handler.go`
  - `backend/api-gateway/internal/voice/delivery/http/routes.go`
  - `frontend/apps/web/src/features/voice/components/voice-idle-warning-banner.tsx`
- **Dependencies:** `M3-1`
- **Size:** **M** (5 files, ~320 lines)
- **Verification step:** `make -C backend/ms-core check && cd frontend && pnpm --filter web-client check` (unit tests with fake clock verifying 28m warning, activity reset, and 30m removal).

---

### Milestone 4 — Call Recording (LiveKit Egress → MinIO)

#### `M4-1`: `ms-core` LiveKit Egress Recording Controller & Webhook Lifecycle
- **Owning area:** `ms-core`
- **What gets built:**
  - Extend `backend/ms-core/pkg/livekit-client/client.go` with `StartRoomCompositeEgress` and `StopEgress` targeting MinIO (`S3Upload` configured with `MINIO_ENDPOINT`, `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`, `MINIO_BUCKET`, `ForcePathStyle: true`, output key `recordings/{group_id}/{channel_id}/{session_id}/{timestamp}.mp4`).
  - Implement `StartRecording`, `StopRecording`, and `ListChannelRecordings` in `ms-core/internal/voice/service/recording.go`:
    - Check `record_call` (or `administrator` / group owner) via `ms-user.AssertGroupChannelAccess`.
    - Prevent concurrent recordings on the same session (`voice.recording_already_active`).
    - Insert `call_recordings` row (`status = 'starting'`), set `is_recording = true` in Redis occupancy state, and emit `voice.recording.updated` event.
    - Handle LiveKit `egress_started`, `egress_updated`, and `egress_ended` webhooks: update `call_recordings` (`status = 'active' | 'completed' | 'failed'`, `file_size_bytes`, `duration_seconds`, `ended_at`), clear `is_recording` in Redis on completion, and broadcast `voice.recording.updated`.
  - Add presigned GET URL generation via `wi-shared/infra/minio` (or `ms-core` MinIO repository) so authorized channel members can download/play completed recordings, and link retention cleanup to workspace `retention_days`.
- **Files/areas touched:**
  - `backend/ms-core/pkg/livekit-client/client.go` + `test/client_test.go`
  - `backend/ms-core/internal/voice/repository/pg_repository.go` + `mock/pg_repository_mock.go`
  - `backend/ms-core/internal/voice/service/recording.go` + `test/recording_test.go`
  - `backend/ms-core/internal/voice/delivery/grpc/server.go`
- **Dependencies:** `M0-1`, `M1-3`
- **Size:** **M** (7 files, ~430 lines)
- **Verification step:** `make -C backend/ms-core check`

#### `M4-2`: Gateway Recording Endpoints & Mandatory Recording UI (Indicator + Pre-Join Notice)
- **Owning area:** `api-gateway` / `frontend`
- **What gets built:**
  - Add gateway routes in `backend/api-gateway/internal/voice/delivery/http/`:
    - `POST /api/v1/channels/{channelId}/voice/recording/start`
    - `POST /api/v1/channels/{channelId}/voice/recording/stop`
    - `GET /api/v1/channels/{channelId}/voice/recordings`
  - Frontend mandatory transparency UI (Spec §5.8):
    - **Active Call Indicator:** non-dismissible pulsing red dot + `"Recording"` badge in both the `VoiceStageView` header and `VoiceCallControls` meet strip visible to **all** participants whenever `is_recording` is true.
    - **Pre-Join Recording Notice:** when a user clicks a channel where `is_recording === true` in sidebar occupancy (or `IssueVoiceToken` returns `is_recording: true`), display a confirmation dialog (*"This call is currently being recorded. By joining, you consent to being recorded."*) before connecting the LiveKit room.
    - **Start/Stop Control:** Record button in `VoiceCallControls` for users with `record_call` / `administrator` permission, plus `record_call` toggle in Group Role & Channel Permission settings UI.
- **Files/areas touched:**
  - `backend/api-gateway/internal/voice/delivery/http/{handler.go,routes.go}`
  - `frontend/apps/web/src/features/voice/components/voice-recording-badge.tsx`
  - `frontend/apps/web/src/features/voice/components/voice-recording-consent-dialog.tsx`
  - `frontend/apps/web/src/features/voice/components/voice-call-controls.tsx`
  - `frontend/apps/web/src/features/group-settings/lib/group-settings-roles-view.ts`
- **Dependencies:** `M4-1`
- **Size:** **M** (6 files, ~350 lines)
- **Verification step:** `make -C backend/api-gateway check && cd frontend && pnpm --filter web-client check`

---

### Milestone 5 — Quota Monitoring, Observability & End-to-End Acceptance

#### `M5-1`: Monthly Participant-Minutes Quota Ledger & 80% Admin Warning Banner
- **Owning area:** `ms-core` / `api-gateway` / `frontend`
- **What gets built:**
  - Implement `GetQuotaUsage` in `ms-core`: sums `total_participant_seconds` from `call_sessions` for the current UTC calendar month (plus active session elapsed time from Redis) against configurable `LIVEKIT_MONTHLY_QUOTA_MINUTES` (default `5000` for Cloud Build plan).
  - Reject new room creations when `used_minutes >= 100%` of hard cap (when `LIVEKIT_ENFORCE_HARD_CAP=true`) with `voice.quota_exceeded`.
  - Expose `GET /api/v1/voice/quota` on `api-gateway`.
  - Render a dismissible Admin/Owner warning banner in the workspace shell (`voice-quota-banner.tsx`) when monthly usage reaches **≥ 80% (4,000 / 5,000 participant-minutes)** advising upgrade to LiveKit Ship or switching `LIVEKIT_URL` to self-hosted LiveKit (Spec §8).
- **Files/areas touched:**
  - `backend/ms-core/internal/voice/service/quota.go` + `test/quota_test.go`
  - `backend/api-gateway/internal/voice/delivery/http/{handler.go,routes.go}`
  - `frontend/apps/web/src/features/voice/components/voice-quota-banner.tsx`
  - `frontend/apps/web/src/features/shell/shell-layout.tsx`
- **Dependencies:** `M1-3`
- **Size:** **M** (6 files, ~280 lines)
- **Verification step:** `make -C backend check && cd frontend && pnpm --filter web-client check`

#### `M5-2`: Playwright E2E & Integration Verification Suite for Voice & Video
- **Owning area:** `frontend` (`apps/web/e2e`) / `ms-core`
- **What gets built:**
  - Add Playwright E2E spec `frontend/apps/web/e2e/channels/voice-call.spec.ts` using Chromium fake media flags (`--use-fake-ui-for-media-stream`, `--use-fake-device-for-media-stream`):
    - Two browser contexts join a `text_and_voice` channel drop-in; verify bidirectional audio track state, sidebar occupancy for a 3rd unjoined user, speaking ring, server-mute + disconnect landing in <2s, activity status switching to `"In a call"` and restoring on leave, and recording indicator visibility.
- **Files/areas touched:**
  - `frontend/apps/web/playwright.config.ts`
  - `frontend/apps/web/e2e/channels/voice-call.spec.ts`
- **Dependencies:** `M1`–`M5-1`
- **Size:** **M** (2 files, ~260 lines)
- **Verification step:** `cd frontend/apps/web && pnpm test:e2e e2e/channels/voice-call.spec.ts`

---

## 3. Contracts

### 3.1 `ms-core` gRPC / Protobuf Surface (`backend/wi-shared/models/proto/core/v1/voice.proto`)

- **Package:** `core.v1` (`go_package = "github.com/wichat/wichat/backend/wi-shared/models/gen/core/v1;corev1"`)
- **Service `GeneralService`** (`general.proto`):
  - `Health(HealthRequest) returns (HealthResponse)` — standard liveness probe matching `ms-user` and `ms-chat`.
- **Service `VoiceService`** (`voice.proto`):
  1. `IssueVoiceToken(IssueVoiceTokenRequest) returns (IssueVoiceTokenResponse)`
     - **Request:** `access_token`, `group_id` (optional for DM), `channel_id` (24-char public ID).
     - **Response:** `token` (signed LiveKit JWT), `livekit_url` (`wss://...`), `room_name` (`channel:<uuid>`), `channel_id`, `expires_at` (RFC3339), `initial_server_muted` (`bool`), `can_publish_audio` (`bool`), `can_publish_video` (`bool`), `use_voice_activity` (`bool`), `bitrate_kbps` (`int32`), `video_quality` (`string`), `is_recording` (`bool`), `occupants` (`repeated VoiceOccupant`).
  2. `ModerateParticipant(ModerateParticipantRequest) returns (ModerateParticipantResponse)`
     - **Request:** `access_token`, `group_id`, `channel_id`, `target_user_id` (UUID), `action` (`"mute" | "unmute" | "deafen" | "undeafen" | "disconnect"`), `reason` (optional).
     - **Response:** `ok` (`bool`), `occupant` (`VoiceOccupant`).
  3. `HandleWebhook(HandleWebhookRequest) returns (HandleWebhookResponse)`
     - **Request:** `auth_header` (`Authorization` header from LiveKit), `raw_body` (`bytes`).
     - **Response:** `ok` (`bool`).
  4. `ListGroupOccupancy(ListGroupOccupancyRequest) returns (ListGroupOccupancyResponse)`
     - **Request:** `access_token`, `group_id`.
     - **Response:** `channels` (`repeated ChannelOccupancy` — `channel_id`, `is_recording`, `occupants`).
  5. `StartRecording(StartRecordingRequest) returns (StartRecordingResponse)`
     - **Request:** `access_token`, `group_id`, `channel_id`.
     - **Response:** `recording` (`CallRecordingSummary`: `id`, `session_id`, `egress_id`, `status`, `started_by`, `started_at`).
  6. `StopRecording(StopRecordingRequest) returns (StopRecordingResponse)`
     - **Request:** `access_token`, `group_id`, `channel_id`, `recording_id` (optional; defaults to active recording in channel).
     - **Response:** `recording` (`CallRecordingSummary`).
  7. `ListChannelRecordings(ListChannelRecordingsRequest) returns (ListChannelRecordingsResponse)`
     - **Request:** `access_token`, `group_id`, `channel_id`, `limit`, `cursor`.
     - **Response:** `recordings` (`repeated CallRecordingSummary` including presigned `download_url`, `file_size_bytes`, `duration_seconds`), `next_cursor`.
  8. `AcknowledgeIdleWarning(AcknowledgeIdleWarningRequest) returns (AcknowledgeIdleWarningResponse)`
     - **Request:** `access_token`, `group_id`, `channel_id`.
     - **Response:** `ok` (`bool`).
  9. `GetQuotaUsage(GetQuotaUsageRequest) returns (GetQuotaUsageResponse)`
     - **Request:** `access_token`, `group_id`.
     - **Response:** `month` (`"YYYY-MM"`), `used_participant_minutes` (`int64`), `limit_participant_minutes` (`int64`), `usage_percent` (`double`), `warning_active` (`bool`).

### 3.2 REST Endpoints via `api-gateway` (`/api/v1`)

All client responses are wrapped in `response.NewDataEnvelope(...)` (`{ "data": ... }`) and errors use `{ "error": { "code": "<i18n_key>", "params": { ... } } }`.

| Method & Path | Auth | Rate Limit | Description |
|---|---|---|---|
| `POST /api/v1/channels/{channelId}/voice/token` *(plus `/api/v1/groups/{groupId}/channels/{channelId}/voice/token`)* | Bearer JWT | 10 req/min/user | Issues a 15-min LiveKit JWT after checking membership, `connect` permission, and 50-user cap. |
| `POST /api/v1/channels/{channelId}/voice/moderate` | Bearer JWT | 60 req/min/user | Executes `mute`, `unmute`, `deafen`, `undeafen`, or `disconnect` on `target_user_id`. |
| `GET /api/v1/groups/{groupId}/voice/occupancy` | Bearer JWT | 120 req/min/user | Returns live occupancy map + recording flag for all voice-capable channels in the group. |
| `POST /api/v1/channels/{channelId}/voice/keepalive` | Bearer JWT | 30 req/min/user | Resets the caller's 30-min idle-kick timer (`AcknowledgeIdleWarning`). |
| `POST /api/v1/channels/{channelId}/voice/recording/start` | Bearer JWT | 10 req/min/user | Starts LiveKit Egress MP4 recording to MinIO (requires `record_call` or `administrator`). |
| `POST /api/v1/channels/{channelId}/voice/recording/stop` | Bearer JWT | 10 req/min/user | Stops active LiveKit Egress recording on the channel. |
| `GET /api/v1/channels/{channelId}/voice/recordings` | Bearer JWT | 60 req/min/user | Lists completed/active recordings for the channel with short-lived presigned MinIO URLs. |
| `GET /api/v1/voice/quota` | Bearer JWT | 60 req/min/user | Returns current month's participant-minutes usage vs Cloud Build 5,000-min cap. |
| `POST /api/v1/webhooks/livekit` | LiveKit `Authorization` JWT | Public (Signed) | Receives LiveKit Cloud/Server webhooks and forwards raw body + signature to `ms-core`. |

### 3.3 Webhook Event Handling & Outbound Realtime Events

**Inbound LiveKit Webhooks (`POST /api/v1/webhooks/livekit` → `ms-core`):**
- Verified using `livekit/protocol/webhook.ReceiveWebhookEvent(r, authProvider)` which validates the `Authorization` JWT and SHA-256 payload hash.
- Deduplicated via Redis key `voice:webhook:dedup:{event_id}` (TTL `24h`).
- Handled events:
  - `room_started` → Upsert active `call_sessions` row; initialize Redis session metadata.
  - `participant_joined` → Add occupant to Redis hash; insert `call_participants`; update `peak_participants`; set user activity status `"In a call"`; publish `voice.occupancy.updated`.
  - `participant_left` → Remove occupant from Redis hash; close `call_participants` (`left_at`, `duration_seconds`); increment `call_sessions.total_participant_seconds`; restore user activity status; publish `voice.occupancy.updated`.
  - `track_published` / `track_unpublished` → Update occupant track flags (`is_camera_on`, `is_screen_sharing`, `is_muted`); enforce ≤25 active video publishers; update activity status to/from `"In a meeting"`; publish `voice.occupancy.updated`.
  - `room_finished` → Close `call_sessions.ended_at` and any open `call_participants`; clean up Redis room keys; publish `voice.occupancy.updated` (empty occupants).
  - `egress_started` / `egress_updated` / `egress_ended` → Update `call_recordings` status, file size, and duration; update Redis `is_recording`; publish `voice.recording.updated`.

**Outbound Realtime WebSocket Events (via Kafka `chat-events` → `api-gateway` `hub.DeliverGroup` & `hub.Deliver`):**
- `voice.occupancy.updated`:
  - Fields: `type: "voice.occupancy.updated"`, `group_id`, `channel_id`, `is_recording: bool`, `occupants: [{ user_id, display_name, username, avatar_url, is_muted, is_deafened, is_server_muted, is_server_deafened, is_camera_on, is_screen_sharing, joined_at }]`.
- `voice.recording.updated`:
  - Fields: `type: "voice.recording.updated"`, `group_id`, `channel_id`, `is_recording: bool`, `recording_id`, `started_by_user_id`, `started_at`.
- `voice.moderation.applied`:
  - Fields: `type: "voice.moderation.applied"`, `group_id`, `channel_id`, `target_user_id`, `action`, `actor_user_id`.
- `voice.idle.warning`:
  - Fields: `type: "voice.idle.warning"`, `group_id`, `channel_id`, `user_id`, `expires_in_seconds: 120`.

### 3.4 Redis Occupancy Schema (Owned by `ms-core`)

| Key Pattern | Type | Value / Fields | TTL |
|---|---|---|---|
| `voice:room:meta:{channel_uuid}` | `HASH` | `session_id`, `group_public_id`, `channel_public_id`, `room_name`, `started_at`, `is_recording`, `active_recording_id`, `video_publisher_count` | `12h` (refreshed on webhook/activity; deleted on `room_finished`) |
| `voice:room:occupants:{channel_uuid}` | `HASH` | Field: `{user_uuid}` → JSON `{ user_id, display_name, username, avatar_url, is_muted, is_deafened, is_server_muted, is_server_deafened, is_camera_on, is_screen_sharing, joined_at, last_active_at, idle_warned }` | `12h` (deleted on `room_finished`) |
| `voice:group:channels:{group_public_id}` | `SET` | Member: `{channel_uuid}` of currently occupied channels in that group | `12h` |
| `voice:user:active_room:{user_uuid}` | `STRING` | `{channel_uuid}` (tracks which room the user is currently in for activity status & single-room enforcement) | `2h` |
| `voice:ratelimit:token:{user_uuid}` | `STRING` | Counter (`INCR`) for token issuance rate limiting (max 10 / 60s) | `60s` |
| `voice:webhook:dedup:{event_id}` | `STRING` | `"1"` (prevents duplicate webhook processing) | `24h` |

### 3.5 Database Migrations (`00049_voice_schema_and_permissions.sql`)

Planned relational shapes in PostgreSQL (`public` schema, owned by `ms-core`):

1. **`channels` constraint update:**
   - `channels_type_check`: `CHECK (type IN ('broadcast', 'text_only', 'voice_only', 'text_and_voice', 'forum', 'dm'))`.
2. **`call_sessions` table:**
   - `id uuid PRIMARY KEY DEFAULT gen_random_uuid()`
   - `public_id text NOT NULL UNIQUE CHECK (public_id ~ '^[0-9a-f]{24}$')`
   - `group_id uuid REFERENCES groups(id) ON DELETE CASCADE` *(nullable for future global DMs, indexed)*
   - `channel_id uuid NOT NULL REFERENCES channels(id) ON DELETE CASCADE`
   - `room_name text NOT NULL`
   - `started_at timestamptz NOT NULL DEFAULT now()`
   - `ended_at timestamptz`
   - `peak_participants integer NOT NULL DEFAULT 1`
   - `total_participant_seconds bigint NOT NULL DEFAULT 0`
   - Indexes: `(channel_id, started_at DESC)`, `(started_at)` for monthly quota aggregation, partial index `WHERE ended_at IS NULL` for active sessions.
3. **`call_participants` table:**
   - `id uuid PRIMARY KEY DEFAULT gen_random_uuid()`
   - `session_id uuid NOT NULL REFERENCES call_sessions(id) ON DELETE CASCADE`
   - `user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE`
   - `joined_at timestamptz NOT NULL DEFAULT now()`
   - `left_at timestamptz`
   - `duration_seconds integer NOT NULL DEFAULT 0`
   - `server_muted boolean NOT NULL DEFAULT false`
   - `server_deafened boolean NOT NULL DEFAULT false`
   - `disconnect_reason text NOT NULL DEFAULT ''` *(`'' | 'left' | 'kicked' | 'idle_timeout' | 'room_closed'`)*
   - Indexes: `(session_id, user_id)`, `(user_id, joined_at DESC)`.
4. **`call_recordings` table:**
   - `id uuid PRIMARY KEY DEFAULT gen_random_uuid()`
   - `public_id text NOT NULL UNIQUE CHECK (public_id ~ '^[0-9a-f]{24}$')`
   - `session_id uuid NOT NULL REFERENCES call_sessions(id) ON DELETE CASCADE`
   - `channel_id uuid NOT NULL REFERENCES channels(id) ON DELETE CASCADE`
   - `egress_id text NOT NULL UNIQUE`
   - `status text NOT NULL CHECK (status IN ('starting', 'active', 'stopping', 'completed', 'failed'))`
   - `file_key text NOT NULL DEFAULT ''`
   - `file_size_bytes bigint NOT NULL DEFAULT 0`
   - `duration_seconds integer NOT NULL DEFAULT 0`
   - `started_by uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT`
   - `stopped_by uuid REFERENCES users(id) ON DELETE SET NULL`
   - `started_at timestamptz NOT NULL DEFAULT now()`
   - `ended_at timestamptz`
   - Indexes: `(channel_id, started_at DESC)`, `(session_id)`.

---

## 4. Risks & Mitigations

| # | Risk | Impact | Mitigation in Plan |
|---|---|---|---|
| **R1** | **LiveKit Cloud Build Free Quota Exhaustion** (5,000 participant-minutes/month ≈ 83 participant-hours, 100 concurrent connections, 50 GB downstream/month). Forgotten open tabs or heavy video calls can exhaust the monthly pool mid-month. | High (calls blocked until the 1st of next month) | 1. **30-min Idle-Kick (`M3-3`)** with 2-min warning removes inactive tabs automatically.<br>2. **Room/Video Caps (`M1-2`, `M1-3`)**: hard cap of 50 participants and 25 video publishers + `dynacast` & `simulcast` (`M2-2`) so unsubscribed video streams consume 0 downstream bytes.<br>3. **Monthly Quota Ledger (`M5-1`)**: compares webhook participant-seconds against the 5,000-min budget and triggers an Admin warning banner at 80% (4,000 mins).<br>4. **Zero-code Scale Path**: switching `LIVEKIT_URL`, `LIVEKIT_API_KEY`, `LIVEKIT_API_SECRET` in `ms-core/.env` points tokens to Cloud Ship or self-hosted LiveKit immediately. |
| **R2** | **LiveKit Cloud Build Egress / Recording Metering & Availability** (Uncorroborated in repo docs whether Cloud Build free tier permits composite egress or how egress minutes count against the 5,000-min pool, plus LiveKit Cloud Egress needs a publicly reachable S3 endpoint if writing to MinIO). | High (recording is a locked v1 requirement) | **Explicit Early Verification Task (`M0-1`)** runs before backend coding: verifies LiveKit Cloud Build Egress behavior and S3 reachability. For local dev and self-hosted MinIO behind NAT (where LiveKit Cloud cannot reach `localhost:9000`), `M0-1` provides `deploy/compose/livekit.yml` (`livekit-server` + `livekit-egress` on the local Docker network) or tunnel/S3 config so recording works reliably in both dev and production. |
| **R3** | **Webhook Reliability, Ordering & Ghost Occupants** (Dropped webhooks during `ms-core` restart or out-of-order `participant_left` / `participant_joined` leaving stale avatars in sidebar and inflating quota ledger). | High | 1. **Idempotency & Timestamp Ordering (`M1-3`)**: deduplicate by `event.id`; compare event `created_at` against occupant `joined_at`.<br>2. **Periodic Reconciliation (`M3-3`)**: the 30s background monitor in `ms-core` cross-checks Redis occupied rooms against `LiveKit.ListParticipants` / `ListRooms` and purges any ghost sessions whose LiveKit room no longer exists. |
| **R4** | **Token Security & Permission Escalation** (Long-lived JWTs reused after a user is kicked, banned, or has `speak`/`video` revoked). | High | 1. **Short 15-min JWT TTL (`M1-2`)**: forces periodic token refresh via `IssueVoiceToken` which re-evaluates group membership and channel permissions.<br>2. **Immediate SFU Revocation (`M3-1`)**: server-mute, deafen, channel permission change, or group kick/ban immediately calls LiveKit `UpdateParticipant` / `RemoveParticipant` from `ms-core`.<br>3. **Secrets Isolation**: `LIVEKIT_API_KEY` and `LIVEKIT_API_SECRET` exist only in `ms-core` env. |
| **R5** | **Browser Screen-Share System Audio Limits** (macOS/Linux browsers and Firefox/Safari often cannot capture system audio on window shares; only Chrome/Electron tab or full-screen shares support audio). | Medium (user confusion) | **Capability Detection & Picker Hint (`M2-2`)**: detect `navigator.userAgent` / platform in `screen-share-picker-hint.tsx` and display a clear inline notice explaining when system audio is available (Chrome tab / Electron / Windows screen share) vs video-only. |
| **R6** | **`check-layout` & Scoped Coverage Gate Failures** (`backend/tools/check-layout` rejects non-whitelisted repo filenames, missing `mock/pg_repository_mock.go`, or non-1:1 test files; `coverage-gate.sh` requires ≥90%). | Medium (CI/pre-push block) | Every backend task in §2 explicitly names whitelisted repository files (`pg_repository.go`, `redis_occupancy.go`), includes 1:1 `test/<file>_test.go` files, uses `miniredis` for Redis tests, and runs `make -C backend check` as its gate. |

---

## 5. Testing & Acceptance Mapping

Every acceptance criterion in Spec §11 is mapped below to its implementing tasks and concrete verification check:

| # | Spec §11 Acceptance Criterion | Implementing Tasks | Concrete Test / Verification Step |
|---|---|---|---|
| **AC-1** | Two users in a `text_and_voice` channel can hold a drop-in voice call from web (Chrome) and Electron. | `M0-2`, `M1-2`, `M1-5`, `M1-6`, `M2-1`, `M5-2` | **E2E (`M5-2`):** Playwright test creates a `text_and_voice` channel, joins from two browser contexts with fake media streams, asserts LiveKit connection state `connected`, audio track publication/subscription, and `VoiceCallControls` meet strip active. |
| **AC-2** | A user without `speak` joins server-muted; without `video` they cannot publish camera/share. | `M1-2`, `M2-1`, `M5-2` | **Unit (`M1-2`) + E2E (`M5-2`):** `service_test.go` verifies `IssueVoiceToken` sets `CanPublishSources = []` (no mic) and `initial_server_muted = true` when `speak` is absent, and omits `CAMERA`/`SCREEN_SHARE` when `video` is absent. E2E verifies Camera/Share buttons are disabled and mic is locked server-muted. |
| **AC-3** | An admin can server-mute and disconnect a member; the effect lands on all clients in under 2 seconds. | `M3-1`, `M5-2` | **Unit (`M3-1`) + E2E (`M5-2`):** `moderation_test.go` verifies `ModerateParticipant` calls `UpdateParticipant`/`RemoveParticipant` and emits `voice.moderation.applied`. E2E measures wall-clock time from Admin clicking "Server Mute" / "Disconnect" to Member UI reflecting locked mute / room exit (`< 2000ms`). |
| **AC-4** | The sidebar shows room occupants live for users who have not joined. | `M1-3`, `M1-4`, `M1-6`, `M5-2` | **Unit (`M1-3`, `M1-4`) + E2E (`M5-2`):** User A joins voice; User B (viewing `#general` text channel, not joined to voice) receives `voice.occupancy.updated` over `/api/v1/realtime` WS and sees User A's avatar + speaking/mute state under the voice channel in `ProductionChannelSidebar`. |
| **AC-5** | Screen share text (IDE code) is readable on a typical home connection. | `M2-2` | **Unit (`M2-2`) + Manual/E2E:** Verify `livekit-media-options.ts` sets screen share encoding to `1080p` (`maxBitrate: 3_000_000`, `maxFramerate: 15/30`) and `videoTrack.contentHint = "detail"`, with `simulcast` and `dynacast` enabled. |
| **AC-6** | Reconnecting after a 15-second network drop restores audio without a manual rejoin. | `M1-5`, `M5-2` | **E2E (`M5-2`):** In `voice-call.spec.ts`, toggle Playwright `context.setOffline(true)` for 15s (or simulate WebRTC transport drop), verify `VoiceCallControls` shows `reconnecting` → `connected`, token refreshes if needed, and local `muted`/`deafened` state + audio track publish resume automatically. |
| **AC-7** | Activity status switches to "In a call" on join and is restored on leave. | `M3-2`, `M5-2` | **Unit (`M3-2`) + E2E (`M5-2`):** `presence/service/test/service_test.go` verifies `SetVoiceActivityStatus` sets `"In a call"` (or `"In a meeting"` with video), yields to any manual `status_note_text`, and restores prior status on leave. E2E verifies `ShellUserDock` and `MemberPanel` status text transition. |
| **AC-8** | Recorded participant-minutes match the LiveKit dashboard (±5%). | `M1-1`, `M1-3`, `M5-1` | **Unit (`M1-3`, `M5-1`) + Integration:** `webhook_test.go` and `quota_test.go` verify `call_participants.duration_seconds` and `call_sessions.total_participant_seconds` computed from `participant_joined`/`participant_left` timestamps + active Redis sessions match simulated session durations within <1% drift. |
| **AC-9** | A host/admin can start/stop recording; all participants see the recording indicator, and the resulting file lands in MinIO linked to the session record. | `M0-1`, `M4-1`, `M4-2`, `M5-2` | **Unit (`M4-1`) + E2E (`M5-2`):** `recording_test.go` verifies `StartRecording` checks `record_call` permission, invokes LiveKit Egress with MinIO S3 config, persists `call_recordings` linked to `call_sessions.id`, and broadcasts `voice.recording.updated`. E2E verifies red `"Recording"` badge on all participants, pre-join consent modal for late joiners, and completed file metadata in `GET /api/v1/channels/{id}/voice/recordings`. |

---

## 6. Open Items

Per the protocol, no product decisions have been invented where the spec or repo leaves a gap. The following items are flagged for product/engineering confirmation before or during execution:

1. **LiveKit Cloud Build Plan Egress Availability & S3 Reachability (`M0-1`):**
   - *Context:* Spec §5.8 notes: *"how egress minutes/data are counted on the Cloud Build plan must be verified during planning; if the free plan restricts egress, recording ships together with the self-host/Ship move."* Additionally, LiveKit Cloud Egress servers run in LiveKit's cloud and cannot write directly to a developer's `127.0.0.1:9000` MinIO without either a public tunnel/S3 endpoint or running a local `livekit/egress` container in Docker Compose.
   - *Action needed:* Confirm during `M0-1` whether local dev should use a local `livekit-server` + `livekit-egress` compose profile for recording tests while staging/prod uses LiveKit Cloud + reachable MinIO/S3 URL.
2. **DM Channel Implementation Status in Repo:**
   - *Context:* Spec §9.3 and §13 lock DM calls as drop-in using a room per conversation. In `wichat-repo`, the `channels` table supports `type = 'dm'`, and the sidebar has a `"Chats"` section (`ProductionChannelSidebar` lines 642–662), but DM channel creation/listing endpoints are currently placeholder (`empty_chats`) in `ms-user`.
   - *Action needed:* `ms-core`'s token/webhook flow will support `type = 'dm'` channels from day one; confirm whether DM channel creation UI lands in a separate Phase 2chat slice or if a minimal DM channel creation helper is needed for testing DM calls.
3. **`text_and_voice` Channel Creation in UI vs Unified Default:**
   - *Context:* Spec §1 states channels have kinds `broadcast`, `text_only`, `voice_only`, `text_and_voice`, `forum`, `dm`. Migration `00020` originally migrated public/private channels to `text_and_voice`, but Migration `00042` later converted `text_and_voice` rows to `text_only` and split default categories into `"Text Channels"` (`text_only`) and `"Voice Channels"` (`voice_only`).
   - *Action needed:* `M0-2` and `M1-5` restore `text_and_voice` in the DB check constraint, backend validator, and frontend `CreateChannelDialog` so Spec §1 and §11 (`text_and_voice` channel calls) work as specified. Confirm whether existing `text_only` channels should remain `text_only` unless created/updated as `text_and_voice`, or if `"Text & Voice"` should be the default kind in `CreateChannelDialog`.
4. **Recording Playback / Download Access Scope:**
   - *Context:* Spec §5.8 specifies that starting/stopping a recording requires `record_call` (default: Admin) and files are linked to `call_sessions` in MinIO, but does not explicitly state whether **all channel members** with `view_channels`/`read_message_history` can view/download past recordings or only users with `record_call`.
   - *Action needed:* Confirm whether `GET /api/v1/channels/{id}/voice/recordings` should be accessible to all channel members (`read_message_history`) or restricted to `record_call` / Admins.
5. **React Native Mobile Client Scope (`apps/mobile`):**
   - *Context:* Spec §3.1 mentions Next.js web/PWA, Electron, and React Native, whereas Spec §11 Acceptance Criteria explicitly test **web (Chrome) and Electron**, and `frontend/apps/` in the repo currently contains only `frontend/apps/web/` (no Electron or React Native app directory exists yet).
   - *Action needed:* Confirm that Phase 3 frontend deliverables target `frontend/apps/web` (which also powers the PWA and Electron wrapper per System Design §12) and `@wichat/contracts`.
