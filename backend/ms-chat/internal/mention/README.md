# Mention inbox (`ms-chat`)

`message_mentions` is owned by **ms-chat** (write path: `message` service + `mention/repository`). SQL migrations live under **`ms-user`** goose because WiChat uses a single Postgres database and applies schema from `backend/ms-user/internal/migrations/` (see `00056_message_mentions.sql`, `00057_message_mentions_read_at.sql`).

- **Inbox use cases:** `internal/mention/service` — list, mark read, unread count (ACL-filtered).
- **Read state:** `read_at` on each `(message_id, mentioned_user_id)` row; `MarkGroupMentionsRead`, `CountGroupMentionsUnread`.