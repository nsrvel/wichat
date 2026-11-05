# minio (wi-shared)

S3-compatible object storage wrapper for WiChat services. One `Client` is bound to **one bucket**; use multiple clients for multiple buckets.

Inject `minio.Client` from service wiring; domain `minio_repository.go` adapters build keys and call `Upload` / `DeletePrefix`.

## Layout

| File | Role |
|------|------|
| `client.go` | `Client` interface, live MinIO impl, disabled no-op |
| `config.go` | `Config` (endpoint, keys, bucket, public base URL) |
| `errors.go` | `ErrDisabled` when endpoint is unset |
| `bootstrap.go` | `EnsurePublicReadBucket` (create bucket + public read policy) |

## Usage

```go
client, err := minio.New(minio.Config{
    Endpoint:      cfg.MinioEndpoint,
    AccessKey:     cfg.MinioAccessKey,
    SecretKey:     cfg.MinioSecretKey,
    Bucket:        cfg.MinioBucket,
    PublicBaseURL: cfg.MinioPublicBaseURL,
    UseSSL:        cfg.MinioUseSSL,
    Region:        cfg.MinioRegion,
}, log)
```

Empty `Endpoint` → disabled client (`ErrDisabled` on I/O). `PublicURL` still returns a path-style URL.

## API

| Method | Notes |
|--------|--------|
| `Get` | Caller must `Close` `Object.Body` |
| `Upload` / `UploadMany` | `UploadMany` is sequential (no atomic batch) |
| `Copy` | Same bucket |
| `Delete` / `DeletePrefix` | Prefix listing is recursive |
| `PublicURL` | Joins `PublicBaseURL` and key |

Repositories stay dumb: plain `error`, no `exception` import (`wichat-go-layers.mdc`).

## Tests

`infra/minio/test/` — disabled client, bucket validation, `PublicURL` (no live MinIO server).
