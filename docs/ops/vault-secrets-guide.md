# WiChat — HashiCorp Vault Secrets Guide

Panduan ini menjelaskan struktur secret di **HashiCorp Vault** (KV v2) untuk seluruh service WiChat.

- **Secret Engine (Mount):** `wistudio`
- **Folder Prefix:** `wichat/`
- **Daftar Path Secret:**
  - `wistudio/wichat/wichat-api-gateway`
  - `wistudio/wichat/wichat-ms-user`
  - `wistudio/wichat/wichat-ms-chat`
  - `wistudio/wichat/wichat-ms-notification`
  - `wistudio/wichat/wichat-frontend-web`

> [!IMPORTANT]
> - `INTERNAL_GRPC_TOKEN` **harus sama persis** di `wichat-api-gateway`, `wichat-ms-user`, `wichat-ms-chat`, dan `wichat-ms-notification`.
> - `INVITE_ONLY` **harus sama** di `wichat-api-gateway` dan `wichat-ms-user`.
> - `MFA_ENCRYPTION_KEY` **harus berbeda** dari `AUTH_JWT_SECRET` (keduanya minimal 32 karakter).
> - Untuk container/Kubernetes/Docker, set `GRPC_BIND=0.0.0.0` pada `ms-user`, `ms-chat`, dan `ms-notification`.

---

## 1. Enable KV v2 Engine (Jika Belum Ada)

```bash
vault secrets enable -path=wistudio kv-v2
```

---

## 2. `wistudio/wichat/wichat-api-gateway`

```bash
vault kv put wistudio/wichat/wichat-api-gateway \
  ENV="production" \
  LOG_LEVEL="info" \
  PORT="3000" \
  MS_USER_HOST="wichat-ms-user:3001" \
  MS_CHAT_HOST="wichat-ms-chat:3002" \
  MS_NOTIFICATION_HOST="wichat-ms-notification:3005" \
  WEB_ORIGIN="https://chat.example.com" \
  INTERNAL_GRPC_TOKEN="<generate-32plus-char-secret>" \
  KAFKA_BROKERS="kafka:9092" \
  KAFKA_CHAT_EVENTS_TOPIC="chat-events" \
  KAFKA_REALTIME_CONSUMER_GROUP="api-gateway-realtime" \
  INSTANCE_ID="" \
  MS_ATTEMPT_TIMEOUT_MS="5000" \
  MS_OVERALL_TIMEOUT_MS="15000" \
  MS_MAX_RETRIES="3" \
  MS_RETRY_BACKOFF_MS="100" \
  MS_BREAKER_FAILURE_THRESHOLD="5" \
  MS_BREAKER_OPEN_MS="30000" \
  INVITE_ONLY="false" \
  TURNSTILE_ENABLED="true" \
  TURNSTILE_SITE_KEY="<cloudflare-turnstile-site-key>" \
  TURNSTILE_SECRET_KEY="<cloudflare-turnstile-secret-key>" \
  TRUSTED_PROXIES="127.0.0.1,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16"
```

---

## 3. `wistudio/wichat/wichat-ms-user`

```bash
vault kv put wistudio/wichat/wichat-ms-user \
  ENV="production" \
  LOG_LEVEL="info" \
  PORT="3001" \
  GRPC_BIND="0.0.0.0" \
  INTERNAL_GRPC_TOKEN="<same-as-api-gateway>" \
  DATABASE_URL="postgres://wichat:<db-password>@postgres:5432/wichat?sslmode=disable" \
  REDIS_URL="redis://redis:6379/0" \
  MS_NOTIFICATION_HOST="wichat-ms-notification:3005" \
  WEB_ORIGIN="https://chat.example.com" \
  INVITE_ONLY="false" \
  AUTH_JWT_SECRET="<openssl-rand-base64-32>" \
  MFA_ENABLED="true" \
  MFA_ENCRYPTION_KEY="<distinct-32plus-char-secret>" \
  WEBAUTHN_RP_ID="chat.example.com" \
  WEBAUTHN_RP_ORIGINS="https://chat.example.com" \
  WEBAUTHN_RP_DISPLAY_NAME="WiChat" \
  GOOGLE_CLIENT_ID="<google-oauth-client-id>" \
  GOOGLE_CLIENT_SECRET="<google-oauth-client-secret>" \
  GOOGLE_REDIRECT_URL="https://chat.example.com/api/auth/callback/google" \
  MINIO_ENDPOINT="minio:9000" \
  MINIO_ACCESS_KEY="<minio-access-key>" \
  MINIO_SECRET_KEY="<minio-secret-key>" \
  MINIO_BUCKET="wichat-public" \
  MINIO_PUBLIC_BASE_URL="https://cdn.example.com/wichat-public" \
  MINIO_USE_SSL="false" \
  MINIO_REGION=""
```

---

## 4. `wistudio/wichat/wichat-ms-chat`

```bash
vault kv put wistudio/wichat/wichat-ms-chat \
  ENV="production" \
  LOG_LEVEL="info" \
  PORT="3002" \
  GRPC_BIND="0.0.0.0" \
  INTERNAL_GRPC_TOKEN="<same-as-api-gateway>" \
  DATABASE_URL="postgres://wichat:<db-password>@postgres:5432/wichat?sslmode=disable" \
  MS_USER_HOST="wichat-ms-user:3001" \
  KAFKA_BROKERS="kafka:9092" \
  KAFKA_CHAT_EVENTS_TOPIC="chat-events" \
  KAFKA_CONSUMER_GROUP="ms-chat-search-indexer" \
  ELASTICSEARCH_URL="http://elasticsearch:9200" \
  ELASTICSEARCH_MESSAGES_INDEX="wichat-messages" \
  MINIO_ENDPOINT="minio:9000" \
  MINIO_ACCESS_KEY="<minio-access-key>" \
  MINIO_SECRET_KEY="<minio-secret-key>" \
  MINIO_BUCKET="wichat-public" \
  MINIO_PUBLIC_BASE_URL="https://cdn.example.com/wichat-public" \
  MINIO_USE_SSL="false" \
  MINIO_REGION=""
```

---

## 5. `wistudio/wichat/wichat-ms-notification`

```bash
vault kv put wistudio/wichat/wichat-ms-notification \
  ENV="production" \
  LOG_LEVEL="info" \
  PORT="3005" \
  GRPC_BIND="0.0.0.0" \
  INTERNAL_GRPC_TOKEN="<same-as-api-gateway>" \
  SMTP_HOST="smtp-relay.brevo.com" \
  SMTP_PORT="587" \
  SMTP_SECURE="false" \
  SMTP_USER="<smtp-user>" \
  SMTP_PASSWORD="<smtp-password>" \
  SMTP_FROM="WiChat <noreply@example.com>"
```

---

## 6. `wistudio/wichat/wichat-frontend-web`

> [!NOTE]
> Variabel `NEXT_PUBLIC_*` dibaca saat build/runtime Next.js dan bersifat publik ke browser. Jangan pernah menaruh secret di `NEXT_PUBLIC_*`. `KLIPY_APP_KEY` dan `GIPHY_API_KEY` hanya dipakai di server Next.js.

```bash
vault kv put wistudio/wichat/wichat-frontend-web \
  NEXT_PUBLIC_API_URL="https://chat.example.com" \
  API_GATEWAY_PROXY_TARGET="http://wichat-api-gateway:3000" \
  NEXT_PUBLIC_REALTIME_WS_URL="wss://chat.example.com/api/v1/realtime" \
  NEXT_PUBLIC_OBJECT_PUBLIC_BASE_URL="https://cdn.example.com/wichat-public" \
  KLIPY_APP_KEY="<optional-klipy-app-key>" \
  GIPHY_API_KEY="<optional-giphy-api-key>"
```

---

## 7. Verifikasi Pembacaan Secret

```bash
vault kv get wistudio/wichat/wichat-api-gateway
vault kv get wistudio/wichat/wichat-ms-user
vault kv get wistudio/wichat/wichat-ms-chat
vault kv get wistudio/wichat/wichat-ms-notification
vault kv get wistudio/wichat/wichat-frontend-web
```
