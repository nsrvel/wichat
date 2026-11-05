package minio

import "errors"

// ErrDisabled is returned when object storage is not configured (empty endpoint).
var ErrDisabled = errors.New("object storage disabled")
