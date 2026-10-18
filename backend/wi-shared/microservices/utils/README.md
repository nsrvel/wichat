# microservices/utils

Shared primitives for **outbound** service calls: deadlines, retries, circuit breaker, and gRPC transport helpers. Service-specific clients under `../services/` import this package; they do **not** import the root `microservices` package (avoids import cycles).

Full env defaults and retry rules: [docs/architecture/microservices.md](../../../docs/architecture/microservices.md).

## Layout

| Path | Role |
|------|------|
| `config.go` | `Settings`, `CallConfig`, `RetryPolicy`, `DefaultSettings` |
| `call.go` | `CallUnary` — per-attempt + overall timeout, retries |
| `retry.go`, `breaker.go`, `outbound.go` | Retry classification, breaker, `IsOutboundUnavailable` (gateway 503) |
| `correlation/` | `RequestIDHeader`, `WithRequestID`, `RequestIDFromContext` |
| `grpc/` | `Dial`, `NewServer`, client/server request-ID interceptors |
| `test/` | Breaker and `CallUnary` unit tests |

## Settings (env)

| Key | Default |
|-----|---------|
| `ms_attempt_timeout_ms` | 5000 |
| `ms_overall_timeout_ms` | 15000 |
| `ms_max_retries` | 3 |
| `ms_retry_backoff_ms` | 100 |
| `ms_breaker_failure_threshold` | 5 (0 disables) |
| `ms_breaker_open_ms` | 30000 |

Squash into caller config (e.g. api-gateway `Config.Microservices`).

## Typical client method

```go
cfg := utils.CallConfigFromSettings(settings, utils.RetryIdempotent)
cfg.Breaker = utils.NewBreaker(settings)
err := utils.CallUnary(ctx, cfg, log, "service.Method/Name", func(callCtx context.Context) error {
    return stub.DoSomething(callCtx, req)
})
```

Inbound gRPC servers (e.g. ms-user) use `utils/grpc.NewServer(log)` for recovery and request ID on the server side.
