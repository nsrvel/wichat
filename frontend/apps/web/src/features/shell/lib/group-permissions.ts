import type { Group } from "@wichat/contracts";

/** ponytail: extend when group RBAC lands — today only owner vs member. */
export function canManageGroup(role: Group["role"] | undefined): boolean {
	return role === "owner";
}

export function isGroupMember(role: Group["role"] | undefined): boolean {
	return role === "member";
}
