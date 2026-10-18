# api-gateway

Public HTTP API (REST `/api/v1/`, auth today). Outbound: `microservices.Open` → `ms.User.GRPC.*`.

```bash
cp .env.example .env   # optional overrides
make dev    # Air live reload (start ms-user if gateway calls it)
make run    # single run
make check  # fmt, lint, test
```

From `backend/`: `make dev-api-gateway`.

## Layout

| Path | Role |
|------|------|
| `cmd/` | `main` |
| `internal/config/` | Service config (`config.Load`, `utils.Settings` squash) |
| `internal/server/` | HTTP listen, `microservices.Open`, route wiring |
| `internal/middleware/` | Request ID, CORS, rate limit, recover |
| `internal/auth/delivery/http/` | `handler.go`, `routes.go` — see [internal/auth/README.md](internal/auth/README.md) |
| `pkg/` | Gateway-only helpers — see [pkg/README.md](pkg/README.md) |

Default listen: **`PORT=3000`**.
