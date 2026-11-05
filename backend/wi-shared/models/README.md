# models

gRPC/protobuf API contracts shared by WiChat services.

- **Sources:** `proto/<ms>/v1/*.proto` — one folder per microservice (e.g. `proto/user/v1/auth.proto`, `proto/user/v1/general.proto` for **ms-user**). All files in a version folder share `package <ms>.v1` and one Go import (`userv1` for `user`).
- **Generated:** `gen/` — run `make proto` from `wi-shared/` (requires [buf](https://buf.build/docs/installation)). Do not hand-edit `gen/`.

When services live in separate repos, depend on a tagged **wi-shared** module release that includes `models/gen`.
