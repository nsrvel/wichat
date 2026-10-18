# pkg (ms-user)

Helpers shared **inside this microservice** only (`internal/auth`, `internal/server`, …). Not imported by `api-gateway` or other `ms-*`.

Cross-microservice code belongs in [`wi-shared`](../../wi-shared/).

| Package | Role |
|---------|------|
| `password` | Argon2id hash / verify |
| `jwt` | Access token issue / parse |
| `token` | Refresh secrets, opaque IDs |
| `google` | OAuth `Config`, PKCE, code exchange |
| `auth-pb` (`package authPb`) | `models` ↔ auth v1 protobuf |
