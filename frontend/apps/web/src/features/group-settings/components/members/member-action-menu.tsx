"use client";
import { WiChatMenuPositioner } from "@/components/ui/menu-positioner";

import { ApiError } from "@wichat/api-client/api-error";
import type { GroupMember } from "@wichat/contracts/group/member";
import type { GroupRole } from "@wichat/contracts/group/role";
import type { UserRelationshipKind } from "@wichat/contracts/user";
import { Menu } from "@base-ui/react/menu";
import { useQueryClient } from "@tanstack/react-query";
import { Check, ChevronRight, Fingerprint } from "lucide-react";
import { useLayoutEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import type { MemberModerationPolicy } from "@/features/group-settings/lib/member-moderation-policy";
import { hasCustomRoleColor } from "@/lib/member-role-display";
import {
	menuCheckboxBoxClass,
	menuCheckboxItemClass,
	menuChevronClass,
	menuDestructiveItemClass,
	menuItemClass,
	menuPopupClass,
	menuPositionerClass,
	menuSeparatorClass,
	submenuTriggerClass,
} from "@/lib/menu-chrome";
import { cn } from "@/lib/utils";
import { groupRoleKeys } from "@/services/group-role/group-role.key";
import {
	addGroupRoleMembers,
	removeGroupRoleMember,
} from "@/services/group-role/group-role.service";
import { Q_INVITE_MEMBERS } from "@/services/invite/invite.query";

export type MemberActionMenuState = {
	member: GroupMember;
	policy: MemberModerationPolicy;
	anchor: HTMLElement | { getBoundingClientRect: () => DOMRect };
	align: "start" | "end";
};

export function createMemberContextMenuState(
	member: GroupMember,
	policy: MemberModerationPolicy,
	clientX: number,
	clientY: number,
): MemberActionMenuState {
	return {
		member,
		policy,
		anchor: {
			getBoundingClientRect: () => new DOMRect(clientX, clientY, 0, 0),
		},
		align: "start",
	};
}

type MemberActionMenuProps = {
	groupId: string;
	menuState: MemberActionMenuState | null;
	relationshipKind?: UserRelationshipKind | null;
	onCloseMenu: () => void;
	onOpenProfile: (member: GroupMember) => void;
	onRequestIgnore: (member: GroupMember) => void;
	onRequestBlock: (member: GroupMember) => void;
	onClearRelationship: (member: GroupMember, previousKind: UserRelationshipKind) => void;
	onRequestTimeout: (member: GroupMember) => void;
	onClearTimeout: (member: GroupMember) => void;
	onRequestKick: (member: GroupMember) => void;
	onRequestBan: (member: GroupMember) => void;
};

/**
 * Discord-style floating action/context menu for a member row, anchored either to
 * the right-click cursor position or the row's `EllipsisVertical` button.
 */
export function MemberActionMenu({
	groupId,
	menuState,
	relationshipKind = null,
	onCloseMenu,
	onOpenProfile,
	onRequestIgnore,
	onRequestBlock,
	onClearRelationship,
	onRequestTimeout,
	onClearTimeout,
	onRequestKick,
	onRequestBan,
}: MemberActionMenuProps) {
	const { t } = useTranslation();
	const queryClient = useQueryClient();
	const [pendingRoleId, setPendingRoleId] = useState<string | null>(null);
	const [rolesSubmenuOpen, setRolesSubmenuOpen] = useState(false);
	const [nowMs] = useState(() => Date.now());
	const [submenuSide, setSubmenuSide] = useState<"right" | "left">("right");
	const rolesTriggerWrapRef = useRef<HTMLDivElement>(null);

	useLayoutEffect(() => {
		if (!rolesSubmenuOpen || !rolesTriggerWrapRef.current) {
			return;
		}
		const rect = rolesTriggerWrapRef.current.getBoundingClientRect();
		setSubmenuSide(rect.right + 240 > window.innerWidth ? "left" : "right");
	}, [rolesSubmenuOpen]);

	const member = menuState?.member ?? null;
	const policy = menuState?.policy ?? null;

	const memberRoleIds = useMemo(
		() => new Set((member?.roles ?? []).map((r) => r.id)),
		[member?.roles],
	);

	const memberName = member?.display_name.trim() || t("invite.member_name_fallback");
	const hasRolesSubmenu = Boolean(policy?.canManageRoles && policy.assignableRoles.length > 0);
	const hasModerationSection = Boolean(policy?.canTimeout || policy?.canKick || policy?.canBan);

	const copyText = async (value: string, toastMessage: string) => {
		onCloseMenu();
		try {
			await navigator.clipboard.writeText(value);
			toast.success(toastMessage);
		} catch {
			toast.error(t("common.error_generic"));
		}
	};

	const handleToggleRole = async (role: GroupRole) => {
		if (!member || pendingRoleId) {
			return;
		}
		const assigned = memberRoleIds.has(role.id);
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
		<Menu.Root
			modal={false}
			open={menuState !== null}
			onOpenChange={(open) => {
				if (!open) {
					setRolesSubmenuOpen(false);
					onCloseMenu();
				}
			}}
		>
			<Menu.Portal>
				<WiChatMenuPositioner
					anchor={menuState?.anchor ?? null}
					className={menuPositionerClass}
					side="bottom"
					align={menuState?.align ?? "end"}
					sideOffset={4}
				>
					<Menu.Popup className={menuPopupClass}>
						{menuState && member && policy ? (
							<>
								<Menu.Item
									className={menuItemClass}
									onClick={() => {
										onCloseMenu();
										onOpenProfile(member);
									}}
								>
									{t("group_settings.members.context_menu.profile")}
								</Menu.Item>
								{!policy.isSelf ? (
									<Menu.Item
										className={menuItemClass}
										onClick={() => {
											onCloseMenu();
											toast.info(t("group_settings.members.context_menu.message_coming_soon"));
										}}
									>
										{t("group_settings.members.context_menu.message")}
									</Menu.Item>
								) : null}

								{!policy.isSelf ? (
									<>
										<Menu.Separator className={menuSeparatorClass} />
										<Menu.Item
											className={menuItemClass}
											onClick={() => {
												onCloseMenu();
												if (relationshipKind === "ignore") {
													onClearRelationship(member, "ignore");
												} else {
													onRequestIgnore(member);
												}
											}}
										>
											{t(
												relationshipKind === "ignore"
													? "group_settings.members.context_menu.unignore"
													: "group_settings.members.context_menu.ignore",
											)}
										</Menu.Item>
										<Menu.Item
											className={menuDestructiveItemClass}
											onClick={() => {
												onCloseMenu();
												if (relationshipKind === "block") {
													onClearRelationship(member, "block");
												} else {
													onRequestBlock(member);
												}
											}}
										>
											{t(
												relationshipKind === "block"
													? "group_settings.members.context_menu.unblock"
													: "group_settings.members.context_menu.block",
											)}
										</Menu.Item>
									</>
								) : null}

								{hasRolesSubmenu ? (
									<>
										<Menu.Separator className={menuSeparatorClass} />
										<div
											ref={rolesTriggerWrapRef}
											className="relative"
											onMouseEnter={() => setRolesSubmenuOpen(true)}
											onMouseLeave={() => setRolesSubmenuOpen(false)}
										>
											<button
												type="button"
												aria-haspopup="menu"
												aria-expanded={rolesSubmenuOpen}
												className={cn(
													submenuTriggerClass,
													rolesSubmenuOpen && "bg-sidebar-accent text-sidebar-accent-foreground",
												)}
												onClick={(event) => {
													event.preventDefault();
													event.stopPropagation();
													setRolesSubmenuOpen((prev) => !prev);
												}}
											>
												<span>{t("group_settings.members.context_menu.roles")}</span>
												<ChevronRight className={menuChevronClass} strokeWidth={2} aria-hidden />
											</button>

											{rolesSubmenuOpen ? (
												<div
													className={cn(
														"absolute top-0 z-[110]",
														submenuSide === "right" ? "left-full pl-1.5" : "right-full pr-1.5",
													)}
												>
													<div
														role="menu"
														className={cn(
															menuPopupClass,
															"w-fit min-w-[9rem] max-w-[16rem] max-h-64 overflow-y-auto thin-scroll",
														)}
													>
														{policy.assignableRoles.map((role) => {
															const checked = memberRoleIds.has(role.id);
															const dotColor = hasCustomRoleColor(role.color)
																? role.color
																: "#64748b";
															const isPending = pendingRoleId === role.id;
															return (
																<button
																	key={role.id}
																	type="button"
																	role="menuitemcheckbox"
																	aria-checked={checked}
																	disabled={isPending}
																	onClick={(event) => {
																		event.stopPropagation();
																		void handleToggleRole(role);
																	}}
																	className={cn(
																		menuCheckboxItemClass,
																		"group/checkbox hover:bg-sidebar-accent hover:text-sidebar-accent-foreground disabled:pointer-events-none disabled:opacity-55",
																	)}
																	data-checked={checked ? "" : undefined}
																>
																	<span className="flex min-w-0 flex-1 items-center gap-2 pr-1">
																		<span
																			className="size-2.5 shrink-0 rounded-full"
																			style={{ backgroundColor: dotColor }}
																			aria-hidden
																		/>
																		<span className="truncate">{role.name}</span>
																	</span>
																	<span
																		className={cn(
																			menuCheckboxBoxClass,
																			checked &&
																				"border-primary bg-primary text-primary-foreground",
																		)}
																		aria-hidden
																	>
																		{checked ? (
																			<Check className="size-3" strokeWidth={2.5} />
																		) : null}
																	</span>
																</button>
															);
														})}
													</div>
												</div>
											) : null}
										</div>
									</>
								) : null}

								{hasModerationSection ? (
									<>
										<Menu.Separator className={menuSeparatorClass} />
										{policy.canTimeout ? (
											<Menu.Item
												className={menuDestructiveItemClass}
												onClick={() => {
													onCloseMenu();
													if (
														member.timeout_until &&
														new Date(member.timeout_until).getTime() > nowMs
													) {
														onClearTimeout(member);
													} else {
														onRequestTimeout(member);
													}
												}}
											>
												{t(
													member.timeout_until && new Date(member.timeout_until).getTime() > nowMs
														? "group_settings.members.context_menu.remove_timeout_member"
														: "group_settings.members.context_menu.timeout_member",
													{
														name: memberName,
													},
												)}
											</Menu.Item>
										) : null}
										{policy.canKick ? (
											<Menu.Item
												className={menuDestructiveItemClass}
												onClick={() => {
													onCloseMenu();
													onRequestKick(member);
												}}
											>
												{t("group_settings.members.context_menu.kick_member", {
													name: memberName,
												})}
											</Menu.Item>
										) : null}
										{policy.canBan ? (
											<Menu.Item
												className={menuDestructiveItemClass}
												onClick={() => {
													onCloseMenu();
													onRequestBan(member);
												}}
											>
												{t("group_settings.members.context_menu.ban_member", {
													name: memberName,
												})}
											</Menu.Item>
										) : null}
									</>
								) : null}

								{member.username ? (
									<>
										<Menu.Separator className={menuSeparatorClass} />
										<Menu.Item
											className={menuItemClass}
											onClick={() =>
												void copyText(
													`@${member.username}`,
													t("group_settings.members.context_menu.copied_username", {
														username: member.username,
													}),
												)
											}
										>
											<Fingerprint className="size-3.5 shrink-0 opacity-45" aria-hidden />
											<span>{t("group_settings.members.context_menu.copy_username")}</span>
										</Menu.Item>
									</>
								) : null}
							</>
						) : null}
					</Menu.Popup>
				</WiChatMenuPositioner>
			</Menu.Portal>
		</Menu.Root>
	);
}
