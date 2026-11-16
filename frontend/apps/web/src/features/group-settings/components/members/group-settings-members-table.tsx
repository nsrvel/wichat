"use client";

import type { GroupMember } from "@wichat/contracts/group/member";
import type { GroupRole } from "@wichat/contracts/group/role";
import type { UserRelationshipKind } from "@wichat/contracts/user";
import { useQuery } from "@tanstack/react-query";
import {
	flexRender,
	getCoreRowModel,
	useReactTable,
	type ColumnDef,
	type OnChangeFn,
	type RowSelectionState,
} from "@tanstack/react-table";
import { Check, EllipsisVertical } from "lucide-react";
import { useCallback, useEffect, useMemo, useRef, useState, type MouseEvent } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { OwnerCrownIcon } from "@/components/ui/owner-crown-icon";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { UserAvatar } from "@/components/ui/user-avatar";
import {
	createMemberContextMenuState,
	MemberActionMenu,
	type MemberActionMenuState,
} from "@/features/group-settings/components/members/member-action-menu";
import { MemberBlockDialog } from "@/features/group-settings/components/members/member-block-dialog";
import { MemberIgnoreDialog } from "@/features/group-settings/components/members/member-ignore-dialog";
import { MemberProfileDialog } from "@/features/group-settings/components/members/member-profile-dialog";
import { MemberRolePickerPopover } from "@/features/group-settings/components/members/member-role-picker-popover";
import { MemberTimeoutDialog } from "@/features/group-settings/components/members/member-timeout-dialog";
import { GroupSettingsMembersPagination } from "@/features/group-settings/components/members/group-settings-members-pagination";
import { useTimeoutGroupMemberMutation } from "@/services/group/group.mutation";
import { useSetUserRelationshipMutation } from "@/services/user/user.mutation";
import { userRelationshipsQueryOptions } from "@/services/user/user.query";
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
import {
	getMemberModerationPolicy,
	type MemberModerationPolicy,
} from "@/features/group-settings/lib/member-moderation-policy";
import { isNativeContextMenuAllowed } from "@/lib/context-menu-chrome";
import type { GroupPermissionSubject } from "@/lib/group-permissions";
import { resolveMemberHoistedRole, resolveMemberNameColor } from "@/lib/member-role-display";
import { menuCheckboxBoxClass } from "@/lib/menu-chrome";
import { userInitials } from "@/lib/user-initials";
import { cn } from "@/lib/utils";
import { MEMBER_PAGE_SIZE } from "@/services/invite/invite.query";

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
			{checked ? <Check className="size-3" strokeWidth={2.5} aria-hidden /> : null}
			{indeterminate && !checked ? (
				<span className="block h-0.5 w-2 rounded-full bg-primary-foreground" aria-hidden />
			) : null}
		</button>
	);
}

type GroupSettingsMembersTableProps = {
	groupId: string;
	rows: GroupSettingsMemberRow[];
	loading: boolean;
	failed?: boolean;
	actorUserId?: string;
	actorGroup: GroupPermissionSubject | null | undefined;
	actorBestRolePosition: number;
	allRoles: readonly GroupRole[];
	enableBulkSelect?: boolean;
	hideRoles?: boolean;
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
	onRequestKick: (member: GroupMember) => void;
	onRequestBan: (member: GroupMember) => void;
	emptySearch?: boolean;
};

const thClass =
	"px-3 py-2.5 text-left text-[11px] font-medium tracking-normal text-muted-foreground sm:text-xs";
const tdClass = "px-3 py-2.5 align-middle text-sm text-foreground";

export function GroupSettingsMembersTable({
	groupId,
	rows,
	loading,
	failed = false,
	actorUserId,
	actorGroup,
	actorBestRolePosition,
	allRoles,
	enableBulkSelect = false,
	hideRoles = false,
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
	onRequestKick,
	onRequestBan,
	emptySearch = false,
}: GroupSettingsMembersTableProps) {
	const { t, i18n } = useTranslation();
	const locale = i18n.language;
	const [menuState, setMenuState] = useState<MemberActionMenuState | null>(null);
	const [profileMemberId, setProfileMemberId] = useState<string | null>(null);
	const [ignoreMemberId, setIgnoreMemberId] = useState<string | null>(null);
	const [blockMemberId, setBlockMemberId] = useState<string | null>(null);
	const [timeoutMemberId, setTimeoutMemberId] = useState<string | null>(null);

	const { data: relationships = [] } = useQuery(userRelationshipsQueryOptions());
	const setRelationship = useSetUserRelationshipMutation();
	const timeoutMutation = useTimeoutGroupMemberMutation(groupId);

	const relationshipByUserId = useMemo(() => {
		const map = new Map<string, UserRelationshipKind>();
		for (const rel of relationships) {
			map.set(rel.target_user_id, rel.kind);
		}
		return map;
	}, [relationships]);

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

	const policyByUserId = useMemo(() => {
		const map = new Map<string, MemberModerationPolicy>();
		for (const row of rows) {
			map.set(
				row.member.user_id,
				getMemberModerationPolicy({
					actorUserId,
					actorGroup,
					actorBestRolePosition,
					target: row.member,
					allRoles,
				}),
			);
		}
		return map;
	}, [actorBestRolePosition, actorGroup, actorUserId, allRoles, rows]);

	const profileMember = useMemo(
		() =>
			profileMemberId
				? (rows.find((row) => row.member.user_id === profileMemberId)?.member ?? null)
				: null,
		[rows, profileMemberId],
	);
	const ignoreMember = useMemo(
		() =>
			ignoreMemberId
				? (rows.find((row) => row.member.user_id === ignoreMemberId)?.member ?? null)
				: null,
		[rows, ignoreMemberId],
	);
	const blockMember = useMemo(
		() =>
			blockMemberId
				? (rows.find((row) => row.member.user_id === blockMemberId)?.member ?? null)
				: null,
		[rows, blockMemberId],
	);
	const timeoutMember = useMemo(
		() =>
			timeoutMemberId
				? (rows.find((row) => row.member.user_id === timeoutMemberId)?.member ?? null)
				: null,
		[rows, timeoutMemberId],
	);
	const profilePolicy = profileMember ? (policyByUserId.get(profileMember.user_id) ?? null) : null;

	const openRowContextMenu = useCallback(
		(event: MouseEvent<HTMLElement>, member: GroupMember) => {
			if (isNativeContextMenuAllowed(event.target)) {
				return;
			}
			const policy = policyByUserId.get(member.user_id);
			if (!policy) {
				return;
			}
			event.preventDefault();
			event.stopPropagation();
			setMenuState(createMemberContextMenuState(member, policy, event.clientX, event.clientY));
		},
		[policyByUserId],
	);

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
					const nameColor = resolveMemberNameColor(member.roles, hideRoles);
					return (
						<div className="flex min-w-0 items-center gap-3">
							<UserAvatar
								src={member.avatar_url}
								initials={userInitials(member.display_name, member.username)}
								size="md"
								status={member.presence_status}
								statusRingClassName="ring-card"
							/>
							<div className="min-w-0">
								<div className="flex min-w-0 items-center gap-1.5">
									<p
										className="truncate font-medium"
										style={nameColor ? { color: nameColor } : undefined}
									>
										{name}
									</p>
									{member.role === "owner" ? (
										<OwnerCrownIcon label={t("invite.role_owner")} />
									) : null}
								</div>
								{member.username ? (
									<p
										className="truncate select-text text-xs text-muted-foreground"
										data-wichat-selectable="true"
									>
										@{member.username}
									</p>
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
				cell: ({ row }) => {
					const { joinMethod, member } = row.original;
					const inviter = member.invited_by_username
						? `@${member.invited_by_username}`
						: member.invited_by_display_name.trim();
					const content = (
						<span className="inline-flex flex-wrap items-center gap-1.5 text-muted-foreground">
							<span>{t(`group_settings.members.join_method.${joinMethod}`)}</span>
							{joinMethod === "invite_link" && member.invite_code ? (
								<code className="rounded bg-muted px-1.5 py-0.5 font-mono text-[11px] text-foreground/85">
									{member.invite_code}
								</code>
							) : null}
						</span>
					);
					if (!inviter) {
						return content;
					}
					return (
						<TooltipProvider delay={200}>
							<Tooltip>
								<TooltipTrigger className="cursor-default text-left">{content}</TooltipTrigger>
								<TooltipContent side="top">
									{t("group_settings.members.join_method.invited_by", { inviter })}
								</TooltipContent>
							</Tooltip>
						</TooltipProvider>
					);
				},
			},
			{
				id: "roles",
				...MEMBERS_COLUMN_SIZE.roles,
				header: () => t("group_settings.members.col_roles"),
				accessorFn: (row) => row.roles,
				cell: ({ row }) => {
					const member = row.original.member;
					const policy = policyByUserId.get(member.user_id);
					const assignableRoleIds = new Set((policy?.assignableRoles ?? []).map((role) => role.id));
					return (
						<MemberRolePickerPopover
							groupId={groupId}
							member={member}
							allRoles={allRoles}
							assignableRoleIds={assignableRoleIds}
							canManageRoles={policy?.canManageRoles ?? false}
							hideRoles={hideRoles}
						/>
					);
				},
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
					const member = row.original.member;
					const policy = policyByUserId.get(member.user_id);
					const name = member.display_name.trim() || t("invite.member_name_fallback");
					return (
						<div className="flex justify-end">
							<button
								type="button"
								className="flex h-8 min-w-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-foreground"
								aria-label={t("group_settings.members.row_actions_aria", { name })}
								onClick={(event) => {
									event.stopPropagation();
									if (!policy) {
										return;
									}
									setMenuState({
										member,
										policy,
										anchor: event.currentTarget,
										align: "end",
									});
								}}
								onContextMenu={(event) => openRowContextMenu(event, member)}
							>
								<EllipsisVertical className="size-4" strokeWidth={2} aria-hidden />
							</button>
						</div>
					);
				},
			},
		);

		return defs;
	}, [
		allRoles,
		enableBulkSelect,
		groupId,
		hideRoles,
		locale,
		openRowContextMenu,
		policyByUserId,
		t,
	]);

	// eslint-disable-next-line react-hooks/incompatible-library -- TanStack Table returns unstable table instance by design.
	const table = useReactTable({
		data: rows,
		columns,
		getCoreRowModel: getCoreRowModel(),
		getRowId: (row) => row.member.user_id,
		enableRowSelection: (row) =>
			enableBulkSelect && (policyByUserId.get(row.original.member.user_id)?.canBulkSelect ?? false),
		onRowSelectionChange,
		state: { rowSelection },
	});

	const showEmpty = !loading && !failed && rows.length === 0;
	const showFailed = failed;
	const showSkeleton = loading && rows.length === 0;

	const leafColumns = table.getVisibleLeafColumns();
	const tableWidthStyle = dataTableMinTableStyle(leafColumns);
	const renderedRows = table.getRowModel().rows;
	const hoistedSections = useMemo(() => {
		if (hideRoles || renderedRows.length === 0) {
			return null;
		}
		const hoistedMap = new Map<
			string,
			{ label: string; position: number; rows: typeof renderedRows }
		>();
		const unhoisted: typeof renderedRows = [];
		for (const row of renderedRows) {
			const hoisted = resolveMemberHoistedRole(row.original.member.roles, hideRoles);
			if (hoisted) {
				const existing = hoistedMap.get(hoisted.id);
				if (existing) {
					existing.rows.push(row);
				} else {
					hoistedMap.set(hoisted.id, {
						label: hoisted.name,
						position: hoisted.position,
						rows: [row],
					});
				}
			} else {
				unhoisted.push(row);
			}
		}
		if (hoistedMap.size === 0) {
			return null;
		}
		const sections = Array.from(hoistedMap.entries())
			.sort(([, a], [, b]) => a.position - b.position)
			.map(([id, entry]) => ({
				key: `role:${id}`,
				label: entry.label,
				rows: entry.rows,
			}));
		if (unhoisted.length > 0) {
			sections.push({
				key: "members",
				label: t("group_settings.nav.members"),
				rows: unhoisted,
			});
		}
		return sections;
	}, [hideRoles, renderedRows, t]);

	const renderDataRow = (row: (typeof renderedRows)[number]) => (
		<tr
			key={row.id}
			className={cn(groupSettingsDataTableRowClass, "group/row cursor-pointer")}
			onClick={(event) => {
				const target = event.target as HTMLElement | null;
				if (target?.closest('button, a, input, label, [role="button"], [role="checkbox"]')) {
					return;
				}
				setProfileMemberId(row.original.member.user_id);
			}}
			onContextMenu={(event) => openRowContextMenu(event, row.original.member)}
		>
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
	);

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
							? hoistedSections
								? hoistedSections.flatMap((section) => [
										<tr key={section.key} className="bg-muted/35">
											<td
												colSpan={columns.length}
												className="px-4 py-1.5 text-[11px] font-semibold uppercase tracking-wide text-muted-foreground"
											>
												{section.label} — {section.rows.length}
											</td>
										</tr>,
										...section.rows.map(renderDataRow),
									])
								: renderedRows.map(renderDataRow)
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

			<MemberActionMenu
				groupId={groupId}
				menuState={menuState}
				relationshipKind={
					menuState ? (relationshipByUserId.get(menuState.member.user_id) ?? null) : null
				}
				onCloseMenu={() => setMenuState(null)}
				onOpenProfile={(member) => setProfileMemberId(member.user_id)}
				onRequestIgnore={(member) => setIgnoreMemberId(member.user_id)}
				onRequestBlock={(member) => setBlockMemberId(member.user_id)}
				onClearRelationship={handleClearRelationship}
				onRequestTimeout={(member) => setTimeoutMemberId(member.user_id)}
				onClearTimeout={handleClearTimeout}
				onRequestKick={onRequestKick}
				onRequestBan={onRequestBan}
			/>

			<MemberProfileDialog
				groupId={groupId}
				member={profileMember}
				policy={profilePolicy}
				relationshipKind={
					profileMember ? (relationshipByUserId.get(profileMember.user_id) ?? null) : null
				}
				allRoles={allRoles}
				hideRoles={hideRoles}
				onClose={() => setProfileMemberId(null)}
				onRequestIgnore={(member) => setIgnoreMemberId(member.user_id)}
				onRequestBlock={(member) => setBlockMemberId(member.user_id)}
				onClearRelationship={handleClearRelationship}
				onRequestTimeout={(member) => setTimeoutMemberId(member.user_id)}
				onClearTimeout={handleClearTimeout}
				onRequestKick={onRequestKick}
				onRequestBan={onRequestBan}
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
		</div>
	);
}
