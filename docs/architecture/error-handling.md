# Error handling (backend)

WiChat services use a shared error model in `backend/wi-shared`: semantic errors at the use-case layer, plain `error` below that, and **response writers** at HTTP/gRPC controller boundaries. Package notes: [exception/README.md](../../backend/wi-shared/exception/README.md), [response/README.md](../../backend/wi-shared/response/README.md).

## Layers

| Layer | Returns | Writes to client |
|-------|---------|------------------|
| Repository / infra | `error` (driver, I/O) | — |
| Service / use case | `error` wrapping `*exception.Error` | — |
| Microservice gRPC handler | — | `response/grpc.ToStatus(err)` |
| Gateway HTTP handler | — | `WriteError` (domain), `WriteOutboundError` (downstream gRPC/transport), or `WriteErrorFromGRPC` |

Clients resolve messages with i18n: `t(error.code, error.params)`.

## `exception` / `response` packages

API and constructors: [exception/README.md](../../backend/wi-shared/exception/README.md), writers: [response/README.md](../../backend/wi-shared/response/README.md). Each new `MessageCode` needs a frontend locale key. Errors without `*exception.Error` become **internal** at the writer.

## HTTP JSON shape

```json
{
  "error": {
    "code": "user.not_found",
    "params": { "user_id": "…" }
  }
}
```

- **5xx**: always `code: "error.internal"`, `params` omitted or null.
- **4xx**: `params` optional (validation, interpolation).

## gRPC wire convention

- `status.Code()` from `Kind` mapping.
- `status.Message()` = string `MessageCode` (i18n key), not translated text.
- Optional `google.rpc.ErrorInfo` in details:
  - `Reason` = same code
  - `Metadata["params"]` = JSON object for interpolation

Gateway uses `response/http.WriteOutboundError` for outbound calls (gRPC status → HTTP, transport/unavailable → **503** `error.unavailable`) and `WriteErrorFromGRPC` when you already have a gRPC status error.

## Kind policy

| Kind | HTTP | gRPC |
|------|------|------|
| InvalidInput | 400 | InvalidArgument |
| Unauthenticated | 401 | Unauthenticated |
| PermissionDenied | 403 | PermissionDenied |
| NotFound | 404 | NotFound |
| Conflict | 409 | AlreadyExists |
| StateConflict | 409 | Aborted |
| RateLimited | 429 | ResourceExhausted |
| Unavailable | 503 | Unavailable |
| DeadlineExceeded | 504 | DeadlineExceeded |
| Internal | 500 | Internal |

Gateway `ParseStatus` rejects gRPC status messages that are not valid `MessageCode` strings (maps to `error.internal`).

## Security

- Never expose `Cause`, driver messages, or stack traces in HTTP/gRPC client payloads.
- Do not put secrets or passwords in `Params`.

## Adding a new error code

1. Add `MessageCode` const in `wi-shared/exception/code.go` (or `code_<domain>.go` when the catalog grows).
2. Add matching keys in frontend i18n.
3. Return the code from the service via the appropriate constructor.
4. If a new `Kind` is introduced (rare), extend `exception/test/kind_test.go` and `kind.go` mappings.

## Observability

`response/http.WriteError` accepts optional `WithLogger` for logging server-side errors before returning `error.internal`:

- Raw (non-exception) errors → 500
- `exception.Internal(cause)` → logs `cause`, client still gets `error.internal`
- Expected 4xx exceptions are not logged by the writer

Use `logger.HTTPWriteErrorLogger(appLog)` from `wi-shared/infra/logger`. See [logging.md](logging.md).
