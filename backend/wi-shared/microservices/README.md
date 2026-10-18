# microservices (wi-shared)

Outbound **dependency injection** for callers (api-gateway today; other `ms-*` later). One `Open` per process wires dial targets, timeouts, and typed clients; handlers call `ms.User.GRPC.Auth.*` instead of raw gRPC stubs.

Canonical architecture: [docs/architecture/microservices.md](../../../docs/architecture/microservices.md).

## Layout

| Path | Role |
|------|------|
| `clients.go` | `Config`, `Open`, `Close`, `Clients` |
| `services/<ms-name>/` | Typed facade per downstream service — see [services/ms-user/README.md](services/ms-user/README.md) |
| `utils/` | Shared `CallUnary` / `CallUnaryResult`, `Settings`, dial/server helpers — see [utils/README.md](utils/README.md) |

## Usage (api-gateway)

```go
ms, err := microservices.Open(microservices.Config{
    Log:      log,
    Settings: cfg.Microservices, // utils.Settings from env (ms_* keys)
    UserAddr: cfg.UserHost,
})
defer ms.Close()

out, err := ms.User.GRPC.Auth.Login(ctx, &authv1.LoginRequest{...})
```

Inject `*microservices.Clients` into HTTP delivery; map gRPC errors with `wi-shared/response/http` and `exception`.

## Adding a downstream service

1. Add `services/ms-<name>/` (`ms_<name>.go` + `grpc/<domain>/client.go`).
2. Extend `Clients` and `Open` in `clients.go` (new dial target in `Config`, `Close` for extra conns).
3. Document in `services/ms-<name>/README.md`.
