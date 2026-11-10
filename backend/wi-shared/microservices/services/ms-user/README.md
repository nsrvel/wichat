# ms-user (outbound client)

Typed **gRPC client** for the [ms-user](../../../../ms-user/README.md) process. This folder is wi-shared wiring only — domain logic lives in `backend/ms-user/internal/`.

Proto v1 still exposes one `AuthService`; outbound clients are split by **domain** to match `internal/auth`, `internal/user`, `internal/group`, and `internal/channel`. All domain clients share one `AuthServiceClient` stub on a single connection.

## Layout

| Path | Role |
|------|------|
| `ms_user.go` | `Service`, `GRPC` struct, `NewGRPC(conn, settings, log)` |
| `grpc/auth/client.go` | Credentials, sessions, OAuth, password reset |
| `grpc/user/client.go` | Instance policy, setup, guest registration |
| `grpc/group/client.go` | Group membership and icon upload |
| `grpc/channel/client.go` | Channel list and access assertion |
| `grpc/general/client.go` | `GeneralService/Health` (liveness), `RetryIdempotent` |
| `grpc/user/test/` | Client retry tests |

## Call surface

```go
ms.User.GRPC.Auth.Login(ctx, &userv1.LoginRequest{...})
ms.User.GRPC.User.SetupStatus(ctx, &userv1.SetupStatusRequest{})
ms.User.GRPC.Group.CreateGroup(ctx, &userv1.CreateGroupRequest{...})
ms.User.GRPC.Channel.ListGroupChannels(ctx, &userv1.ListGroupChannelsRequest{...})
ms.User.GRPC.General.Health(ctx, &userv1.HealthRequest{})
```

Interface for tests: implement each domain `Client` (no `Close` on the facade).

## Retry policy

| RPC | Policy |
|-----|--------|
| `User.SetupStatus`, `Auth.Me`, `Group.List*`, `Channel.*`, `General.Health` | `RetryIdempotent` |
| Writes (login, register, create group, join, …) | `RetryNone` |

## Add another gRPC domain

1. Create `grpc/<domain>/client.go` with a small `Client` interface + `New(stub, settings, log)`.
2. Add a field on `GRPC` in `ms_user.go` and wire it in `NewGRPC` (reuse `authStub`).

HTTP outbound (if ever needed): add `HTTP` on `Service` alongside `GRPC` — not implemented in v1.
