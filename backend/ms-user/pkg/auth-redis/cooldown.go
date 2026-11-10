package authRedis

import "time"

// CooldownSecs rounds a wait duration up to whole seconds for rate-limit metadata.
func CooldownSecs(d time.Duration) int {
	if d <= 0 {
		return 0
	}
	secs := int((d + time.Second - 1) / time.Second)
	if secs < 1 {
		return 1
	}

	return secs
}

// CooldownInt32 returns CooldownSecs capped for protobuf int32 fields.
func CooldownInt32(d time.Duration) int32 {
	secs := CooldownSecs(d)
	if secs > 1_000_000 {
		return 1_000_000
	}

	return int32(secs) //nolint:gosec // G115 — bounded above
}
