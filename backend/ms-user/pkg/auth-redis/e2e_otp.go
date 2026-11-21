package authRedis

import "strings"

// DevE2EOTPKey holds a plaintext OTP for Playwright/e2e in development only (never written in production).
func DevE2EOTPKey(challengeID string) string {
	return KeyPrefix + "dev:e2e-otp:" + strings.TrimSpace(challengeID)
}
