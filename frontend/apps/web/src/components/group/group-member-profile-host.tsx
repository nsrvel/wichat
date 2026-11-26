"use client";

import {
	GroupMemberProfileSurface,
	type MemberProfileFallbackUser,
} from "@/components/group/group-member-profile-surface";

export type { MemberProfileFallbackUser };

type GroupMemberProfileHostProps = {
	groupId: string;
	memberId: string | null;
	fallbackUser?: MemberProfileFallbackUser | null;
	onClose: () => void;
};

/**
 * Modal member profile (click) — same `MemberProfileDialog` body as mention hover popover.
 */
export function GroupMemberProfileHost({
	groupId,
	memberId,
	fallbackUser = null,
	onClose,
}: GroupMemberProfileHostProps) {
	return (
		<GroupMemberProfileSurface
			groupId={groupId}
			memberId={memberId}
			fallbackUser={fallbackUser}
			onClose={onClose}
			presentation="dialog"
			enabled={Boolean(groupId && memberId)}
		/>
	);
}