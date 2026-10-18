# logger (wi-shared)

Structured application logging via **`log/slog`**. Inject `logger.Logger` into service, repository, and delivery (see `wichat-go-layers.mdc`).

Operations: [docs/architecture/logging.md](../../../docs/architecture/logging.md).

## Layout

| File | Role |
|------|------|
| `logger.go` | `Logger` interface, `New(Options)` |
| `config.go` | `ParseLevel`, level string helpers |

## Usage

```go
log := logger.New(logger.Options{
    Service: "api-gateway",
    Env:     cfg.Env,
    Level:   cfg.LogLevel,
})
log.Info("http listening", "addr", addr)
log.With("request_id", id).Error("downstream failed", "err", err)
```

| `Env` | Handler |
|-------|---------|
| `production` | JSON to stdout |
| Other | Text to stdout |

Default attributes on every line: `service`, `env`.

## Tests

`infra/logger/test/` — level parsing and handler selection.
