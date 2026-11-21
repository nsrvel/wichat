# CSRF and session cookies (api-gateway)

## Model

- **Refresh tokens** are HttpOnly cookies (`SameSite=Lax`, `Secure` when not `ENV=development`).
- **Access tokens** are short-lived JWTs returned in JSON; the web client stores them in `sessionStorage` and sends `Authorization: Bearer` on API calls.
- **CORS** allows a single configured `WEB_ORIGIN` with `Access-Control-Allow-Credentials: true`.

## CSRF

Browser cookie-authenticated requests (`POST`/`PUT`/`PATCH`/`DELETE` with cookies) are protected by:

1. **`SameSite=Lax`** on refresh cookies — cross-site POSTs from arbitrary origins do not attach cookies.
2. **`CrossOriginStateChange` middleware** — if the request includes an `Origin` header, it must match `WEB_ORIGIN` (same rules as CORS). Wrong origin → `403` before handlers run.
3. **CORS** — only the trusted origin receives credentialed responses.

State-changing calls that use **only** `Authorization: Bearer` (no cookies) rely on token secrecy, not double-submit CSRF tokens.

## WebSocket

Realtime auth uses `Sec-WebSocket-Protocol: bearer.<access_token>` (preferred). Query `access_token` is legacy and avoided by the web client.

## Production checklist

- `WEB_ORIGIN` must be `https://…`
- `TURNSTILE_ENABLED=true` with real secrets
- `INTERNAL_GRPC_TOKEN` set consistently on all core services