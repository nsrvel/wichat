# Service-to-service (`wi-shared/microservices`)

Outbound calls: one `microservices.Open` per process, inject `*microservices.Clients`, call typed facades (e.g. `ms.User.GRPC.Auth.Login`). Shared timeouts, retries, dial, and request ID live in **`microservices/utils`**.

**Module docs:** [microservices/README.md](../../backend/wi-shared/microservices/README.md), [utils/README.md](../../backend/wi-shared/microservices/utils/README.md), [services/ms-user/README.md](../../backend/wi-shared/microservices/services/ms-user/README.md).

Protobuf contracts: `wi-shared/models` (including `general.v1.GeneralService/Health` for **liveness**). Each gRPC `ms-*` implements `internal/general/delivery/grpc` like auth; callers use `ms.User.GRPC.General.Health`. Gateway downstream failures: `response/http.WriteOutboundError`.

## Retry policy (summary)

| Policy | Use |
|--------|-----|
| `RetryNone` | Writes without idempotency key |
| `RetryIdempotent` | Safe reads |
| `RetrySafe` | Reserved for idempotency-key writes |

No retry on client gRPC codes (`InvalidArgument`, `NotFound`, …). Retries `Unavailable`, `ResourceExhausted`, and transport-style failures. Env defaults and breaker: `utils` README.
