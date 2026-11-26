import type { InstancePolicy } from "@wichat/contracts/auth";

/** Whether MFA UI and client assumptions should be shown (BE still enforces via setup-status). */
export function isMfaFeatureEnabled(policy: InstancePolicy | undefined): boolean {
	if (process.env.NEXT_PUBLIC_MFA_ENABLED === "false") {
		return false;
	}
	return policy?.mfa_enabled !== false;
}