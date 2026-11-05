# pkg (ms-user)

Helpers shared **inside this microservice** only. Not imported by `api-gateway` or other `ms-*`. Cross-service code → [`wi-shared`](../../wi-shared/).

**Layout:** one role folder per concern; files at that level (`oauth/google.go`, not nested leaf dirs). Kebab-case dir → camelCase `package` when multi-word. See `wichat-go-style.mdc`.

**Do not add micro-helpers** in `pkg/` (one-liner wrappers, single-call-site utils). Inline at the caller or keep unexported helpers in the owning `internal/*/service` file when the logic is domain-local.

```
pkg/
├── constants/          # sentinels, limits, domain enums, SQL literals (no const in repository/)
├── validator/          # input validation (register, guest, setup, …)
├── converter/          # authPb — models ↔ user v1 proto (`auth_pb.go`)
├── jwt/ token/ password/ otp/
├── auth-redis/         # authRedis — OTP HMAC, resend backoff
├── access/             # access token → user + session
├── oauth/              # google — `google.go` (Google OAuth helpers)
├── db/                 # pgErr — Postgres → constants (`map.go`)
├── username/           # handle allocation from email
└── test-support/       # testSupport — mocks + miniredis wiring (*_test only)
```

Import canonical paths only — **no** re-exports from `internal/`.
