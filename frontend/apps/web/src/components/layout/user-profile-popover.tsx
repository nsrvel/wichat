"use client";

import type { User } from "@wichat/contracts";
import {
	Check,
	ChevronRight,
	Pencil,
	Plus,
	Users,
	Volume2,
	X,
} from "lucide-react";
import { useState, type ReactElement } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { UserAvatar, UserProfileIndicator } from "@/components/ui/user-avatar";
import { cn } from "@/lib/utils";
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
	onStatusChange: (status: UserPresenceStatus) => void;
	statusNote: StatusNote;
	onOpenSetStatus: () => void;
	onOpenEditProfile?: () => void;
	onClearStatusNote: () => void;
	voiceSession?: VoiceSessionSummary | null;
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
	onStatusChange,
	statusNote,
	onOpenSetStatus,
	onOpenEditProfile,
	onClearStatusNote,
	voiceSession,
	trigger,
}: UserProfilePopoverProps) {
	const { t } = useTranslation();
	const [open, setOpen] = useState(false);
	const [statusMenuOpen, setStatusMenuOpen] = useState(false);

	const avatarUrl = user?.avatar_url?.trim() || "";
	const username = user?.username?.trim() || "";
	const activeStatus = USER_STATUS_BY_ID[status];
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
				sideOffset={10}
				className="w-[320px] overflow-visible rounded-2xl border-border bg-card p-0 text-card-foreground shadow-elevated"
			>
				{/* WiChat profile banner: ~half above avatar, ~half behind top half of avatar */}
				<div className="h-[92px] overflow-hidden rounded-t-2xl bg-popover" />

				<div className="px-5 pb-5">
					{/* Large avatar (center aligns with banner bottom) + status note */}
					<div className="-mt-[44px] flex items-center gap-2.5">
						<UserAvatar
							src={avatarUrl}
							initials={initials}
							size="2xl"
							status={status}
							ringClassName="ring-[6px] ring-card"
						/>

						<div className="min-w-0 flex-1">
							<div className="inline-flex w-fit max-w-full items-center gap-1.5 rounded-2xl bg-muted px-3.5 py-2.5 text-xs text-muted-foreground shadow-[0_6px_16px_rgba(0,0,0,0.28)] transition-colors hover:bg-muted/85 hover:text-foreground">
								<button
									type="button"
									onClick={handleEditNoteClick}
									className="flex min-w-0 items-center gap-1.5 text-left"
								>
									{!statusNoteLabel ? (
										<Plus className="size-3.5 shrink-0 opacity-80" aria-hidden />
									) : null}
									<span className="line-clamp-2 break-words leading-snug">
										{statusNoteLabel || t("shell.user_dock.status_note_placeholder")}
									</span>
								</button>
								{statusNoteLabel ? (
									<button
										type="button"
										onClick={onClearStatusNote}
										aria-label={t("shell.user_dock.status_note_clear")}
										className="shrink-0 rounded-full p-0.5 text-muted-foreground/80 transition-colors hover:bg-foreground/10 hover:text-foreground"
									>
										<X className="size-3.5" aria-hidden />
									</button>
								) : null}
							</div>
						</div>
					</div>

					{/* Identity */}
					<div className="mt-3.5 min-w-0">
						<p className="truncate text-[17px] font-bold leading-tight tracking-tight text-foreground">
							{displayName}
						</p>
						{username ? (
							<p className="mt-0.5 truncate text-[13px] font-medium text-muted-foreground">
								@{username}
							</p>
						) : null}
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

						<div className="relative">
							<button
								type="button"
								aria-expanded={statusMenuOpen}
								onClick={() => setStatusMenuOpen((prev) => !prev)}
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
							</button>

							{statusMenuOpen ? (
								<div
									role="menu"
									className="absolute bottom-0 left-full z-10 ml-2 w-[248px] rounded-xl border border-border bg-card p-1.5 text-card-foreground shadow-elevated"
								>
									<div className="space-y-0.5">
										{USER_STATUS_OPTIONS.map((option) => {
											const selected = option.id === status;
											return (
												<button
													key={option.id}
													type="button"
													role="menuitemradio"
													aria-checked={selected}
													onClick={() => {
														onStatusChange(option.id);
														setStatusMenuOpen(false);
													}}
													className={cn(
														"flex w-full cursor-default select-none items-start gap-2.5 rounded-md px-2.5 py-2 text-left hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
														selected && "bg-sidebar-accent/70",
													)}
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
												</button>
											);
										})}
									</div>
								</div>
							) : null}
						</div>
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
							<ChevronRight
								className="size-3.5 shrink-0 text-muted-foreground/80"
								aria-hidden
							/>
						</button>
					</div>
				</div>
			</PopoverContent>
		</Popover>
	);
}
