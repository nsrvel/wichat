"use client";

import type { GroupMember } from "@wichat/contracts/group/member";
import type { OnChangeFn, RowSelectionState } from "@tanstack/react-table";
import { useQuery } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { GroupSettingsMembersSelectionFooter } from "@/features/group-settings/components/members/group-settings-members-selection-footer";
import { GroupSettingsMembersTable } from "@/features/group-settings/components/members/group-settings-members-table";
import { GroupSettingsMembersToolbar } from "@/features/group-settings/components/members/group-settings-members-toolbar";
import {
	MemberBanDialog,
	MemberKickDialog,
} from "@/features/group-settings/components/members/member-moderation-dialogs";
import {
	groupSettingsMainInnerClass,
	groupSettingsMainScrollClass,
	groupSettingsMainScrollDirtyPadClass,
	groupSettingsMembersWorkspaceInnerClass,
	groupSettingsMembersWorkspaceScrollClass,
	groupSettingsSectionHeaderClass,
	groupSettingsSectionLeadClass,
	groupSettingsSectionTitleClass,
} from "@/features/group-settings/lib/group-settings-chrome";
import type { GroupSettingsTarget } from "@/features/group-settings/lib/group-settings-target";
import {
	applyMemberView,
	toMemberRow,
	type MembersSortKey,
} from "@/features/group-settings/lib/group-settings-members-view";
import { resolveCallerBestRolePosition } from "@/features/group-settings/lib/member-moderation-policy";
import { hasAnyGroupPermission, hasGroupPermission } from "@/lib/group-permissions";
import { cn } from "@/lib/utils";
import { sessionQueryOpts } from "@/services/auth/auth.query";
import { myGroupsQueryOpts } from "@/services/group/group.query";
import { groupRolesQueryOpts } from "@/services/group-role/group-role.query";
import { groupMembersQueryOpts } from "@/services/invite/invite.query";

type GroupSettingsMembersSectionProps = {
	target: GroupSettingsTarget;
	layout?: "settings" | "workspace";
};

export function GroupSettingsMembersSection({
	target,
	layout = "settings",
}: GroupSettingsMembersSectionProps) {
	const scrollClass =
		layout === "workspace"
			? groupSettingsMembersWorkspaceScrollClass
			: groupSettingsMainScrollClass;
	const innerClass =
		layout === "workspace"
			? groupSettingsMembersWorkspaceInnerClass
			: cn(groupSettingsMainInnerClass, "flex max-w-none flex-col gap-6 pb-4 lg:max-w-5xl");
	const { t } = useTranslation();
	const groupId = target.groupId;
	const [cursorStack, setCursorStack] = useState<string[]>([""]);
	const [search, setSearch] = useState("");
	const [sort, setSort] = useState<MembersSortKey>("name_asc");
	const cursor = cursorStack[cursorStack.length - 1] ?? "";
	const [selectionState, setSelectionState] = useState<{
		cursor: string;
		rows: RowSelectionState;
	}>({ cursor: "", rows: {} });
	const rowSelection = selectionState.cursor === cursor ? selectionState.rows : {};
	const setRowSelection: OnChangeFn<RowSelectionState> = (updater) => {
		setSelectionState((prev) => {
			const base = prev.cursor === cursor ? prev.rows : {};
			const next = typeof updater === "function" ? updater(base) : updater;
			return { cursor, rows: next };
		});
	};

	const { data: session } = useQuery(sessionQueryOpts());
	const actorUserId = session?.id;
	const { data: myGroups = [] } = useQuery(myGroupsQueryOpts());
	const rolesQuery = useQuery({
		...groupRolesQueryOpts(groupId),
		enabled: Boolean(groupId),
	});
	const allRoles = useMemo(() => rolesQuery.data?.roles ?? [], [rolesQuery.data?.roles]);
	const [kickTarget, setKickTarget] = useState<GroupMember | null>(null);
	const [banTargets, setBanTargets] = useState<GroupMember[]>([]);

	const currentGroup = useMemo(() => myGroups.find((g) => g.id === groupId), [groupId, myGroups]);
	const isOwner = currentGroup?.role === "owner";
	const hideRoles = currentGroup?.hide_roles ?? false;
	const targetIsOwner = target.role === "owner";
	const effectiveGroup = currentGroup ?? {
		role: target.role ?? "member",
		effective_permissions: [],
	};
	const membersQuery = useQuery({
		...groupMembersQueryOpts(groupId, cursor),
		enabled: Boolean(groupId),
	});

	const membersPage = membersQuery.data;
	const members = useMemo(() => membersPage?.members ?? [], [membersPage?.members]);
	const total = membersPage?.total ?? members.length;
	const membersPending = membersQuery.isPending && !membersPage;
	const membersFailed = membersQuery.isError;
	const hasNext = Boolean(membersPage?.next_cursor);
	const hasPrev = cursorStack.length > 1;

	const rows = useMemo(() => members.map(toMemberRow), [members]);
	const visibleRows = useMemo(() => applyMemberView(rows, { search, sort }), [rows, search, sort]);

	const emptySearch = search.trim().length > 0 && visibleRows.length === 0 && rows.length > 0;
	const actorBestRolePosition = useMemo(
		() => resolveCallerBestRolePosition(actorUserId, members),
		[actorUserId, members],
	);
	const canPrune = isOwner || targetIsOwner || hasGroupPermission(effectiveGroup, "kick_members");
	const canBulk =
		isOwner ||
		targetIsOwner ||
		hasAnyGroupPermission(effectiveGroup, ["kick_members", "ban_members"]);
	const selectedCount = Object.keys(rowSelection).filter((id) => rowSelection[id]).length;

	const onBanSelected = () => {
		const selectedIds = new Set(Object.keys(rowSelection).filter((id) => rowSelection[id]));
		const selectedMembers = members.filter((m) => selectedIds.has(m.user_id));
		if (selectedMembers.length > 0) {
			setBanTargets(selectedMembers);
		}
	};

	return (
		<div className="flex min-h-0 flex-1 flex-col">
			<div
				className={cn(
					scrollClass,
					selectedCount > 0 && layout !== "workspace" && groupSettingsMainScrollDirtyPadClass,
				)}
			>
				<div className={innerClass}>
					{layout === "settings" ? (
						<div className={groupSettingsSectionHeaderClass}>
							<h2 className={groupSettingsSectionTitleClass}>
								{t("group_settings.members.title")}
							</h2>
							<p className={groupSettingsSectionLeadClass}>{t("group_settings.members.lead")}</p>
						</div>
					) : null}

					<GroupSettingsMembersToolbar
						search={search}
						onSearchChange={setSearch}
						sort={sort}
						onSortChange={setSort}
						groupName={target.groupName}
						rows={rows}
						canPrune={canPrune}
					/>

					<GroupSettingsMembersTable
						groupId={groupId}
						rows={visibleRows}
						loading={membersPending}
						failed={membersFailed}
						actorUserId={actorUserId}
						actorGroup={effectiveGroup}
						actorBestRolePosition={actorBestRolePosition}
						allRoles={allRoles}
						enableBulkSelect={canBulk}
						hideRoles={hideRoles}
						rowSelection={rowSelection}
						onRowSelectionChange={setRowSelection}
						pageIndex={cursorStack.length}
						total={total}
						hasPrev={hasPrev}
						hasNext={hasNext}
						maxVisitedPage={cursorStack.length}
						emptySearch={emptySearch}
						onRequestKick={setKickTarget}
						onRequestBan={(member) => setBanTargets([member])}
						onPrev={() => setCursorStack((s) => (s.length > 1 ? s.slice(0, -1) : s))}
						onNext={() => {
							const next = membersQuery.data?.next_cursor;
							if (next) {
								setCursorStack((s) => [...s, next]);
							}
						}}
						onGoToPage={(page) => {
							if (page < 1) {
								return;
							}
							setCursorStack((s) => {
								if (page <= s.length) {
									return s.slice(0, page);
								}
								return s;
							});
						}}
					/>
					{layout === "workspace" ? (
						<GroupSettingsMembersSelectionFooter
							count={selectedCount}
							layout="workspace"
							onClear={() => setRowSelection({})}
							onBan={onBanSelected}
						/>
					) : null}
				</div>
			</div>
			{layout !== "workspace" ? (
				<GroupSettingsMembersSelectionFooter
					count={selectedCount}
					layout="settings"
					onClear={() => setRowSelection({})}
					onBan={onBanSelected}
				/>
			) : null}

			<MemberKickDialog groupId={groupId} member={kickTarget} onClose={() => setKickTarget(null)} />
			<MemberBanDialog
				groupId={groupId}
				members={banTargets}
				onClose={() => setBanTargets([])}
				onBanned={() => setRowSelection({})}
			/>
		</div>
	);
}
