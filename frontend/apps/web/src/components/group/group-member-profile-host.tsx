"use client";

import { useQuery } from "@tanstack/react-query";
import { GroupMemberSchema, type GroupMember } from "@wichat/contracts/group/member";
import type { UserRelationshipKind } from "@wichat/contracts/user";
import { useCallback, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { MemberBlockDialog } from "@/features/group-settings/components/members/member-block-dialog";
import { MemberIgnoreDialog } from "@/features/group-settings/components/members/member-ignore-dialog";
import {
	MemberBanDialog,
	MemberKickDialog,
} from "@/features/group-settings/components/members/member-moderation-dialogs";
import { MemberProfileDialog } from "@/features/group-settings/components/members/member-profile-dialog";
import { MemberTimeoutDialog } from "@/features/group-settings/components/members/member-timeout-dialog";
import {
	getMemberModerationPolicy,
	resolveCallerBestRolePosition,
} from "@/features/group-settings/lib/member-moderation-policy";
import { sessionQueryOpts } from "@/services/auth/auth.query";
import { useTimeoutGroupMemberMutation } from "@/services/group/group.mutation";
import { myGroupsQueryOpts } from "@/services/group/group.query";
import { groupRolesQueryOpts } from "@/services/group-role/group-role.query";
import { groupMembersQueryOpts } from "@/services/invite/invite.query";
import { useSetUserRelationshipMutation } from "@/services/user/user.mutation";
import { userRelationshipsQueryOptions } from "@/services/user/user.query";

export type MemberProfileFallbackUser = {
	userId: string;
	username?: string;
	displayName?: string;
	avatarUrl?: string;
};

type GroupMemberProfileHostProps = {
	groupId: string;
	memberId: string | null;
	fallbackUser?: MemberProfileFallbackUser | null;
	onClose: () => void;
};

/**
 * Reusable host for opening `MemberProfileDialog` (and its relationship /
 * moderation sub-dialogs) from any group surface (member panel, chat, roles,
 * audit logs, invite access list).
 */
export function GroupMemberProfileHost({
	groupId,
	memberId,
	fallbackUser = null,
	onClose,
}: GroupMemberProfileHostProps) {
	const { t } = useTranslation();
	const isOpen = Boolean(groupId && memberId);

	const [ignoreMemberId, setIgnoreMemberId] = useState<string | null>(null);
	const [blockMemberId, setBlockMemberId] = useState<string | null>(null);
	const [timeoutMemberId, setTimeoutMemberId] = useState<string | null>(null);
	const [kickTarget, setKickTarget] = useState<GroupMember | null>(null);
	const [banTargets, setBanTargets] = useState<GroupMember[]>([]);

	const { data: session } = useQuery({
		...sessionQueryOpts(),
		enabled: isOpen,
	});
	const { data: myGroups = [] } = useQuery({
		...myGroupsQueryOpts(),
		enabled: isOpen,
	});
	const { data: rolesData } = useQuery({
		...groupRolesQueryOpts(groupId),
		enabled: isOpen,
	});
	const { data: membersData } = useQuery({
		...groupMembersQueryOpts(groupId, ""),
		enabled: isOpen,
	});
	const { data: relationships = [] } = useQuery({
		...userRelationshipsQueryOptions(),
		enabled: isOpen,
	});

	const setRelationship = useSetUserRelationshipMutation();
	const timeoutMutation = useTimeoutGroupMemberMutation(groupId);

	const members = useMemo(() => membersData?.members ?? [], [membersData?.members]);
	const allRoles = useMemo(() => rolesData?.roles ?? [], [rolesData?.roles]);
	const currentGroup = useMemo(() => myGroups.find((g) => g.id === groupId), [groupId, myGroups]);
	const hideRoles = currentGroup?.hide_roles ?? false;
	const actorUserId = session?.id;

	const actorBestRolePosition = useMemo(
		() => resolveCallerBestRolePosition(actorUserId, members),
		[actorUserId, members],
	);

	const relationshipByUserId = useMemo(() => {
		const map = new Map<string, UserRelationshipKind>();
		for (const rel of relationships) {
			map.set(rel.target_user_id, rel.kind);
		}
		return map;
	}, [relationships]);

	const resolvedMember = useMemo<GroupMember | null>(() => {
		if (!memberId) {
			return null;
		}
		const found = members.find((m) => m.user_id === memberId);
		if (found) {
			return found;
		}
		if (fallbackUser && fallbackUser.userId === memberId) {
			return GroupMemberSchema.parse({
				user_id: fallbackUser.userId,
				username: fallbackUser.username ?? "",
				display_name: fallbackUser.displayName ?? fallbackUser.username ?? "",
				avatar_url: fallbackUser.avatarUrl ?? "",
				role: "member",
				joined_at: "",
			});
		}
		return null;
	}, [fallbackUser, memberId, members]);

	const isActiveMember = Boolean(memberId && members.some((m) => m.user_id === memberId));

	const profilePolicy = useMemo(() => {
		if (!resolvedMember || !isActiveMember) {
			return null;
		}
		return getMemberModerationPolicy({
			actorUserId,
			actorGroup: currentGroup,
			actorBestRolePosition,
			target: resolvedMember,
			allRoles,
		});
	}, [actorBestRolePosition, actorUserId, allRoles, currentGroup, isActiveMember, resolvedMember]);

	const findDialogMember = useCallback(
		(id: string | null) => {
			if (!id) {
				return null;
			}
			if (resolvedMember?.user_id === id) {
				return resolvedMember;
			}
			return members.find((m) => m.user_id === id) ?? null;
		},
		[members, resolvedMember],
	);

	const ignoreMember = useMemo(
		() => findDialogMember(ignoreMemberId),
		[findDialogMember, ignoreMemberId],
	);
	const blockMember = useMemo(
		() => findDialogMember(blockMemberId),
		[blockMemberId, findDialogMember],
	);
	const timeoutMember = useMemo(
		() => findDialogMember(timeoutMemberId),
		[findDialogMember, timeoutMemberId],
	);

	const handleClearRelationship = useCallback(
		(member: GroupMember, previousKind: UserRelationshipKind) => {
			const name = member.display_name || member.username;
			setRelationship.mutate(
				{ targetUserId: member.user_id, kind: "none" },
				{
					onSuccess: () => {
						toast.success(
							t(
								previousKind === "ignore"
									? "group_settings.members.ignore_dialog.unignore_success"
									: "group_settings.members.block_dialog.unblock_success",
								{ name },
							),
						);
					},
					onError: () => {
						toast.error(t("error.unknown"));
					},
				},
			);
		},
		[setRelationship, t],
	);

	const handleClearTimeout = useCallback(
		(member: GroupMember) => {
			const name = member.display_name || member.username;
			timeoutMutation.mutate(
				{ userId: member.user_id, duration_seconds: 0, reason: "" },
				{
					onSuccess: () => {
						toast.success(
							t("group_settings.members.timeout_dialog.remove_success_toast", { name }),
						);
					},
					onError: () => {
						toast.error(t("common.error_generic"));
					},
				},
			);
		},
		[t, timeoutMutation],
	);

	return (
		<>
			<MemberProfileDialog
				groupId={groupId}
				groupName={currentGroup?.name}
				member={resolvedMember}
				policy={profilePolicy}
				relationshipKind={
					resolvedMember ? (relationshipByUserId.get(resolvedMember.user_id) ?? null) : null
				}
				allRoles={allRoles}
				hideRoles={hideRoles}
				onClose={onClose}
				onRequestIgnore={(member) => setIgnoreMemberId(member.user_id)}
				onRequestBlock={(member) => setBlockMemberId(member.user_id)}
				onClearRelationship={handleClearRelationship}
				onRequestTimeout={(member) => setTimeoutMemberId(member.user_id)}
				onClearTimeout={handleClearTimeout}
				onRequestKick={(member) => setKickTarget(member)}
				onRequestBan={(member) => setBanTargets([member])}
			/>

			<MemberIgnoreDialog
				member={ignoreMember}
				onClose={() => setIgnoreMemberId(null)}
				onSwitchToBlock={(member) => {
					setIgnoreMemberId(null);
					setBlockMemberId(member.user_id);
				}}
			/>

			<MemberBlockDialog
				member={blockMember}
				onClose={() => setBlockMemberId(null)}
				onSwitchToIgnore={(member) => {
					setBlockMemberId(null);
					setIgnoreMemberId(member.user_id);
				}}
			/>

			<MemberTimeoutDialog
				groupId={groupId}
				member={timeoutMember}
				onClose={() => setTimeoutMemberId(null)}
			/>

			<MemberKickDialog groupId={groupId} member={kickTarget} onClose={() => setKickTarget(null)} />

			<MemberBanDialog groupId={groupId} members={banTargets} onClose={() => setBanTargets([])} />
		</>
	);
}
