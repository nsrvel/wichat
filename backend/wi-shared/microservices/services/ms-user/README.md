# ms-user (outbound client)

Typed **gRPC client** for the [ms-user](../../../../ms-user/README.md) process. This folder is wi-shared wiring only — domain logic lives in `backend/ms-user/internal/`.

## Layout

| Path | Role |
|------|------|
| `ms_user.go` | `Service`, `GRPC` struct, `NewGRPC(conn, settings, log)` |
| `grpc/auth/client.go` | `auth.Client` — AuthService RPCs wrapped with `utils.CallUnaryResult` |
| `grpc/auth/test/` | Bufconn retry test for idempotent reads |

## Call surface

```go
ms.User.GRPC.Auth.Login(ctx, &authv1.LoginRequest{...})
```

Interface for tests: `auth.Client` (implement all RPC methods; no `Close`).

## Retry policy

| RPC | Policy |
|-----|--------|
| `SetupStatus`, `Me` | `RetryIdempotent` |
| `Setup`, `Login`, `Refresh`, `Logout`, `GoogleStart`, `GoogleFinish` | `RetryNone` |

## Add another gRPC domain

1. Create `grpc/<domain>/client.go` with a small `Client` interface + `New(conn, settings, log)`.
2. Add a field on `GRPC` in `ms_user.go` and wire it in `NewGRPC`.
3. Connection stays shared: one `*grpc.ClientConn` per ms-user target in `microservices.Clients.Open`.

HTTP outbound (if ever needed): add `HTTP` on `Service` alongside `GRPC` — not implemented in v1.
