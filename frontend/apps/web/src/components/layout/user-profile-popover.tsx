"use client";

import type { User } from "@wichat/contracts/user";
import { Check, ChevronRight, Pencil, Users, Volume2 } from "lucide-react";
import { useState, type ReactElement } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Menu } from "@base-ui/react/menu";
import { WiChatMenuPositioner } from "@/components/ui/menu-positioner";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { UserAvatar, UserProfileIndicator } from "@/components/ui/user-avatar";
import { UserStatusNotePill } from "@/components/user/user-status-note-pill";
import { UserProfileCardShell } from "@/components/user/user-profile-card-shell";
import { cn } from "@/lib/utils";
import {
	USER_PROFILE_AVATAR_OVERLAP_PX,
	USER_PROFILE_AVATAR_RING_CLASS,
	USER_PROFILE_AVATAR_SIZE,
	USER_PROFILE_BIO_CLASS,
	USER_PROFILE_CARD_WIDTH_PX,
	USER_PROFILE_DISPLAY_NAME_CLASS,
	USER_PROFILE_META_CLASS,
} from "@/lib/user-profile-card-layout";
import {
	formatStatusNoteLabel,
	USER_STATUS_BY_ID,
	USER_STATUS_OPTIONS,
	type StatusNote,
	type UserPresenceStatus,
} from "@/lib/user-status";

export type VoiceSessionSummary = {
	channelName: string;
	groupName: string;
	onOpenVoice?: () => void;
};

type UserProfilePopoverProps = {
	user: User | null | undefined;
	displayName: string;
	initials: string;
	status: UserPresenceStatus;
	/** Effective presence dot on avatars; manual `status` stays for the status picker. */
	indicatorStatus?: UserPresenceStatus | "offline";
	onStatusChange: (status: UserPresenceStatus) => void;
	statusNote: StatusNote;
	onOpenSetStatus: () => void;
	onOpenEditProfile?: () => void;
	onClearStatusNote: () => void;
	voiceSession?: VoiceSessionSummary | null;
	widthPx?: number;
	trigger: ReactElement;
};

const cardSectionClass = "rounded-xl bg-muted/45 p-1.5";

const actionRowClass =
	"flex w-full cursor-default select-none items-center gap-2.5 rounded-md px-2.5 py-2 text-left text-[13px] font-medium text-card-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground";

export function UserProfilePopover({
	user,
	displayName,
	initials,
	status,
	indicatorStatus,
	onStatusChange,
	statusNote,
	onOpenSetStatus,
	onOpenEditProfile,
	onClearStatusNote,
	voiceSession,
	widthPx = USER_PROFILE_CARD_WIDTH_PX,
	trigger,
}: UserProfilePopoverProps) {
	const { t } = useTranslation();
	const [open, setOpen] = useState(false);
	const [statusMenuOpen, setStatusMenuOpen] = useState(false);

	const avatarUrl = user?.avatar_url?.trim() || "";
	const username = user?.username?.trim() || "";
	const pronouns = user?.pronouns?.trim() || "";
	const bio = user?.bio?.trim() || "";
	const bannerColor = user?.banner_color?.trim() || "";
	const activeStatus = USER_STATUS_BY_ID[status];
	const avatarStatus = indicatorStatus ?? status;
	const statusNoteLabel = formatStatusNoteLabel(statusNote);

	const handleOpenChange = (nextOpen: boolean) => {
		setOpen(nextOpen);
		if (!nextOpen) {
			setStatusMenuOpen(false);
		}
	};

	const handleEditNoteClick = () => {
		setOpen(false);
		setStatusMenuOpen(false);
		onOpenSetStatus();
	};

	const handleComingSoon = () => {
		toast.info(t("shell.group_rail_menu.coming_soon"));
	};

	return (
		<Popover open={open} onOpenChange={handleOpenChange}>
			<PopoverTrigger render={trigger} />
			<PopoverContent
				side="top"
				align="start"
				sideOffset={14}
				className="-ml-2 w-auto overflow-visible rounded-2xl border-border bg-card p-0 text-card-foreground shadow-elevated"
			>
				<UserProfileCardShell widthPx={widthPx} bannerColor={bannerColor || undefined}>
						<div
							className="flex items-center gap-2.5"
							style={{ marginTop: -USER_PROFILE_AVATAR_OVERLAP_PX }}
						>
							<UserAvatar
								src={avatarUrl}
								initials={initials}
								size={USER_PROFILE_AVATAR_SIZE}
								status={avatarStatus}
								ringClassName={USER_PROFILE_AVATAR_RING_CLASS}
							/>

							<UserStatusNotePill
								label={statusNoteLabel}
								onEdit={handleEditNoteClick}
								onClear={onClearStatusNote}
							/>
						</div>

						{/* Identity */}
						<div className="mt-3.5 min-w-0">
							<p className={USER_PROFILE_DISPLAY_NAME_CLASS}>{displayName}</p>
							{username || pronouns ? (
								<p className={USER_PROFILE_META_CLASS}>
									{username ? `@${username}` : ""}
									{username && pronouns ? " • " : ""}
									{pronouns}
								</p>
							) : null}
							{bio ? <p className={USER_PROFILE_BIO_CLASS}>{bio}</p> : null}
						</div>

						{/* Conditional Voice Card */}
						{voiceSession ? (
							<div className="mt-3 rounded-lg border border-border/70 bg-muted/35 p-2.5">
								<p className="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">
									{t("shell.user_dock.in_voice")}
								</p>
								<div className="mt-1.5 flex items-center gap-2">
									<span className="flex size-8 shrink-0 items-center justify-center rounded-md bg-emerald-500/15 text-emerald-500">
										<Volume2 className="size-4" aria-hidden />
									</span>
									<div className="min-w-0 flex-1">
										<p className="truncate text-xs font-semibold text-foreground">
											{voiceSession.channelName}
										</p>
										<p className="truncate text-[11px] text-muted-foreground">
											{voiceSession.groupName}
										</p>
									</div>
								</div>
								{voiceSession.onOpenVoice ? (
									<button
										type="button"
										onClick={voiceSession.onOpenVoice}
										className="mt-2.5 w-full rounded-md border border-border/70 bg-background/60 px-2.5 py-1.5 text-xs font-medium text-foreground transition-colors hover:bg-sidebar-accent"
									>
										{t("shell.user_dock.open_voice")}
									</button>
								) : null}
							</div>
						) : null}

						{/* Primary Actions Card */}
						<div className={cn("mt-3", cardSectionClass)}>
							<button
								type="button"
								onClick={() => {
									if (onOpenEditProfile) {
										setOpen(false);
										setStatusMenuOpen(false);
										onOpenEditProfile();
										return;
									}
									handleComingSoon();
								}}
								className={actionRowClass}
							>
								<Pencil className="size-4 shrink-0 opacity-75" aria-hidden />
								<span className="flex-1 truncate">{t("shell.user_dock.edit_profile")}</span>
							</button>

							<div className="mx-2 my-1 h-px bg-border/60" />

							<Menu.Root open={statusMenuOpen} onOpenChange={setStatusMenuOpen}>
								<Menu.Trigger
									type="button"
									className={cn(
										actionRowClass,
										"justify-between",
										statusMenuOpen && "bg-sidebar-accent text-sidebar-accent-foreground",
									)}
								>
									<span className="flex min-w-0 items-center gap-2.5">
										<UserProfileIndicator status={status} className="size-3 shrink-0" />
										<span className="truncate">{t(activeStatus.labelKey)}</span>
									</span>
									<ChevronRight
										className="size-3.5 shrink-0 text-muted-foreground/80"
										aria-hidden
									/>
								</Menu.Trigger>
								<Menu.Portal>
									<WiChatMenuPositioner side="right" align="end" sideOffset={8}>
										<Menu.Popup className="w-[248px] rounded-xl border border-border bg-card p-1.5 text-card-foreground shadow-elevated outline-none">
											<div className="space-y-0.5">
												{USER_STATUS_OPTIONS.map((option) => {
													const selected = option.id === status;
													return (
														<Menu.Item
															key={option.id}
															className={cn(
																"flex w-full cursor-default select-none items-start gap-2.5 rounded-md px-2.5 py-2 text-left outline-none",
																"data-highlighted:bg-sidebar-accent data-highlighted:text-sidebar-accent-foreground",
																selected && "bg-sidebar-accent/70",
															)}
															onClick={() => onStatusChange(option.id)}
														>
															<UserProfileIndicator
																status={option.id}
																className="mt-1 size-3 shrink-0"
															/>
															<span className="min-w-0 flex-1">
																<span className="block text-[13px] font-medium leading-snug text-card-foreground">
																	{t(option.labelKey)}
																</span>
																{option.descriptionKey ? (
																	<span className="mt-0.5 block text-[11px] font-normal leading-snug text-muted-foreground">
																		{t(option.descriptionKey)}
																	</span>
																) : null}
															</span>
															{selected ? (
																<Check
																	className="mt-0.5 size-3.5 shrink-0 text-primary"
																	aria-hidden
																/>
															) : null}
														</Menu.Item>
													);
												})}
											</div>
										</Menu.Popup>
									</WiChatMenuPositioner>
								</Menu.Portal>
							</Menu.Root>
						</div>

						{/* Account Actions Card */}
						<div className={cn("mt-2 space-y-0.5", cardSectionClass)}>
							<button
								type="button"
								onClick={handleComingSoon}
								className={cn(actionRowClass, "justify-between")}
							>
								<span className="flex min-w-0 items-center gap-2.5">
									<Users className="size-4 shrink-0 opacity-75" aria-hidden />
									<span className="truncate">{t("shell.user_dock.switch_accounts")}</span>
								</span>
								<ChevronRight className="size-3.5 shrink-0 text-muted-foreground/80" aria-hidden />
							</button>
						</div>
				</UserProfileCardShell>
			</PopoverContent>
		</Popover>
	);
}
