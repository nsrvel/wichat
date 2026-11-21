# ms-chat

Channel messages, outbox relay, and Elasticsearch indexing for WiChat.

## Domains

| Domain | Path | Delivery |
|--------|------|----------|
| `message` | `internal/message/` | gRPC (`MessageService`) |
| `outbox` | `internal/outbox/` | background relay → Kafka |
| `search` | `internal/search/` | Kafka consumer indexer (query API deferred) |
| `general` | `internal/general/` | gRPC health only |

## Shared database migrations

Goose runs from **ms-user** (`internal/migrations/`), but several migrations are **owned by ms-chat** data (comments in SQL):

| Migration | Schema |
|-----------|--------|
| `00021_messages.sql` | `messages`, reactions, attachments |
| `00049_message_channel_seq_outbox.sql` | `messages.channel_seq`, `channels.message_seq`, `event_outbox` |

Apply with:

```bash
cd backend && make migrate-up
```

Both **ms-user** and **ms-chat** use the same `DATABASE_URL` (single Postgres instance).

## Local checks

```bash
cd backend/ms-chat && make check
```