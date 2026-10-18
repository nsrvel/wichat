# wi-shared

Shared libraries for WiChat backend services.

## Module READMEs

| Module | Description |
|--------|-------------|
| [config/README.md](config/README.md) | Viper load: defaults, `.env`, environment |
| [exception/README.md](exception/README.md) | Semantic errors, i18n codes, HTTP/gRPC kinds |
| [response/README.md](response/README.md) | JSON envelopes; `http` and `grpc` writers |
| [infra/logger/README.md](infra/logger/README.md) | `slog` logger (`Logger`, `New`) |
| [models/README.md](models/README.md) | Protobuf contracts, `make proto` |
| [microservices/README.md](microservices/README.md) | Outbound `Open`, `User.GRPC.*`, `utils` |

## Architecture docs

- Errors and API envelopes: [docs/architecture/error-handling.md](../../docs/architecture/error-handling.md)
- Logging: [docs/architecture/logging.md](../../docs/architecture/logging.md)
- Service-to-service clients: [docs/architecture/microservices.md](../../docs/architecture/microservices.md)

Service layout in `ms-*` / gateway: `.cursor/rules/wichat-go-layers.mdc`.

From `backend/wi-shared/`: `make check`, `make proto` (requires [buf](https://buf.build/docs/installation)).
