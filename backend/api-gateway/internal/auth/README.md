# auth (api-gateway)

Public **HTTP** auth routes only — no local repository or service. Use cases run in **ms-user**; delivery calls `ms.User.GRPC.Auth` for credentials and sessions. Other domains (`user`, `group`, `group-member`, `group-role`, `channel`, `invite`) each own their own `internal/<domain>/delivery/http/`.

## Delivery

[`delivery/http/`](delivery/http/): `handler.go` (orchestrator), `routes.go` (`Register`). Tests: `delivery/http/test/`.

Helpers: [`pkg/cookie`](../../pkg/cookie), [`pkg/auth-http`](../../pkg/auth-http).

See `wichat-go-layers.mdc` (HTTP delivery) and [`wi-shared/microservices`](../../wi-shared/microservices/README.md).
