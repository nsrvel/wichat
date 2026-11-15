"use client";

import { useQuery } from "@tanstack/react-query";
import { isAppLocale } from "@wichat/i18n/locale";
import { Headphones, Mic, MicOff, Settings } from "lucide-react";
import { useParams } from "next/navigation";
import { useEffect, useMemo, useState, useSyncExternalStore } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { SetStatusDialog } from "@/components/layout/set-status-dialog";
import { ICON_RAIL_WIDTH_PX, useShellChrome } from "@/components/layout/shell-chrome-context";
import { usePresenceSync } from "@/hooks/use-presence-sync";
import { CHANNEL_SIDEBAR_MIN_WIDTH } from "@/hooks/use-resizable-sidebar-width";
import { UserProfileButton } from "@/components/layout/user-profile-button";
import { UserProfilePopover } from "@/components/layout/user-profile-popover";
import {
	WorkspaceHeaderIconButton,
	workspaceHeaderIconClass,
} from "@/components/layout/workspace-header-icon-button";
import { WorkspaceSplitIconControl } from "@/components/layout/workspace-split-icon-control";
import { UserSettingsDialog } from "@/features/user-settings/components/user-settings-dialog";
import type { UserSettingsSectionId } from "@/features/user-settings/lib/user-settings-nav";
import { applyAppLocale } from "@/integrations/i18n/apply-app-locale";
import { translateError } from "@/integrations/i18n/translate-error";
import { readStoredTimeFormat, writeStoredTimeFormat } from "@/lib/time-format";
import {
	EMPTY_STATUS_NOTE,
	formatStatusNoteLabel,
	readStoredUserStatus,
	statusNoteFromUser,
	subscribeUserPresence,
	writeStoredUserStatus,
	type StatusNote,
	type UserPresenceStatus,
} from "@/lib/user-status";
import { cn } from "@/lib/utils";
import { sessionQueryOpts } from "@/services/auth/auth.query";
import {
	useUpdateMyPresenceMutation,
	useUpdateMyStatusNoteMutation,
} from "@/services/user/user.mutation";
import { myPresenceQueryOptions } from "@/services/user/user.query";

const dockSurfaceClass = "bg-chat-surface/85 backdrop-blur-md text-foreground";

function initialsForName(name: string): string {
	const parts = name.trim().split(/\s+/).filter(Boolean);
	if (parts.length === 0) return "W";
	if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
	return `${parts[0][0] ?? ""}${parts[1][0] ?? ""}`.toUpperCase();
}

export function ShellUserDock() {
	const { t, i18n } = useTranslation();
	const params = useParams<{ groupId?: string }>();
	const activeGroupId = typeof params?.groupId === "string" ? params.groupId : undefined;
	const { railOpen, sidebarWidth } = useShellChrome();
	const { data: user } = useQuery(sessionQueryOpts());
	usePresenceSync(activeGroupId, Boolean(user));
	const myPresenceQuery = useQuery({
		...myPresenceQueryOptions(),
		enabled: Boolean(user),
	});
	const updatePresence = useUpdateMyPresenceMutation();
	const updateStatusNote = useUpdateMyStatusNoteMutation();
	const [muted, setMuted] = useState(false);
	const [deafened, setDeafened] = useState(false);
	const [statusDialogOpen, setStatusDialogOpen] = useState(false);
	const [settingsOpen, setSettingsOpen] = useState(false);
	const [settingsTab, setSettingsTab] = useState<UserSettingsSectionId>("account");
	const [expiryTick, setExpiryTick] = useState(0);
	const status = useSyncExternalStore(
		subscribeUserPresence,
		readStoredUserStatus,
		() => "online" as const,
	);
	const statusNote = useMemo(
		() => statusNoteFromUser(user),
		// eslint-disable-next-line react-hooks/exhaustive-deps
		[user?.status_note_text, user?.status_note_emoji, user?.status_note_expires_at, expiryTick],
	);
	const statusNoteLabel = formatStatusNoteLabel(statusNote);

	useEffect(() => {
		if (!statusNote.expiresAt) return;
		const delayMs = Math.max(0, statusNote.expiresAt - Date.now());
		const timer = window.setTimeout(() => {
			setExpiryTick((prev) => prev + 1);
		}, Math.min(delayMs, 2_147_483_647));
		return () => window.clearTimeout(timer);
	}, [statusNote.expiresAt]);

	useEffect(() => {
		if (!user) return;
		if (user.locale && isAppLocale(user.locale)) {
			const current = (i18n.resolvedLanguage ?? i18n.language ?? "").split("-")[0];
			if (current !== user.locale) {
				void applyAppLocale(user.locale, i18n);
			}
		}
		if (user.time_format && user.time_format !== readStoredTimeFormat()) {
			writeStoredTimeFormat(user.time_format);
		}
	}, [user, i18n]);

	useEffect(() => {
		const remoteManual = myPresenceQuery.data?.manual_status;
		if (remoteManual && remoteManual !== readStoredUserStatus()) {
			writeStoredUserStatus(remoteManual);
		}
	}, [myPresenceQuery.data?.manual_status]);

	const handleStatusChange = (nextStatus: UserPresenceStatus) => {
		writeStoredUserStatus(nextStatus);
		updatePresence.mutate({ status: nextStatus });
	};

	const handleStatusNoteSave = async (nextNote: StatusNote) => {
		await updateStatusNote.mutateAsync({
			text: nextNote.text,
			emoji: nextNote.emoji,
			expires_at:
				nextNote.expiresAt !== null ? new Date(nextNote.expiresAt).toISOString() : null,
		});
	};

	const handleClearStatusNote = () => {
		void handleStatusNoteSave(EMPTY_STATUS_NOTE).catch((err) => {
			toast.error(translateError(t, err));
		});
	};

	const dockWidth = (railOpen ? ICON_RAIL_WIDTH_PX : 0) + sidebarWidth;
	const popoverWidth = ICON_RAIL_WIDTH_PX + CHANNEL_SIDEBAR_MIN_WIDTH - 16;
	const displayName = user?.display_name?.trim() || user?.username?.trim() || "WiChat User";
	const initials = initialsForName(displayName);
	const secondaryText =
		statusNoteLabel ||
		(user?.username ? `@${user.username}` : t("shell.user_dock.status_online"));
	const avatarUrl = user?.avatar_url?.trim() || "";

	return (
		<>
			<div
				className="pointer-events-none fixed bottom-0 left-0 z-50 box-border hidden p-2 transition-[width] duration-200 ease-out lg:block"
				style={{ width: dockWidth }}
			>
			<div
				className={cn(
					"pointer-events-auto w-full min-w-0 overflow-hidden rounded-xl border border-chat-surface-border shadow-[var(--shadow-elevated)]",
					dockSurfaceClass,
				)}
			>
				<div className="flex h-[52px] min-w-0 items-center gap-1 px-2">
					<UserProfilePopover
						user={user}
						displayName={displayName}
						initials={initials}
						status={status}
						onStatusChange={handleStatusChange}
						statusNote={statusNote}
						widthPx={popoverWidth}
						onOpenSetStatus={() => setStatusDialogOpen(true)}
						onOpenEditProfile={() => {
							setSettingsTab("profiles");
							setSettingsOpen(true);
						}}
						onClearStatusNote={handleClearStatusNote}
						trigger={
							<UserProfileButton
								displayName={displayName}
								secondaryText={secondaryText}
								avatarUrl={avatarUrl}
								initials={initials}
								status={status}
								statusRingClassName="ring-chat-surface"
								aria-label={t("shell.user_dock.open_profile")}
							/>
						}
					/>
					<div className="flex shrink-0 items-center gap-0.5">
						<WorkspaceSplitIconControl
							mainAriaLabel={muted ? "Unmute microphone" : "Mute microphone"}
							optionsAriaLabel="Microphone options"
							mainActive={muted}
							onMainClick={() => setMuted((prev) => !prev)}
						>
							{muted ? (
								<MicOff className={workspaceHeaderIconClass} strokeWidth={2} />
							) : (
								<Mic className={workspaceHeaderIconClass} strokeWidth={2} />
							)}
						</WorkspaceSplitIconControl>
						<WorkspaceSplitIconControl
							mainAriaLabel={deafened ? "Undeafen" : "Deafen"}
							optionsAriaLabel="Headphone options"
							mainActive={deafened}
							onMainClick={() => setDeafened((prev) => !prev)}
						>
							<Headphones className={workspaceHeaderIconClass} strokeWidth={2} />
						</WorkspaceSplitIconControl>
						<WorkspaceHeaderIconButton
							aria-label="User settings"
							onClick={() => {
								setSettingsTab("account");
								setSettingsOpen(true);
							}}
						>
							<Settings className={workspaceHeaderIconClass} strokeWidth={2} />
						</WorkspaceHeaderIconButton>
					</div>
				</div>
			</div>
		</div>
		<SetStatusDialog
				open={statusDialogOpen}
				onOpenChange={setStatusDialogOpen}
				displayName={displayName}
				username={user?.username?.trim() || ""}
				avatarUrl={avatarUrl}
				bannerColor={user?.banner_color?.trim() || ""}
				initials={initials}
				status={status}
				initialNote={statusNote}
				onSave={handleStatusNoteSave}
			/>
			<UserSettingsDialog
				open={settingsOpen}
				onOpenChange={setSettingsOpen}
				initialSection={settingsTab}
				user={user}
				displayName={displayName}
				secondaryText={secondaryText}
				avatarUrl={avatarUrl}
				initials={initials}
				status={status}
				statusNoteLabel={statusNoteLabel}
			/>
		</>
	);
}

