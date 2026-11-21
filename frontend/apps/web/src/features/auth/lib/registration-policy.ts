import type { InstancePolicy } from "@wichat/contracts/auth";

export function inviteTokenFromReturnPath(returnTo: string): string | undefined {
	const trimmed = returnTo.trim();
	const match = /^\/invite\/([^/?#]+)/.exec(trimmed);
	if (!match?.[1]) {
		return undefined;
	}
	try {
		return decodeURIComponent(match[1]).trim() || undefined;
	} catch {
		return match[1].trim() || undefined;
	}
}

/** When invite-only enforcement is on, public sign-up needs a valid invite token in the flow. */
export function isInviteOnlyRegistrationBlocked(
	policy: InstancePolicy | undefined,
	inviteToken: string | undefined,
): boolean {
	if (!policy?.invite_only) {
		return false;
	}
	return !inviteToken?.trim();
}
