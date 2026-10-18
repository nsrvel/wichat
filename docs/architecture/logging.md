# Application logging

WiChat services use **`wi-shared/infra/logger`** (`log/slog`) for structured application logs. Module notes: [infra/logger/README.md](../../backend/wi-shared/infra/logger/README.md).

Container platforms and **Promtail/Alloy → Loki → Grafana** collect stdout; the app does not push logs directly in v1. `logger.New`, levels, and injection: [infra/logger/README.md](../../backend/wi-shared/infra/logger/README.md).

## HTTP errors and logging

Gateway handlers should pass:

```go
responsehttp.WriteError(w, r, err, responsehttp.WithLogger(logger.HTTPWriteErrorLogger(appLog)))
```

| Situation | Logged? |
|-----------|---------|
| Raw error → 500 | Yes |
| `exception.Internal(cause)` → 500 | Yes (`cause`) |
| Expected 4xx (`NotFound`, `InvalidInput`, …) | No |
| gRPC downstream internal | Yes (`WriteErrorFromGRPC`) |

Client responses stay i18n codes only — see [error-handling.md](error-handling.md).

## Do not log

- Passwords, tokens, refresh cookies, API secrets
- Full message bodies or PII unless explicitly required for audit (audit uses Postgres, not app logs)

## Gateway request ID

`api-gateway/internal/middleware` sets `correlation.RequestIDHeader` (`X-Request-ID`) on every response and stores the value on `r.Context()` via `correlation.WithRequestID` (`microservices/utils/correlation`).

Outbound gRPC from api-gateway uses `wi-shared/microservices` client interceptors to send `x-request-id` metadata; ms-user server interceptors read it back into context (recovery logs include `request_id` when present).

## Follow-up

- Add `request_id` to slog lines from context (automatic on every log call)

See [microservices.md](microservices.md) for CallUnary and retries.
