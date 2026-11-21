"use client";

import type { Group } from "@wichat/contracts/group";
import type { GroupMember } from "@wichat/contracts/group/member";
import type { GroupRole } from "@wichat/contracts/group/role";
import type { UserRelationshipKind } from "@wichat/contracts/user";
import { Menu } from "@base-ui/react/menu";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ChevronRight, Fingerprint, MessageCircle, MoreHorizontal, Users } from "lucide-react";
import { useId, useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { WiStudioLogoMark } from "@/components/brand/wistudio-logo-mark";
import { Button } from "@/components/ui/button";
import { OwnerCrownIcon } from "@/components/ui/owner-crown-icon";
import { UserAvatar } from "@/components/ui/user-avatar";
import { WiChatDialog, WiChatDialogBody } from "@/components/ui/wichat-dialog";
import { UserStatusNotePill } from "@/components/user/user-status-note-pill";
import { MemberRolePickerPopover } from "@/features/group-settings/components/members/member-role-picker-popover";
import { formatShortDate } from "@/features/group-settings/lib/group-settings-members-format";
import type { MemberModerationPolicy } from "@/features/group-settings/lib/member-moderation-policy";
import { hasGroupPermission } from "@/lib/group-permissions";
import {
	menuChevronClass,
	menuDestructiveItemClass,
	menuItemClass,
	menuPopupClass,
	menuPositionerClass,
	menuSeparatorClass,
	submenuTriggerClass,
} from "@/lib/menu-chrome";
import { userInitials } from "@/lib/user-initials";
import {
	formatStatusNoteLabel,
	statusNoteFromUser,
	type UserPresenceStatus,
} from "@/lib/user-status";
import { cn } from "@/lib/utils";
import { myGroupsQueryOpts } from "@/services/group/group.query";
import { Q_INVITE_MEMBERS } from "@/services/invite/invite.query";
import { TurnstileWidget } from "@/features/auth/components/turnstile-widget";
import { useInviteTurnstile } from "@/features/invite/hooks/use-invite-turnstile";
import { sendGroupInvites } from "@/services/invite/invite.service";

type MemberProfileDialogProps = {
	groupId: string;
	groupName?: string;
	member: GroupMember | null;
	policy: MemberModerationPolicy | null;
	relationshipKind?: UserRelationshipKind | null;
	allRoles: readonly GroupRole[];
	hideRoles?: boolean;
	onClose: () => void;
	onRequestIgnore: (member: GroupMember) => void;
	onRequestBlock: (member: GroupMember) => void;
	onClearRelationship: (member: GroupMember, previousKind: UserRelationshipKind) => void;
	onRequestTimeout: (member: GroupMember) => void;
	onClearTimeout: (member: GroupMember) => void;
	onRequestKick: (member: GroupMember) => void;
	onRequestBan: (member: GroupMember) => void;
};

/**
 * Discord-style Member Profile Card opened from the members table context menu.
 */
export function MemberProfileDialog({
	groupId,
	groupName,
	member,
	policy,
	relationshipKind = null,
	allRoles,
	hideRoles = false,
	onClose,
	onRequestIgnore,
	onRequestBlock,
	onClearRelationship,
	onRequestTimeout,
	onClearTimeout,
	onRequestKick,
	onRequestBan,
}: MemberProfileDialogProps) {
	const { t, i18n } = useTranslation();
	const queryClient = useQueryClient();
	const titleId = useId();
	const [nowMs] = useState(() => Date.now());
	const locale = i18n.language || "en";

	const { data: myGroups = [] } = useQuery({
		...myGroupsQueryOpts(),
		enabled: member !== null && !policy?.isSelf,
	});

	const invitableGroups = myGroups.filter(
		(g) => g.id !== groupId && hasGroupPermission(g, "create_invite"),
	);
	const captcha = useInviteTurnstile(member !== null && invitableGroups.length > 0);

	const memberName = member?.display_name.trim() || t("invite.member_name_fallback");

	const inviteToGroupMutation = useMutation({
		mutationFn: async (targetGroup: Group) => {
			if (!member) {
				throw new Error("Missing member");
			}
			return sendGroupInvites(targetGroup.id, {
				user_ids: [member.user_id],
				emails: [],
				turnstile_token: captcha.turnstileField,
			});
		},
		onSuccess: (result, targetGroup) => {
			void queryClient.invalidateQueries({ queryKey: [Q_INVITE_MEMBERS, targetGroup.id] });
			if (result.added > 0) {
				toast.success(
					t("group_settings.members.profile_dialog.invite_to_group_success", {
						name: memberName,
						group: targetGroup.name,
					}),
				);
				return;
			}
			toast.info(
				t("group_settings.members.profile_dialog.invite_to_group_already_member", {
					name: memberName,
					group: targetGroup.name,
				}),
			);
		},
		onError: (_error, targetGroup) => {
			toast.error(
				t("group_settings.members.profile_dialog.invite_to_group_error", {
					name: memberName,
					group: targetGroup.name,
				}),
			);
		},
	});
	const bannerColor = member?.banner_color?.trim() || "";
	const isHexBanner = /^#[0-9a-f]{3,8}$/i.test(bannerColor);
	const presenceStatus: UserPresenceStatus | "offline" = member?.is_online
		? (member.presence_status as UserPresenceStatus) || "online"
		: "offline";
	const statusNote = statusNoteFromUser(member);
	const statusNoteLabel = formatStatusNoteLabel(statusNote);
	const assignableRoleIds = new Set((policy?.assignableRoles ?? []).map((role) => role.id));

	const copyUsername = async () => {
		if (!member?.username) {
			return;
		}
		try {
			await navigator.clipboard.writeText(`@${member.username}`);
			toast.success(
				t("group_settings.members.context_menu.copied_username", {
					username: member.username,
				}),
			);
		} catch {
			toast.error(t("common.error_generic"));
		}
	};

	return (
		<WiChatDialog
			open={member !== null}
			onOpenChange={(open) => {
				if (!open) {
					onClose();
				}
			}}
			titleId={titleId}
			elevated
			className="min-h-[min(440px,90dvh)] max-w-[400px] rounded-2xl border-border bg-card p-0 py-0 text-card-foreground shadow-elevated sm:py-0 md:py-0 max-h-[880px]:py-0 max-h-[880px]:sm:py-0 max-h-[880px]:md:py-0"
		>
			{member ? (
				<WiChatDialogBody insetX={false} className="pt-0 pb-7">
					{/* WiChat profile banner: ~half above avatar, ~half behind top half of avatar */}
					<div
						className="h-[136px] w-full shrink-0 overflow-hidden bg-popover"
						style={isHexBanner ? { backgroundColor: bannerColor } : undefined}
					/>
					<div className="px-6">
						{/* Large avatar (center aligns with banner bottom) + status note */}
						<div className="-mt-[44px] flex items-center gap-2.5">
							<UserAvatar
								src={member.avatar_url}
								initials={userInitials(memberName)}
								size="2xl"
								status={presenceStatus}
								ringClassName="ring-[6px] ring-card"
							/>
							{statusNoteLabel ? <UserStatusNotePill label={statusNoteLabel} /> : null}
						</div>

						{/* Identity */}
						<div className="mt-3.5 min-w-0">
							<div className="flex items-center gap-1.5">
								<h2
									id={titleId}
									className="truncate text-[17px] font-bold leading-tight tracking-tight text-foreground"
								>
									{memberName}
								</h2>
								{member.role === "owner" ? <OwnerCrownIcon label={t("invite.role_owner")} /> : null}
							</div>

							{member.username || member.pronouns ? (
								<p className="mt-0.5 truncate text-[13px] font-medium text-muted-foreground">
									{member.username ? `@${member.username}` : ""}
									{member.username && member.pronouns ? " • " : ""}
									{member.pronouns ?? ""}
								</p>
							) : null}
						</div>

						{!policy?.isSelf ? (
							<div className="mt-3.5 flex items-center gap-2">
								<Button
									type="button"
									size="sm"
									className="h-8 gap-1.5 rounded-lg px-3.5 text-xs font-semibold"
									onClick={() =>
										toast.info(t("group_settings.members.context_menu.message_coming_soon"))
									}
								>
									<MessageCircle className="size-3.5" strokeWidth={2} aria-hidden />
									<span>{t("group_settings.members.context_menu.message")}</span>
								</Button>

								<Menu.Root modal={false}>
									<Menu.Trigger
										aria-label={t("group_settings.members.profile_dialog.more_actions")}
										className="inline-flex size-8 items-center justify-center rounded-lg border border-border/80 bg-muted/50 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
									>
										<MoreHorizontal className="size-4" strokeWidth={2} aria-hidden />
									</Menu.Trigger>
									<Menu.Portal>
										<Menu.Positioner
											className={menuPositionerClass}
											side="bottom"
											align="start"
											sideOffset={4}
										>
											<Menu.Popup className={menuPopupClass}>
												{invitableGroups.length > 0 ? (
													<>
														<Menu.SubmenuRoot>
															<Menu.SubmenuTrigger className={submenuTriggerClass}>
																<span>
																	{t("group_settings.members.profile_dialog.invite_to_group")}
																</span>
																<ChevronRight className={menuChevronClass} aria-hidden />
															</Menu.SubmenuTrigger>
															<Menu.Portal>
																<Menu.Positioner
																	className={menuPositionerClass}
																	side="right"
																	align="start"
																	sideOffset={4}
																>
																	<Menu.Popup
																		className={cn(
																			menuPopupClass,
																			"thin-scroll max-h-60 overflow-y-auto",
																		)}
																	>
																		{invitableGroups.map((targetGroup) => (
																			<Menu.Item
																				key={targetGroup.id}
																				className={menuItemClass}
																				disabled={
																					inviteToGroupMutation.isPending || !captcha.captchaReady
																				}
																				onClick={() => inviteToGroupMutation.mutate(targetGroup)}
																			>
																				<UserAvatar
																					src={targetGroup.logo_url}
																					initials={userInitials(targetGroup.name)}
																					size="xs"
																				/>
																				<span className="truncate">{targetGroup.name}</span>
																			</Menu.Item>
																		))}
																	</Menu.Popup>
																</Menu.Positioner>
															</Menu.Portal>
														</Menu.SubmenuRoot>
														<Menu.Separator className={menuSeparatorClass} />
													</>
												) : null}
												<Menu.Item
													className={menuItemClass}
													onClick={() => {
														onClose();
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
														onClose();
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
												{policy?.canTimeout || policy?.canKick || policy?.canBan ? (
													<>
														<Menu.Separator className={menuSeparatorClass} />
														{policy.canTimeout ? (
															<Menu.Item
																className={menuDestructiveItemClass}
																onClick={() => {
																	onClose();
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
																	member.timeout_until &&
																		new Date(member.timeout_until).getTime() > nowMs
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
																	onClose();
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
																	onClose();
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
															onClick={() => void copyUsername()}
														>
															<Fingerprint className="size-3.5 shrink-0 opacity-45" aria-hidden />
															<span>{t("group_settings.members.context_menu.copy_username")}</span>
														</Menu.Item>
													</>
												) : null}
											</Menu.Popup>
										</Menu.Positioner>
									</Menu.Portal>
								</Menu.Root>
							</div>
						) : null}

						<div className="mt-6 space-y-5">
							{member.bio.trim() ? (
								<div>
									<p className="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">
										{t("group_settings.members.profile_dialog.about_me")}
									</p>
									<p className="mt-1 whitespace-pre-wrap break-words text-xs leading-relaxed text-foreground">
										{member.bio.trim()}
									</p>
								</div>
							) : null}

							<div>
								<p className="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">
									{t("group_settings.members.profile_dialog.member_since")}
								</p>
								<div className="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1 text-xs font-medium text-foreground">
									{member.user_created_at ? (
										<>
											<span className="inline-flex items-center gap-1.5" title="WiChat">
												<WiStudioLogoMark size={16} className="rounded-[4px]" />
												{formatShortDate(member.user_created_at, locale)}
											</span>
											<span className="text-muted-foreground" aria-hidden>
												·
											</span>
										</>
									) : null}
									<span className="inline-flex items-center gap-1.5" title={groupName}>
										<Users className="size-3.5 text-muted-foreground" strokeWidth={2} aria-hidden />
										{formatShortDate(member.joined_at, locale)}
									</span>
								</div>
							</div>

							{!hideRoles ? (
								<div>
									<p className="mb-1.5 text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">
										{t("group_settings.members.profile_dialog.roles")}
									</p>
									<MemberRolePickerPopover
										groupId={groupId}
										member={member}
										allRoles={allRoles}
										assignableRoleIds={assignableRoleIds}
										canManageRoles={policy?.canManageRoles ?? false}
										hideRoles={hideRoles}
										alwaysShowAdd
									/>
								</div>
							) : null}

							{invitableGroups.length > 0 && captcha.siteKey ? (
								<TurnstileWidget
									siteKey={captcha.siteKey}
									resetKey={captcha.resetKey}
									onToken={captcha.setToken}
								/>
							) : null}
						</div>
					</div>
				</WiChatDialogBody>
			) : null}
		</WiChatDialog>
	);
}
