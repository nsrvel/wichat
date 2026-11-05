import { ApiError } from "@wichat/api-client/api-error";

/** Pick API cooldown when parent state is still unset (0). */
export function effectiveResendCooldownSec(parentSec: number, fallbackSec: number): number {
	if (parentSec > 0) {
		return parentSec;
	}
	return fallbackSec > 0 ? fallbackSec : 60;
}

/** Sync UI countdown after auth.reset_rate_limited. */
export function resendRateLimitCooldownSec(error: unknown): number | null {
	if (!(error instanceof ApiError) || error.code !== "auth.reset_rate_limited") {
		return null;
	}
	const raw = error.params?.retry_after_seconds;
	if (typeof raw === "number" && raw > 0) {
		return raw;
	}
	return null;
}
