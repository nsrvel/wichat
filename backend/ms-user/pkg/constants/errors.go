package constants

import "errors"

// Auth persistence sentinel errors (Postgres + Redis). Mapped to exception in service.
var (
	ErrNotFound         = errors.New("auth: not found")
	ErrConflict         = errors.New("auth: conflict")
	ErrAlreadyUsed      = errors.New("auth: refresh already used")
	ErrResetRateLimited = errors.New("auth: reset rate limited")
)
