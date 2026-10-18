# exception (wi-shared)

Semantic errors for **service / use-case** layers: stable i18n `MessageCode`, `Kind` for HTTP and gRPC mapping, optional `Cause` for logs only (never sent to clients).

Full conventions: [docs/architecture/error-handling.md](../../../docs/architecture/error-handling.md).

## Layout

| File | Role |
|------|------|
| `kind.go` | `Kind`, `HTTPStatus`, `GRPCCode`, `KindFromGRPCCode` |
| `code.go` | `MessageCode` constants (`Msg*`), `ValidMessageCode` |
| `error.go` | Constructors, `As`, `Error` type |

## Usage (service)

```go
return exception.NotFound(exception.MsgUserNotFound, map[string]any{"id": id})
return exception.Internal(repoErr)
```

Repository / infra may return plain `error`; service wraps or maps to `exception.*`.

## Usage (delivery)

- **HTTP (gateway):** `response/http.WriteError` — use `exception.As` inside the writer.
- **gRPC (ms-*):** `response/grpc.ToStatus(err)` on the way out.

```go
if ex, ok := exception.As(err); ok {
    // branch on ex.Kind, ex.Code
}
```

Add every new `MessageCode` to frontend locale files when exposing it to clients.

## Tests

`exception/test/` — constructors, `As` through `fmt.Errorf("%w")`, kind mapping.
