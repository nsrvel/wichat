"use client";

import { useQuery } from "@tanstack/react-query";
import { Headphones, Mic, MicOff, Settings } from "lucide-react";
import { useMemo, useState, useSyncExternalStore } from "react";
import { useTranslation } from "react-i18next";
import { SetStatusDialog } from "@/components/layout/set-status-dialog";
import { ICON_RAIL_WIDTH_PX, useShellChrome } from "@/components/layout/shell-chrome-context";
import { UserProfileButton } from "@/components/layout/user-profile-button";
import { UserProfilePopover } from "@/components/layout/user-profile-popover";
import {
	WorkspaceHeaderIconButton,
	workspaceHeaderIconClass,
} from "@/components/layout/workspace-header-icon-button";
import { WorkspaceSplitIconControl } from "@/components/layout/workspace-split-icon-control";
import { UserSettingsDialog } from "@/features/user-settings/components/user-settings-dialog";
import type { UserSettingsSectionId } from "@/features/user-settings/lib/user-settings-nav";
import {
	EMPTY_STATUS_NOTE,
	formatStatusNoteLabel,
	parseStatusNoteSnapshot,
	readStoredStatusNoteSnapshot,
	readStoredUserStatus,
	subscribeUserPresence,
	writeStoredStatusNote,
	writeStoredUserStatus,
	type StatusNote,
	type UserPresenceStatus,
} from "@/lib/user-status";
import { cn } from "@/lib/utils";
import { sessionQueryOpts } from "@/services/auth/auth.query";

const dockSurfaceClass = "bg-chat-surface/85 backdrop-blur-md text-foreground";

function initialsForName(name: string): string {
	const parts = name.trim().split(/\s+/).filter(Boolean);
	if (parts.length === 0) return "W";
	if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
	return `${parts[0][0] ?? ""}${parts[1][0] ?? ""}`.toUpperCase();
}

export function ShellUserDock() {
	const { t } = useTranslation();
	const { railOpen, sidebarWidth } = useShellChrome();
	const { data: user } = useQuery(sessionQueryOpts());
	const [muted, setMuted] = useState(false);
	const [deafened, setDeafened] = useState(false);
	const [statusDialogOpen, setStatusDialogOpen] = useState(false);
	const [settingsOpen, setSettingsOpen] = useState(false);
	const [settingsTab, setSettingsTab] = useState<UserSettingsSectionId>("account");
	const status = useSyncExternalStore(
		subscribeUserPresence,
		readStoredUserStatus,
		() => "online" as const,
	);
	const noteSnapshot = useSyncExternalStore(
		subscribeUserPresence,
		readStoredStatusNoteSnapshot,
		() => "",
	);
	const statusNote = useMemo(() => parseStatusNoteSnapshot(noteSnapshot), [noteSnapshot]);
	const statusNoteLabel = formatStatusNoteLabel(statusNote);

	const handleStatusChange = (nextStatus: UserPresenceStatus) => {
		writeStoredUserStatus(nextStatus);
	};

	const handleStatusNoteSave = (nextNote: StatusNote) => {
		writeStoredStatusNote(nextNote);
	};

	const dockWidth = (railOpen ? ICON_RAIL_WIDTH_PX : 0) + sidebarWidth;
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
						onOpenSetStatus={() => setStatusDialogOpen(true)}
						onOpenEditProfile={() => {
							setSettingsTab("profiles");
							setSettingsOpen(true);
						}}
						onClearStatusNote={() => handleStatusNoteSave(EMPTY_STATUS_NOTE)}
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
			/>
		</>
	);
}

