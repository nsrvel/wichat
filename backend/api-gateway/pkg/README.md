# pkg (api-gateway)

Helpers shared **inside api-gateway** only (`internal/auth`, `internal/server`, …). Not imported by `ms-*` or `wi-shared`.

Cross-service code belongs in [`wi-shared`](../../wi-shared/).

| Package | Role |
|---------|------|
| `constants` | Gateway-only constants (e.g. `MaxUserAgentLen`) — not in delivery/middleware |
| `cookie` | HttpOnly refresh and OAuth PKCE cookies |
| `auth-http` (`authHttp`, `auth_http.go`) | JSON bodies, `WriteSession`, `WriteOutbound`, OAuth `RedirectSignIn`, `GRPCMessageCode` |
