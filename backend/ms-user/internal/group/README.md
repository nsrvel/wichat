# group (ms-user)

Group membership use cases: list, create, join. Postgres I/O stays on [`internal/user/repository`](../user/repository/pg_repository.go).

## Service

[`service/service.go`](service/service.go) — `Service` interface + `NewService(users, sessions, secret, log)`. Caller identity via [`pkg/access`](../../pkg/access/user.go) (JWT + live session).

Tests: [`service/test/service_test.go`](service/test/service_test.go).

gRPC: [`internal/auth/delivery/grpc`](../auth/delivery/grpc/server.go) routes group RPCs on `AuthService`.

See `wichat-go-layers.mdc`.
