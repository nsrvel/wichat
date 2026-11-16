"use client";

import { ApiError } from "@wichat/api-client/api-error";
import type { GroupMember, GroupMemberRoleSummary } from "@wichat/contracts/group/member";
import type { GroupRole } from "@wichat/contracts/group/role";
import { useQueryClient } from "@tanstack/react-query";
import { Check, Plus, Search } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { GroupSettingsMemberRoleBadge } from "@/features/group-settings/components/members/group-settings-member-role-badge";
import { hasCustomRoleColor } from "@/lib/member-role-display";
import { cn } from "@/lib/utils";
import { groupRoleKeys } from "@/services/group-role/group-role.key";
import {
	addGroupRoleMembers,
	removeGroupRoleMember,
} from "@/services/group-role/group-role.service";
import { Q_INVITE_MEMBERS } from "@/services/invite/invite.query";

type MemberRolePickerPopoverProps = {
	groupId: string;
	member: GroupMember;
	allRoles: readonly GroupRole[];
	assignableRoleIds: ReadonlySet<string>;
	canManageRoles: boolean;
	hideRoles: boolean;
	alwaysShowAdd?: boolean;
};

/**
 * Renders a member's roles inside a table cell with adaptive width overflow (`+N`),
 * a hover-revealed `+` button for managers, and a shared role-picker popover with
 * colored dots and checkboxes.
 */
export function MemberRolePickerPopover({
	groupId,
	member,
	allRoles,
	assignableRoleIds,
	canManageRoles,
	hideRoles,
	alwaysShowAdd = false,
}: MemberRolePickerPopoverProps) {
	const { t } = useTranslation();
	const queryClient = useQueryClient();
	const containerRef = useRef<HTMLDivElement>(null);
	const [open, setOpen] = useState(false);
	const [search, setSearch] = useState("");
	const [pendingRoleId, setPendingRoleId] = useState<string | null>(null);
	const [maxVisible, setMaxVisible] = useState(1);

	const isOwner = member.role === "owner";
	const memberCustomRoles: readonly GroupMemberRoleSummary[] = useMemo(
		() => (member.roles ?? []).filter((r) => r.slug !== "everyone"),
		[member.roles],
	);
	const memberRoleIds = useMemo(
		() => new Set(memberCustomRoles.map((r) => r.id)),
		[memberCustomRoles],
	);

	const customGroupRoles = useMemo(
		() =>
			allRoles
				.filter((r) => !r.is_default && r.slug !== "everyone")
				.sort((a, b) => a.position - b.position),
		[allRoles],
	);

	useEffect(() => {
		const el = containerRef.current;
		if (!el) {
			return;
		}
		const updateCapacity = (width: number) => {
			if (width >= 260) {
				setMaxVisible(isOwner ? 2 : 3);
			} else if (width >= 185) {
				setMaxVisible(isOwner ? 1 : 2);
			} else {
				setMaxVisible(isOwner ? 0 : 1);
			}
		};
		updateCapacity(el.clientWidth);
		if (typeof ResizeObserver === "undefined") {
			return;
		}
		const observer = new ResizeObserver((entries) => {
			const entry = entries[0];
			if (entry) {
				updateCapacity(entry.contentRect.width);
			}
		});
		observer.observe(el);
		return () => observer.disconnect();
	}, [isOwner]);

	const visibleCustomRoles = memberCustomRoles.slice(0, maxVisible);
	const hiddenCount = Math.max(0, memberCustomRoles.length - visibleCustomRoles.length);

	const filteredRoles = useMemo(() => {
		const q = search.trim().toLowerCase();
		const source = canManageRoles
			? customGroupRoles
			: customGroupRoles.filter((r) => memberRoleIds.has(r.id));
		if (!q) {
			return source;
		}
		return source.filter((r) => r.name.toLowerCase().includes(q));
	}, [canManageRoles, customGroupRoles, memberRoleIds, search]);

	const showAddTrigger = canManageRoles && assignableRoleIds.size > 0;
	const showTrigger = hiddenCount > 0 || showAddTrigger;

	const handleToggleRole = async (role: GroupRole) => {
		if (!canManageRoles || !assignableRoleIds.has(role.id) || pendingRoleId) {
			return;
		}
		const assigned = memberRoleIds.has(role.id);
		const memberName = member.display_name.trim() || t("invite.member_name_fallback");
		setPendingRoleId(role.id);
		try {
			if (assigned) {
				await removeGroupRoleMember(groupId, role.id, member.user_id);
				toast.success(
					t("group_settings.members.roles_popover.removed_toast", {
						role: role.name,
						name: memberName,
					}),
				);
			} else {
				await addGroupRoleMembers(groupId, role.id, { user_ids: [member.user_id] });
				toast.success(
					t("group_settings.members.roles_popover.added_toast", {
						role: role.name,
						name: memberName,
					}),
				);
			}
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: [Q_INVITE_MEMBERS, groupId] }),
				queryClient.invalidateQueries({ queryKey: groupRoleKeys.list(groupId) }),
				queryClient.invalidateQueries({ queryKey: groupRoleKeys.members(groupId, role.id) }),
			]);
		} catch (error) {
			toast.error(error instanceof ApiError ? error.message : t("common.error_generic"));
		} finally {
			setPendingRoleId(null);
		}
	};

	return (
		<div ref={containerRef} className="flex min-w-0 items-center gap-1">
			{isOwner ? <GroupSettingsMemberRoleBadge key="owner" role="owner" /> : null}
			{visibleCustomRoles.map((r) => (
				<GroupSettingsMemberRoleBadge
					key={r.id}
					role={r.name}
					color={r.color}
					hideColor={hideRoles}
				/>
			))}

			{!isOwner && memberCustomRoles.length === 0 && !showAddTrigger ? (
				<span className="text-muted-foreground">—</span>
			) : null}

			{showTrigger ? (
				<Popover
					open={open}
					onOpenChange={(next) => {
						setOpen(next);
						if (!next) {
							setSearch("");
						}
					}}
				>
					<div className="flex shrink-0 items-center gap-1">
						{hiddenCount > 0 ? (
							<PopoverTrigger
								render={
									<button
										type="button"
										onClick={(event) => event.stopPropagation()}
										aria-label={t("group_settings.members.roles_popover.more_roles_aria", {
											count: hiddenCount,
										})}
										className={cn(
											"inline-flex h-6 shrink-0 items-center justify-center rounded-full border border-border/80 bg-muted/65 px-2 text-[11px] font-semibold leading-none text-foreground/90 transition-colors",
											"hover:border-border hover:bg-sidebar-accent",
											open && "border-border bg-sidebar-accent",
										)}
									>
										+{hiddenCount}
									</button>
								}
							/>
						) : null}

						{showAddTrigger ? (
							<PopoverTrigger
								render={
									<button
										type="button"
										onClick={(event) => event.stopPropagation()}
										aria-label={t("group_settings.members.roles_popover.add_role_aria")}
										className={cn(
											"inline-flex size-6 shrink-0 items-center justify-center rounded-full border border-border/75 bg-muted/45 text-muted-foreground transition-all",
											"hover:border-border hover:bg-sidebar-accent hover:text-foreground",
											!alwaysShowAdd &&
												"opacity-0 group-hover/row:opacity-100 focus-visible:opacity-100",
											open && "border-border bg-sidebar-accent text-foreground opacity-100",
										)}
									>
										<Plus className="size-3.5" strokeWidth={2} aria-hidden />
									</button>
								}
							/>
						) : null}
					</div>

					<PopoverContent
						side="bottom"
						align="start"
						sideOffset={6}
						className={cn(
							"w-fit min-w-[9rem] max-w-[16rem] p-1.5",
							customGroupRoles.length > 6 && "min-w-[13rem]",
						)}
					>
						{customGroupRoles.length > 6 ? (
							<div className="relative mb-1.5">
								<Search
									className="pointer-events-none absolute top-1/2 left-2.5 size-3.5 -translate-y-1/2 text-muted-foreground"
									aria-hidden
								/>
								<input
									type="text"
									value={search}
									onChange={(e) => setSearch(e.target.value)}
									placeholder={t("group_settings.members.roles_popover.search_placeholder")}
									className="h-8 w-full rounded-md border border-border/80 bg-field/60 pr-2.5 pl-8 text-xs text-foreground placeholder:text-muted-foreground focus:border-ring focus:outline-none"
								/>
							</div>
						) : null}

						<div className="thin-scroll max-h-56 overflow-y-auto py-0.5">
							{filteredRoles.length === 0 ? (
								<p className="px-2.5 py-3 text-center text-xs text-muted-foreground">
									{t("group_settings.members.roles_popover.empty")}
								</p>
							) : (
								filteredRoles.map((role) => {
									const checked = memberRoleIds.has(role.id);
									const editable = canManageRoles && assignableRoleIds.has(role.id);
									const isPending = pendingRoleId === role.id;
									const dotColor = hasCustomRoleColor(role.color) ? role.color : "#64748b";

									return (
										<button
											key={role.id}
											type="button"
											disabled={!editable || isPending}
											onClick={() => void handleToggleRole(role)}
											className={cn(
												"flex w-full items-center justify-between gap-2.5 rounded-md px-2.5 py-1.5 text-left text-[13px] font-medium transition-colors",
												"hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
												"disabled:cursor-not-allowed disabled:opacity-55",
											)}
										>
											<span className="flex min-w-0 flex-1 items-center gap-2">
												<span
													className="size-2.5 shrink-0 rounded-full"
													style={{ backgroundColor: dotColor }}
													aria-hidden
												/>
												<span className="truncate">{role.name}</span>
											</span>
											{canManageRoles ? (
												<span
													className={cn(
														"flex size-4 shrink-0 items-center justify-center rounded-[4px] border transition-colors",
														checked
															? "border-primary bg-primary text-primary-foreground"
															: "border-border bg-muted/40",
													)}
													aria-hidden
												>
													{checked ? <Check className="size-3" strokeWidth={2.5} /> : null}
												</span>
											) : null}
										</button>
									);
								})
							)}
						</div>
					</PopoverContent>
				</Popover>
			) : null}
		</div>
	);
}
