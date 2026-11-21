"use client";

import type { GroupMember } from "@wichat/contracts/group/member";
import { useQuery } from "@tanstack/react-query";
import { Search, X } from "lucide-react";
import { useId, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { GroupMemberProfileHost } from "@/components/group/group-member-profile-host";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { OwnerCrownIcon } from "@/components/ui/owner-crown-icon";
import { UserAvatar } from "@/components/ui/user-avatar";
import { RoleAddMembersDialog } from "@/features/group-settings/components/roles/detail/role-add-members-dialog";
import {
	groupSettingsToolbarActionButtonClass,
	groupSettingsToolbarSearchInputClass,
} from "@/features/group-settings/lib/group-settings-chrome";
import {
	roleDisplayTitle,
	type GroupSettingsRoleRow,
} from "@/features/group-settings/lib/group-settings-roles-view";
import { ApiError } from "@wichat/api-client/api-error";
import { authInputClass } from "@/lib/form-field-styles";
import { resolveMemberNameColor } from "@/lib/member-role-display";
import { memberPresenceDot } from "@/features/presence/presence-dot";
import { userInitials } from "@/lib/user-initials";
import { cn } from "@/lib/utils";
import { useRemoveGroupRoleMemberMutation } from "@/services/group-role/group-role.mutation";
import { groupRoleMembersQueryOpts } from "@/services/group-role/group-role.query";

type RoleDetailMembersTabProps = {
	groupId: string;
	role: GroupSettingsRoleRow;
	allMembers: GroupMember[];
	canManage: boolean;
	hideRoles?: boolean;
};

export function RoleDetailMembersTab({
	groupId,
	role,
	allMembers,
	canManage,
	hideRoles = false,
}: RoleDetailMembersTabProps) {
	const { t } = useTranslation();
	const searchId = useId();
	const [search, setSearch] = useState("");
	const [addOpen, setAddOpen] = useState(false);
	const [selectedMemberId, setSelectedMemberId] = useState<string | null>(null);

	const isEveryone = role.isDefault || role.slug === "everyone";
	const roleLabel = roleDisplayTitle(role, t);

	const membersQuery = useQuery({
		...groupRoleMembersQueryOpts(groupId, role.id),
		enabled: Boolean(groupId && role.id && !isEveryone),
	});
	const removeMutation = useRemoveGroupRoleMemberMutation(groupId, role.id);

	const assignedMembers = useMemo(
		() => (isEveryone ? allMembers : (membersQuery.data ?? [])),
		[allMembers, isEveryone, membersQuery.data],
	);

	const assignedUserIds = useMemo(
		() => new Set(assignedMembers.map((m) => m.user_id)),
		[assignedMembers],
	);

	const memberRolesByUserId = useMemo(() => {
		const map = new Map<string, GroupMember["roles"]>();
		for (const m of allMembers) {
			map.set(m.user_id, m.roles);
		}
		return map;
	}, [allMembers]);

	const visibleMembers = useMemo(() => {
		const q = search.trim().toLowerCase();
		if (!q) {
			return assignedMembers;
		}
		return assignedMembers.filter(
			(m) =>
				m.display_name.toLowerCase().includes(q) || (m.username ?? "").toLowerCase().includes(q),
		);
	}, [assignedMembers, search]);

	const handleRemoveMember = async (member: GroupMember) => {
		const memberName = member.display_name.trim() || t("invite.member_name_fallback");
		try {
			await removeMutation.mutateAsync(member.user_id);
			toast.success(
				t("group_settings.roles.detail.remove_member_success", {
					member: memberName,
					name: roleLabel,
				}),
			);
		} catch (error) {
			toast.error(error instanceof ApiError ? error.message : t("common.error_generic"));
		}
	};

	const loading = !isEveryone && membersQuery.isPending && !membersQuery.data;

	return (
		<div className="flex flex-col gap-5 pt-5">
			{isEveryone ? (
				<div className="rounded-xl border border-border/80 bg-field/40 px-4 py-3 text-xs leading-relaxed text-muted-foreground">
					{t("group_settings.roles.detail.everyone_members_notice")}
				</div>
			) : null}

			<div className="flex flex-col gap-2 sm:flex-row sm:items-center">
				<label className="relative block min-w-0 flex-1" htmlFor={searchId}>
					<Search
						className="pointer-events-none absolute top-1/2 left-3.5 z-[1] size-4 -translate-y-1/2 text-muted-foreground"
						aria-hidden
					/>
					<Input
						id={searchId}
						value={search}
						onChange={(e) => setSearch(e.target.value)}
						placeholder={t("group_settings.roles.detail.members_search_placeholder")}
						className={cn(authInputClass, groupSettingsToolbarSearchInputClass)}
						autoComplete="off"
					/>
				</label>

				{!isEveryone && canManage ? (
					<Button
						type="button"
						variant="default"
						className={groupSettingsToolbarActionButtonClass}
						onClick={() => setAddOpen(true)}
					>
						{t("group_settings.roles.detail.add_members")}
					</Button>
				) : null}
			</div>

			{loading ? (
				<div className="flex flex-col gap-2">
					{[0, 1, 2].map((key) => (
						<div key={key} className="h-14 animate-pulse rounded-xl bg-muted/60" />
					))}
				</div>
			) : visibleMembers.length === 0 ? (
				<div className="rounded-xl border border-dashed border-border/80 bg-field/30 px-4 py-8 text-center text-sm text-muted-foreground">
					{t("group_settings.roles.detail.members_empty")}
				</div>
			) : (
				<div className="flex flex-col divide-y divide-border/80 rounded-xl bg-field/40">
					{visibleMembers.map((member) => {
						const name = member.display_name.trim() || t("invite.member_name_fallback");
						const rolesForColor =
							member.roles && member.roles.length > 0
								? member.roles
								: memberRolesByUserId.get(member.user_id);
						const nameColor =
							resolveMemberNameColor(rolesForColor, hideRoles) ??
							(!hideRoles
								? resolveMemberNameColor(
										[
											{
												id: role.id,
												slug: role.slug,
												name: role.name,
												color: role.color,
												hoist: role.hoist,
												position: 0,
											},
										],
										hideRoles,
									)
								: undefined);
						return (
							<div
								key={member.user_id}
								className="flex items-center justify-between gap-3 px-3 py-2.5 transition-colors first:rounded-t-xl last:rounded-b-xl hover:bg-background/40"
							>
								<button
									type="button"
									onClick={() => setSelectedMemberId(member.user_id)}
									className="flex min-w-0 flex-1 cursor-pointer items-center gap-3 text-left"
								>
									<UserAvatar
										src={member.avatar_url}
										initials={userInitials(name, member.username)}
										size="md"
										status={memberPresenceDot(member)}
										statusRingClassName="ring-card"
									/>
									<div className="min-w-0">
										<div className="flex min-w-0 items-center gap-1.5">
											<p
												className="truncate text-sm font-medium text-foreground"
												style={nameColor ? { color: nameColor } : undefined}
											>
												{name}
											</p>
											{member.role === "owner" ? (
												<OwnerCrownIcon label={t("invite.role_owner")} />
											) : null}
										</div>
										{member.username ? (
											<p className="truncate text-xs text-muted-foreground">@{member.username}</p>
										) : null}
									</div>
								</button>

								{!isEveryone && canManage ? (
									<button
										type="button"
										disabled={removeMutation.isPending}
										onClick={() => void handleRemoveMember(member)}
										aria-label={t("group_settings.roles.detail.remove_member_aria", {
											name,
										})}
										className="inline-flex size-7 shrink-0 items-center justify-center rounded-full text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-foreground disabled:pointer-events-none disabled:opacity-45"
									>
										<X className="size-4" strokeWidth={2} aria-hidden />
									</button>
								) : null}
							</div>
						);
					})}
				</div>
			)}

			<RoleAddMembersDialog
				open={addOpen}
				onOpenChange={setAddOpen}
				groupId={groupId}
				roleId={role.id}
				roleLabel={roleLabel}
				allMembers={allMembers}
				assignedUserIds={assignedUserIds}
				hideRoles={hideRoles}
			/>

			<GroupMemberProfileHost
				groupId={groupId}
				memberId={selectedMemberId}
				onClose={() => setSelectedMemberId(null)}
			/>
		</div>
	);
}
