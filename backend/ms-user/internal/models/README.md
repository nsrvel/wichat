# models (ms-user)

Shared **`package models`** for this service — not per-domain `internal/<domain>/model`.

| File pattern | Contents |
|--------------|----------|
| `<domain>_model.go` | Domain types, inputs, results (no ORM imports) |
| `<domain>_repository_model.go` | Bun row structs (`bun:"table:auth.users"`, …), `*RowFrom` / `ToModel` helpers |

Examples today: `auth_model.go`, `auth_repository_model.go`. Future domains add `user_model.go`, `user_repository_model.go`, etc.

Imported by `internal/<domain>/service`, `repository`, and `delivery/grpc`. Repository code uses `models.UserRow`; services use `models.User`.
