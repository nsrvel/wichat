"use client";

import type { OnChangeFn, RowSelectionState } from "@tanstack/react-table";
import { useQuery } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { GroupSettingsMembersSelectionFooter } from "@/features/group-settings/components/members/group-settings-members-selection-footer";
import { GroupSettingsMembersTable } from "@/features/group-settings/components/members/group-settings-members-table";
import { GroupSettingsMembersToolbar } from "@/features/group-settings/components/members/group-settings-members-toolbar";
import { groupSettingsMainInnerClass, groupSettingsMainScrollClass, groupSettingsMainScrollDirtyPadClass, groupSettingsMembersWorkspaceInnerClass, groupSettingsMembersWorkspaceScrollClass, groupSettingsSectionHeaderClass, groupSettingsSectionLeadClass, groupSettingsSectionTitleClass } from "@/features/group-settings/lib/group-settings-chrome";
import type { GroupSettingsTarget } from "@/features/group-settings/lib/group-settings-target";
import {
	applyMemberView,
	toMemberRow,
	type MembersSortKey,
} from "@/features/group-settings/lib/group-settings-members-view";
import { cn } from "@/lib/utils";
import { myGroupsQueryOpts } from "@/services/group/group.query";
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

	const { data: myGroups = [] } = useQuery(myGroupsQueryOpts());

	const isOwner = useMemo(
		() => myGroups.some((g) => g.id === groupId && g.role === "owner"),
		[groupId, myGroups],
	);
	const targetIsOwner = target.role === "owner";
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
	const canBulk = isOwner || targetIsOwner;
	const selectedCount = Object.keys(rowSelection).filter((id) => rowSelection[id]).length;

	const onBanSelected = () => {
		// ponytail: bulk ban RPC when group member moderation ships.
		toast.message(t("group_settings.members.ban_coming_soon"));
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
						canPrune={isOwner || targetIsOwner}
					/>

					<GroupSettingsMembersTable
						rows={visibleRows}
						loading={membersPending || membersQuery.isFetching}
						failed={membersFailed}
						canChangeRoles={canBulk}
						enableBulkSelect={canBulk}
						rowSelection={rowSelection}
						onRowSelectionChange={setRowSelection}
						pageIndex={cursorStack.length}
						total={total}
						hasPrev={hasPrev}
						hasNext={hasNext}
						maxVisitedPage={cursorStack.length}
						emptySearch={emptySearch}
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
		</div>
	);
}
