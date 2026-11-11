"use client";

import { flexRender, getCoreRowModel, useReactTable, type ColumnDef } from "@tanstack/react-table";
import { EllipsisVertical } from "lucide-react";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { GroupSettingsMembersPagination } from "@/features/group-settings/components/members/group-settings-members-pagination";
import {
	groupSettingsDataTableBodyClass,
	groupSettingsDataTableHeadRowClass,
	groupSettingsDataTableRowClass,
	groupSettingsDataTableShellClass,
} from "@/features/group-settings/lib/group-settings-chrome";
import {
	dataTableColumnPercentStyle,
	dataTableEdgeCellClass,
	dataTableMinTableStyle,
	groupSettingsDataTableTdClass,
	groupSettingsDataTableThClass,
	ROLES_COLUMN_SIZE,
} from "@/features/group-settings/lib/group-settings-data-table-column-size";
import type { GroupSettingsRoleRow } from "@/features/group-settings/lib/group-settings-roles-view";
import { cn } from "@/lib/utils";

const ROLE_PAGE_SIZE = 10;

type GroupSettingsRolesTableProps = {
	rows: GroupSettingsRoleRow[];
	loading: boolean;
	total: number;
	pageIndex: number;
	hasPrev: boolean;
	hasNext: boolean;
	maxVisitedPage: number;
	onOpenRole: (roleId: string) => void;
	onPrev: () => void;
	onNext: () => void;
	onGoToPage: (page: number) => void;
};

export function GroupSettingsRolesTable({
	rows,
	loading,
	total,
	pageIndex,
	hasPrev,
	hasNext,
	maxVisitedPage,
	onOpenRole,
	onPrev,
	onNext,
	onGoToPage,
}: GroupSettingsRolesTableProps) {
	const { t } = useTranslation();

	const columns = useMemo<ColumnDef<GroupSettingsRoleRow>[]>(
		() => [
			{
				id: "role",
				...ROLES_COLUMN_SIZE.role,
				header: () => t("group_settings.roles.col_role"),
				accessorFn: (row) => row.name,
				cell: ({ row }) => (
					<span className="block truncate font-medium text-foreground">
						{row.original.isSystem ? t(`invite.role_${row.original.name}`) : row.original.name}
					</span>
				),
			},
			{
				id: "members",
				...ROLES_COLUMN_SIZE.members,
				header: () => t("group_settings.roles.col_members"),
				accessorFn: (row) => row.memberCount,
				cell: ({ getValue }) => (
					<span className="tabular-nums text-muted-foreground">{getValue<number>()}</span>
				),
			},
			{
				id: "actions",
				...ROLES_COLUMN_SIZE.actions,
				header: () => null,
				enableSorting: false,
				cell: ({ row }) => {
					const label = row.original.isSystem
						? t(`invite.role_${row.original.name}`)
						: row.original.name;
					return (
						<button
							type="button"
							className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-sidebar-accent hover:text-foreground"
							aria-label={t("group_settings.roles.row_actions_aria", { name: label })}
							onClick={(event) => {
								event.stopPropagation();
							}}
							onContextMenu={(event) => {
								event.preventDefault();
								event.stopPropagation();
							}}
						>
							<EllipsisVertical className="h-4 w-4" strokeWidth={2} aria-hidden />
						</button>
					);
				},
			},
		],
		[t],
	);

	// eslint-disable-next-line react-hooks/incompatible-library -- TanStack Table returns unstable table instance by design.
	const table = useReactTable({
		data: rows,
		columns,
		getCoreRowModel: getCoreRowModel(),
		getRowId: (row) => row.id,
	});

	const showEmpty = !loading && rows.length === 0;
	const showSkeleton = loading && rows.length === 0;

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
										className={cn("cursor-pointer", groupSettingsDataTableRowClass)}
										onClick={() => onOpenRole(row.original.id)}
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
									{t("group_settings.roles.search_empty")}
								</td>
							</tr>
						) : null}
					</tbody>
				</table>
			</div>

			<div className="px-4 pb-3">
				<GroupSettingsMembersPagination
					variant="roles"
					pageIndex={pageIndex}
					pageSize={ROLE_PAGE_SIZE}
					pageRowCount={rows.length}
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
