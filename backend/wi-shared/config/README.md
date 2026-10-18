# config (wi-shared)

Shared **Viper** loader for service `Config` structs: merge defaults, optional `.env` file, then OS environment. Used by api-gateway, ms-user, and future `ms-*` modules.

## Layout

| File | Role |
|------|------|
| `config.go` | `Load(envFile, dst, defaults)` |
| `apply.go` | `ListenAddress`, `LogLevel` helpers for viper/slog wiring |

## Usage

Define a struct with `mapstructure` tags (and `squash` for nested blocks like `utils.Settings`). Pass a defaults value of the same shape:

```go
var cfg Config
if err := config.Load(".env", &cfg, Defaults); err != nil {
    return err
}
```

**Precedence (low → high):** defaults struct → env file (if path set and file exists) → environment variables.

Missing env file is not an error; empty `envFile` skips file load.

## Tests

`config/test/` — load precedence and `apply` helpers.
