# models

gRPC/protobuf API contracts shared by WiChat services.

- **Sources:** `models/<domain>/v1/*.proto` (add domains as needed, e.g. `auth`, `user`).
- **Generated:** `gen/` — run `make proto` from `wi-shared/` (requires [buf](https://buf.build/docs/installation)). Do not hand-edit `gen/`.

When services live in separate repos, depend on a tagged **wi-shared** module release that includes `models/gen`.
