# group (ms-user)

Group membership use cases: list, create, join. Postgres I/O in [`pg_repository.go`](repository/pg_repository.go); group icons in [`minio_repository.go`](repository/minio_repository.go) on [`wi-shared/infra/minio`](../../../wi-shared/infra/minio/README.md) (disabled when MinIO endpoint is unset). Default channels are **not** inserted here — see [`internal/channel`](../channel/README.md).

## Service

[`service/service.go`](service/service.go) — `NewService(groups, channels, users, objects, sessions, secret, log)`. Caller identity via [`pkg/access`](../../pkg/access/user.go) (JWT + live session).

Tests: [`service/test/service_test.go`](service/test/service_test.go). Repository integration: [`repository/test/pg_repository_test.go`](repository/test/pg_repository_test.go).

gRPC: [`internal/auth/delivery/grpc`](../auth/delivery/grpc/server.go) routes group RPCs on `AuthService`.

See `wichat-go-layers.mdc`.
