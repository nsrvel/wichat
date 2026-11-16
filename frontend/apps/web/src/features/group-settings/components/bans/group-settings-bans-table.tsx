"use client";

import { flexRender, getCoreRowModel, useReactTable, type ColumnDef } from "@tanstack/react-table";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import type { GroupBanEntry } from "@wichat/contracts/group/member";
import { Button } from "@/components/ui/button";
import { UserAvatar } from "@/components/ui/user-avatar";
import { GroupSettingsMembersPagination } from "@/features/group-settings/components/members/group-settings-members-pagination";
import {
	groupSettingsDataTableBodyClass,
	groupSettingsDataTableHeadRowClass,
	groupSettingsDataTableRowClass,
	groupSettingsDataTableShellClass,
} from "@/features/group-settings/lib/group-settings-chrome";
import {
	BANS_COLUMN_SIZE,
	dataTableColumnPercentStyle,
	dataTableEdgeCellClass,
	dataTableMinTableStyle,
	groupSettingsDataTableTdClass,
	groupSettingsDataTableThClass,
} from "@/features/group-settings/lib/group-settings-data-table-column-size";
import { userInitials } from "@/lib/user-initials";
import { cn } from "@/lib/utils";

const BAN_PAGE_SIZE = 20;

type GroupSettingsBansTableProps = {
	bans: GroupBanEntry[];
	loading: boolean;
	total: number;
	pageIndex: number;
	hasPrev: boolean;
	hasNext: boolean;
	maxVisitedPage: number;
	searchQuery: string;
	unbanningUserId?: string;
	onPrev: () => void;
	onNext: () => void;
	onGoToPage: (page: number) => void;
	onRequestUnban: (entry: GroupBanEntry) => void;
};

function formatExpiry(
	expiresAt: string,
	t: (key: string, opts?: Record<string, string>) => string,
) {
	if (!expiresAt) {
		return t("group_settings.bans.permanent");
	}
	const date = new Date(expiresAt);
	if (Number.isNaN(date.getTime())) {
		return expiresAt;
	}
	return t("group_settings.bans.expires_on", {
		date: new Intl.DateTimeFormat(undefined, {
			dateStyle: "medium",
			timeStyle: "short",
		}).format(date),
	});
}

export function GroupSettingsBansTable({
	bans,
	loading,
	total,
	pageIndex,
	hasPrev,
	hasNext,
	maxVisitedPage,
	searchQuery,
	unbanningUserId,
	onPrev,
	onNext,
	onGoToPage,
	onRequestUnban,
}: GroupSettingsBansTableProps) {
	const { t } = useTranslation();

	const columns = useMemo<ColumnDef<GroupBanEntry>[]>(
		() => [
			{
				id: "user",
				...BANS_COLUMN_SIZE.user,
				header: () => t("group_settings.bans.col_user"),
				accessorFn: (row) => row.display_name || row.username,
				cell: ({ row }) => {
					const entry = row.original;
					const displayName = entry.display_name || entry.username;
					return (
						<div className="flex min-w-0 items-center gap-3">
							<UserAvatar
								src={entry.avatar_url || null}
								alt={displayName}
								initials={userInitials(displayName, entry.username)}
								size="sm"
							/>
							<div className="min-w-0">
								<p className="truncate font-medium text-foreground">{displayName}</p>
								{entry.username ? (
									<p className="truncate text-xs text-muted-foreground">@{entry.username}</p>
								) : null}
							</div>
						</div>
					);
				},
			},
			{
				id: "reason",
				...BANS_COLUMN_SIZE.reason,
				header: () => t("group_settings.bans.col_reason"),
				accessorFn: (row) => row.reason,
				cell: ({ row }) =>
					row.original.reason ? (
						<span className="line-clamp-2 text-sm text-foreground/90">{row.original.reason}</span>
					) : (
						<span className="italic text-muted-foreground">
							{t("group_settings.bans.no_reason")}
						</span>
					),
			},
			{
				id: "banned_by",
				...BANS_COLUMN_SIZE.banned_by,
				header: () => t("group_settings.bans.col_banned_by"),
				accessorFn: (row) => row.banned_by_display_name || row.banned_by_username,
				cell: ({ getValue }) => (
					<span className="block truncate text-muted-foreground">
						{getValue<string>() || t("group_settings.bans.unknown_actor")}
					</span>
				),
			},
			{
				id: "expires",
				...BANS_COLUMN_SIZE.expires,
				header: () => t("group_settings.bans.col_expires"),
				accessorFn: (row) => row.expires_at,
				cell: ({ getValue }) => (
					<span className="block truncate whitespace-nowrap text-muted-foreground">
						{formatExpiry(getValue<string>(), t)}
					</span>
				),
			},
			{
				id: "actions",
				...BANS_COLUMN_SIZE.actions,
				header: () => null,
				enableSorting: false,
				cell: ({ row }) => (
					<div className="flex justify-end">
						<Button
							type="button"
							variant="outline"
							size="sm"
							disabled={unbanningUserId === row.original.user_id}
							onClick={() => onRequestUnban(row.original)}
						>
							{t("group_settings.bans.unban_button")}
						</Button>
					</div>
				),
			},
		],
		[onRequestUnban, t, unbanningUserId],
	);

	// eslint-disable-next-line react-hooks/incompatible-library -- TanStack Table returns unstable table instance by design.
	const table = useReactTable({
		data: bans,
		columns,
		getCoreRowModel: getCoreRowModel(),
		getRowId: (row) => row.user_id,
	});

	const showEmpty = !loading && bans.length === 0;
	const showSkeleton = loading && bans.length === 0;

	const leafColumns = table.getVisibleLeafColumns();
	const tableWidthStyle = dataTableMinTableStyle(leafColumns);

	return (
		<div className={groupSettingsDataTableShellClass}>
			<div className="thin-scroll overflow-x-auto">
				<table className="w-full table-fixed border-collapse text-left" style={tableWidthStyle}>
					<thead>
						{table.getHeaderGroups().map((headerGroup) => (
							<tr key={headerGroup.id} className={groupSettingsDataTableHeadRowClass}>
								{headerGroup.headers.map((header) => (
									<th
										key={header.id}
										scope="col"
										style={dataTableColumnPercentStyle(header.column, leafColumns)}
										className={cn(
											groupSettingsDataTableThClass,
											dataTableEdgeCellClass(header.column.id),
										)}
									>
										{header.isPlaceholder
											? null
											: flexRender(header.column.columnDef.header, header.getContext())}
									</th>
								))}
							</tr>
						))}
					</thead>
					<tbody className={groupSettingsDataTableBodyClass} aria-busy={loading}>
						{showSkeleton
							? [0, 1, 2, 3].map((key) => (
									<tr key={key} className="animate-pulse">
										<td colSpan={columns.length} className="px-4 py-3">
											<div className="h-10 rounded-lg bg-muted/70" />
										</td>
									</tr>
								))
							: null}
						{!showSkeleton && !showEmpty
							? table.getRowModel().rows.map((row) => (
									<tr
										key={row.id}
										data-testid={`ban-row-${row.original.user_id}`}
										className={groupSettingsDataTableRowClass}
									>
										{row.getVisibleCells().map((cell) => (
											<td
												key={cell.id}
												style={dataTableColumnPercentStyle(cell.column, leafColumns)}
												className={cn(
													groupSettingsDataTableTdClass,
													dataTableEdgeCellClass(cell.column.id),
												)}
											>
												{flexRender(cell.column.columnDef.cell, cell.getContext())}
											</td>
										))}
									</tr>
								))
							: null}
						{showEmpty ? (
							<tr>
								<td
									colSpan={columns.length}
									className="px-4 py-10 text-center text-sm text-muted-foreground"
								>
									{searchQuery
										? t("group_settings.bans.search_empty", { query: searchQuery })
										: t("group_settings.bans.empty")}
								</td>
							</tr>
						) : null}
					</tbody>
				</table>
			</div>

			<div className="px-4 pb-3">
				<GroupSettingsMembersPagination
					variant="bans"
					pageIndex={pageIndex}
					pageSize={BAN_PAGE_SIZE}
					pageRowCount={bans.length}
					total={total}
					hasPrev={hasPrev}
					hasNext={hasNext}
					maxVisitedPage={maxVisitedPage}
					loading={loading}
					onPrev={onPrev}
					onNext={onNext}
					onGoToPage={onGoToPage}
				/>
			</div>
		</div>
	);
}
