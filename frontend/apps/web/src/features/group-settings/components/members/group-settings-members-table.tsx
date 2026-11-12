"use client";

import { Menu } from "@base-ui/react/menu";
import {
	flexRender,
	getCoreRowModel,
	useReactTable,
	type ColumnDef,
	type OnChangeFn,
	type RowSelectionState,
} from "@tanstack/react-table";
import { Check, EllipsisVertical, Plus } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { GroupSettingsMemberRoleBadge } from "@/features/group-settings/components/members/group-settings-member-role-badge";
import { GroupSettingsMembersPagination } from "@/features/group-settings/components/members/group-settings-members-pagination";
import {
	formatRelativeUtc,
	formatShortDate,
} from "@/features/group-settings/lib/group-settings-members-format";
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
	MEMBERS_COLUMN_SIZE,
} from "@/features/group-settings/lib/group-settings-data-table-column-size";
import type { GroupSettingsMemberRow } from "@/features/group-settings/lib/group-settings-members-view";
import { userInitials } from "@/features/invite/lib/user-initials";
import {
	menuCheckboxBoxClass,
	menuCheckboxItemClass,
	menuPopupClass,
	menuPositionerClass,
} from "@/features/shell/components/group-rail-menu/menu-chrome";
import { cn } from "@/lib/utils";
import { MEMBER_PAGE_SIZE } from "@/services/invite/invite.query";

const MEMBER_ROLES = ["owner", "member"] as const;
type MemberRoleKey = (typeof MEMBER_ROLES)[number];

function normalizeSelectedRoles(roles: string[]): MemberRoleKey[] {
	const set = new Set<MemberRoleKey>();
	for (const r of roles) {
		if (r === "owner" || r === "member") {
			set.add(r);
		}
	}
	if (set.size === 0) {
		set.add("member");
	}
	return MEMBER_ROLES.filter((r) => set.has(r));
}

type MemberTableCheckboxProps = {
	checked: boolean;
	indeterminate?: boolean;
	disabled?: boolean;
	"aria-label": string;
	onCheckedChange: (next: boolean) => void;
};

function MemberTableCheckbox({
	checked,
	indeterminate = false,
	disabled,
	"aria-label": ariaLabel,
	onCheckedChange,
}: MemberTableCheckboxProps) {
	const ref = useRef<HTMLButtonElement>(null);

	useEffect(() => {
		if (ref.current) {
			ref.current.dataset.indeterminate = indeterminate ? "" : undefined;
		}
	}, [indeterminate]);

	return (
		<button
			ref={ref}
			type="button"
			role="checkbox"
			aria-checked={indeterminate ? "mixed" : checked}
			aria-label={ariaLabel}
			disabled={disabled}
			className={cn(
				menuCheckboxBoxClass,
				"group/checkbox",
				(checked || indeterminate) && "border-primary bg-primary text-primary-foreground",
			)}
			data-checked={checked || indeterminate ? "" : undefined}
			onClick={() => onCheckedChange(!checked)}
		>
			{checked ? <Check className="h-3 w-3" strokeWidth={2.5} aria-hidden /> : null}
			{indeterminate && !checked ? (
				<span className="block h-0.5 w-2 rounded-full bg-primary-foreground" aria-hidden />
			) : null}
		</button>
	);
}

type GroupSettingsMembersTableProps = {
	rows: GroupSettingsMemberRow[];
	loading: boolean;
	failed?: boolean;
	canChangeRoles?: boolean;
	enableBulkSelect?: boolean;
	rowSelection: RowSelectionState;
	onRowSelectionChange: OnChangeFn<RowSelectionState>;
	pageIndex: number;
	total: number;
	hasPrev: boolean;
	hasNext: boolean;
	maxVisitedPage: number;
	onPrev: () => void;
	onNext: () => void;
	onGoToPage: (page: number) => void;
	emptySearch?: boolean;
};

/** ponytail: local multi-role UI until SetMemberRoles ships; at least one role stays checked. */
function MemberRolesMenu({ memberId, roles }: { memberId: string; roles: string[] }) {
	const { t } = useTranslation();
	const [open, setOpen] = useState(false);
	const [selected, setSelected] = useState<MemberRoleKey[]>(() => normalizeSelectedRoles(roles));

	useEffect(() => {
		// eslint-disable-next-line react-hooks/set-state-in-effect -- sync local state when external roles/member change
		setSelected(normalizeSelectedRoles(roles));
	}, [memberId, roles]);

	const toggleRole = (role: MemberRoleKey, next: boolean) => {
		setSelected((prev) => {
			if (next) {
				return normalizeSelectedRoles([...prev, role]);
			}
			const filtered = prev.filter((r) => r !== role);
			return filtered.length === 0 ? prev : filtered;
		});
	};

	return (
		<div className="flex flex-wrap items-center gap-1">
			{selected.map((r) => (
				<GroupSettingsMemberRoleBadge key={r} role={r} />
			))}
			<Menu.Root open={open} onOpenChange={setOpen}>
				<Menu.Trigger
					nativeButton
					render={
						<button
							type="button"
							className={cn(
								"flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors",
								"hover:bg-sidebar-accent hover:text-foreground",
								open && "bg-sidebar-accent text-foreground",
							)}
							aria-haspopup="menu"
							aria-expanded={open}
							aria-label={t("group_settings.members.roles_menu_aria")}
						/>
					}
				>
					<Plus className="h-4 w-4" strokeWidth={2} aria-hidden />
				</Menu.Trigger>
				<Menu.Portal>
					<Menu.Positioner className={menuPositionerClass} side="bottom" align="end" sideOffset={4}>
						<Menu.Popup className={menuPopupClass}>
							{MEMBER_ROLES.map((role) => {
								const checked = selected.includes(role);
								return (
									<Menu.CheckboxItem
										key={role}
										className={cn(menuCheckboxItemClass, "group/checkbox")}
										checked={checked}
										closeOnClick={false}
										onCheckedChange={(next) => toggleRole(role, Boolean(next))}
									>
										<span className="min-w-0 flex-1 pr-1">{t(`invite.role_${role}`)}</span>
										<span className={menuCheckboxBoxClass} aria-hidden>
											<Menu.CheckboxItemIndicator
												keepMounted
												className="flex items-center justify-center opacity-0 data-[checked]:opacity-100"
											>
												<Check className="h-3 w-3" strokeWidth={2.5} />
											</Menu.CheckboxItemIndicator>
										</span>
									</Menu.CheckboxItem>
								);
							})}
						</Menu.Popup>
					</Menu.Positioner>
				</Menu.Portal>
			</Menu.Root>
		</div>
	);
}

function RoleBadges({
	roles,
	canChangeRoles,
	memberId,
}: {
	roles: string[];
	canChangeRoles: boolean;
	memberId: string;
}) {
	if (canChangeRoles) {
		return <MemberRolesMenu memberId={memberId} roles={roles} />;
	}

	if (roles.length === 0) {
		return <span className="text-muted-foreground">—</span>;
	}

	return (
		<div className="flex flex-wrap items-center gap-1">
			{roles.map((r) => (
				<GroupSettingsMemberRoleBadge key={r} role={r} />
			))}
		</div>
	);
}

const thClass =
	"px-3 py-2.5 text-left text-[11px] font-medium tracking-normal text-muted-foreground sm:text-xs";
const tdClass = "px-3 py-2.5 align-middle text-sm text-foreground";

export function GroupSettingsMembersTable({
	rows,
	loading,
	failed = false,
	canChangeRoles = false,
	enableBulkSelect = false,
	rowSelection,
	onRowSelectionChange,
	pageIndex,
	total,
	hasPrev,
	hasNext,
	maxVisitedPage,
	onPrev,
	onNext,
	onGoToPage,
	emptySearch = false,
}: GroupSettingsMembersTableProps) {
	const { t, i18n } = useTranslation();
	const locale = i18n.language;

	const columns = useMemo<ColumnDef<GroupSettingsMemberRow>[]>(() => {
		const defs: ColumnDef<GroupSettingsMemberRow>[] = [];

		if (enableBulkSelect) {
			defs.push({
				id: "select",
				...MEMBERS_COLUMN_SIZE.select,
				header: ({ table }) => (
					<MemberTableCheckbox
						checked={table.getIsAllPageRowsSelected()}
						indeterminate={table.getIsSomePageRowsSelected()}
						aria-label={t("group_settings.members.select_all_page")}
						onCheckedChange={(next) => table.toggleAllPageRowsSelected(next)}
					/>
				),
				cell: ({ row }) => {
					const name = row.original.member.display_name.trim() || t("invite.member_name_fallback");
					return (
						<MemberTableCheckbox
							checked={row.getIsSelected()}
							disabled={!row.getCanSelect()}
							aria-label={t("group_settings.members.select_member", { name })}
							onCheckedChange={(next) => row.toggleSelected(next)}
						/>
					);
				},
				enableSorting: false,
				enableHiding: false,
			});
		}

		defs.push(
			{
				id: "name",
				...MEMBERS_COLUMN_SIZE.name,
				header: () => t("group_settings.members.col_name"),
				accessorFn: (row) => row.member.display_name,
				cell: ({ row }) => {
					const { member } = row.original;
					const name = member.display_name.trim() || t("invite.member_name_fallback");
					return (
						<div className="flex min-w-0 items-center gap-3">
							<div
								className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-background text-xs font-semibold text-foreground ring-1 ring-border/60"
								aria-hidden
							>
								{userInitials(member.display_name, member.email)}
							</div>
							<div className="min-w-0">
								<p className="truncate font-medium">{name}</p>
								{member.email ? (
									<p className="truncate text-xs text-muted-foreground">{member.email}</p>
								) : null}
							</div>
						</div>
					);
				},
			},
			{
				id: "member_since",
				...MEMBERS_COLUMN_SIZE.member_since,
				header: () => t("group_settings.members.col_member_since"),
				accessorFn: (row) => row.memberSinceAt,
				cell: ({ getValue }) => {
					const iso = getValue<string>();
					return (
						<time dateTime={iso} className="whitespace-nowrap text-muted-foreground">
							{formatShortDate(iso, locale)}
						</time>
					);
				},
			},
			{
				id: "joined_wichat",
				...MEMBERS_COLUMN_SIZE.joined_wichat,
				header: () => t("group_settings.members.col_joined_wichat"),
				accessorFn: (row) => row.joinedWiChatAt,
				cell: ({ row }) => {
					const iso = row.original.joinedWiChatAt;
					if (!iso) {
						return (
							<span
								className="text-muted-foreground"
								title={t("group_settings.members.joined_wichat_pending")}
							>
								—
							</span>
						);
					}
					return (
						<time dateTime={iso} className="whitespace-nowrap text-muted-foreground">
							{formatShortDate(iso, locale)}
						</time>
					);
				},
			},
			{
				id: "join_method",
				...MEMBERS_COLUMN_SIZE.join_method,
				header: () => t("group_settings.members.col_join_method"),
				accessorFn: (row) => row.joinMethod,
				cell: ({ row }) => (
					<span className="text-muted-foreground">
						{t(`group_settings.members.join_method.${row.original.joinMethod}`)}
					</span>
				),
			},
			{
				id: "roles",
				...MEMBERS_COLUMN_SIZE.roles,
				header: () => t("group_settings.members.col_roles"),
				accessorFn: (row) => row.roles,
				cell: ({ row }) => (
					<RoleBadges
						memberId={row.original.member.user_id}
						roles={row.original.roles}
						canChangeRoles={canChangeRoles}
					/>
				),
			},
			{
				id: "last_updated",
				...MEMBERS_COLUMN_SIZE.last_updated,
				header: () => t("group_settings.members.col_last_updated"),
				accessorFn: (row) => row.lastUpdatedAt,
				cell: ({ getValue }) => {
					const iso = getValue<string>();
					return (
						<time
							dateTime={iso}
							title={formatShortDate(iso, locale)}
							className="whitespace-nowrap text-muted-foreground"
						>
							{formatRelativeUtc(iso, locale)}
						</time>
					);
				},
			},
			{
				id: "actions",
				...MEMBERS_COLUMN_SIZE.actions,
				header: () => null,
				enableSorting: false,
				cell: ({ row }) => {
					const name = row.original.member.display_name.trim() || t("invite.member_name_fallback");
					return (
						<div className="flex justify-end">
							<button
								type="button"
								className="flex h-8 min-w-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-foreground"
								aria-label={t("group_settings.members.row_actions_aria", { name })}
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
						</div>
					);
				},
			},
		);

		return defs;
	}, [canChangeRoles, enableBulkSelect, locale, t]);

	// eslint-disable-next-line react-hooks/incompatible-library -- TanStack Table returns unstable table instance by design.
	const table = useReactTable({
		data: rows,
		columns,
		getCoreRowModel: getCoreRowModel(),
		getRowId: (row) => row.member.user_id,
		enableRowSelection: enableBulkSelect,
		onRowSelectionChange,
		state: { rowSelection },
	});

	const showEmpty = !loading && !failed && rows.length === 0;
	const showFailed = failed;
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
										className={cn(thClass, dataTableEdgeCellClass(header.column.id))}
									>
										{header.isPlaceholder
											? null
											: flexRender(header.column.columnDef.header, header.getContext())}
									</th>
								))}
							</tr>
						))}
					</thead>
					<tbody className={groupSettingsDataTableBodyClass} aria-busy={loading && !failed}>
						{showSkeleton
							? [0, 1, 2, 3].map((key) => (
									<tr key={key} className="animate-pulse">
										<td colSpan={columns.length} className="px-4 py-3">
											<div className="h-10 rounded-lg bg-muted/70" />
										</td>
									</tr>
								))
							: null}
						{!showSkeleton && !showFailed && !showEmpty
							? table.getRowModel().rows.map((row) => (
									<tr key={row.id} className={groupSettingsDataTableRowClass}>
										{row.getVisibleCells().map((cell) => (
											<td
												key={cell.id}
												style={dataTableColumnPercentStyle(cell.column, leafColumns)}
												className={cn(tdClass, dataTableEdgeCellClass(cell.column.id))}
											>
												{flexRender(cell.column.columnDef.cell, cell.getContext())}
											</td>
										))}
									</tr>
								))
							: null}
						{showFailed ? (
							<tr>
								<td
									colSpan={columns.length}
									className="px-4 py-10 text-center text-sm text-muted-foreground"
								>
									{t("invite.members_load_error")}
								</td>
							</tr>
						) : null}
						{showEmpty ? (
							<tr>
								<td
									colSpan={columns.length}
									className="px-4 py-10 text-center text-sm text-muted-foreground"
								>
									{emptySearch
										? t("group_settings.members.search_empty")
										: t("invite.members_empty")}
								</td>
							</tr>
						) : null}
					</tbody>
				</table>
			</div>

			<div className="px-4 pb-3">
				<GroupSettingsMembersPagination
					pageIndex={pageIndex}
					pageSize={MEMBER_PAGE_SIZE}
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
