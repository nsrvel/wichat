# internal/general (ms-user)

Operational gRPC surfaces — not business domains (auth, user, …). Copy this layout to other gRPC `ms-*` processes.

## Layout

| Path | Role |
|------|------|
| `delivery/grpc/server.go` | `Server` implements `general.v1.GeneralServiceServer` (same pattern as auth) — `Health` RPC for liveness |

Wire: `userv1.RegisterGeneralServiceServer(grpcSrv, generalgrpc.NewServer("ms-user"))`. Proto: [`wi-shared/models/proto/user/v1/general.proto`](../../../wi-shared/models/proto/user/v1/general.proto). Callers: `ms.User.GRPC.General.Health`.
