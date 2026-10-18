# response (wi-shared)

JSON **envelopes** and boundary writers: map `exception` (and raw/downstream errors) to HTTP or gRPC for clients.

Conventions: [docs/architecture/error-handling.md](../../../docs/architecture/error-handling.md).

## Layout

| Path | Role |
|------|------|
| `response.go` | `DataEnvelope`, `ErrorEnvelope`, `NewDataEnvelope`, `NewErrorEnvelope` |
| `http/writer.go` | Gateway HTTP: `WriteJSON`, `WriteError`, `WriteOutboundError`, `WriteErrorFromGRPC` |
| `grpc/status.go` | ms-* gRPC: `ToStatus`, `ParseStatus`, `FromError` |

## HTTP (api-gateway)

```go
responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(body))
responsehttp.WriteError(w, r, err) // domain *exception.Error or safe fallback
responsehttp.WriteOutboundError(w, r, grpcErr) // downstream ms unavailable / transport
```

Use `WriteErrorFromGRPC` when propagating a gRPC status from another hop without re-wrapping in service.

## gRPC (ms-*)

```go
return nil, responsegrpc.ToStatus(err)
```

`ToStatus` encodes `exception.Error` into gRPC status details; unknown errors become internal.

## Tests

`http/test/`, `grpc/test/` — envelope shape, no leak of causes, outbound 503 mapping.
