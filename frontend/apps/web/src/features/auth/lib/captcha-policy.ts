import type { InstancePolicy } from "@wichat/contracts/auth";

/** Fail-closed until policy is loaded and any required Turnstile token is present. */
export function authSubmitBlocked(
	policy: InstancePolicy | undefined,
	policyReady: boolean,
	turnstileToken: string,
): boolean {
	if (!policyReady || !policy) {
		return true;
	}
	if (policy.turnstile_required && !policy.turnstile_site_key?.trim()) {
		return true;
	}
	const siteKey = policy.turnstile_site_key?.trim() ?? "";
	if (siteKey !== "" && turnstileToken.trim() === "") {
		return true;
	}
	return false;
}