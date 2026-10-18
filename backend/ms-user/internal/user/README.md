# user (ms-user)

User/workspace domain in the ms-user process: `repository/` → `service/` → `delivery/grpc/` when implemented.

Auth lives in sibling package `internal/auth/`. Domain types (when added): `internal/models/user_model.go` and `user_repository_model.go` — see [`internal/models/README.md`](../../models/README.md).

See `wichat-go-layers.mdc` and `docs/architecture/wichat-system-design.md`.
