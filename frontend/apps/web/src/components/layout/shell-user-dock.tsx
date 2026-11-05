"use client";

import {
	Headphones,
	Megaphone,
	MicOff,
	MonitorUp,
	PhoneOff,
	Wifi,
	Settings,
	Shapes,
	VideoOff,
	Volume2,
} from "lucide-react";
import Image from "next/image";
import type { ComponentProps } from "react";
import { ICON_RAIL_WIDTH_PX, useShellChrome } from "@/components/layout/shell-chrome-context";
import { shellUserDockAnchorClass } from "@/components/layout/shell-rail/shell-rail-aside";
import {
	WorkspaceHeaderIconButton,
	workspaceHeaderIconClass,
} from "@/components/layout/workspace-header-icon-button";
import { WorkspaceSplitIconControl } from "@/components/layout/workspace-split-icon-control";
import { cn } from "@/lib/utils";

/** Outer `p-2` on fixed dock wrapper (each side). */
export const SHELL_USER_DOCK_OUTER_PADDING_PX = 16;

/**
 * Min width of the rounded dock card (profile + controls readable).
 * When the icon rail is hidden, sidebar width is raised to fit this + outer padding.
 */
export const SHELL_USER_DOCK_MIN_INNER_WIDTH_PX = 300;

export function sidebarMinWidthForUserDock(railOpen: boolean, baseMin = 280) {
	if (railOpen) return baseMin;
	return Math.max(baseMin, SHELL_USER_DOCK_MIN_INNER_WIDTH_PX + SHELL_USER_DOCK_OUTER_PADDING_PX);
}

const dockSurfaceClass = "bg-chat-surface text-foreground";
const dockControlSurfaceClass = "bg-chat-composer-field transition-colors hover:bg-sidebar-accent";

function DockIconButton({
	className,
	children,
	...props
}: ComponentProps<typeof WorkspaceHeaderIconButton>) {
	return (
		<WorkspaceHeaderIconButton
			className={cn("min-w-0 flex-1", dockControlSurfaceClass, className)}
			{...props}
		>
			{children}
		</WorkspaceHeaderIconButton>
	);
}

function VoiceConnectedPanel({ onDisconnect }: { onDisconnect: () => void }) {
	return (
		<div className={cn(dockSurfaceClass, "border-b border-chat-surface-border px-3 py-2.5")}>
			<div className="flex items-start gap-2">
				<span
					className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg text-[#16a34a]"
					aria-hidden
				>
					<Wifi className={workspaceHeaderIconClass} strokeWidth={2} />
				</span>
				<div className="min-w-0 flex-1">
					<p className="text-[13px] font-semibold leading-tight text-[#16a34a]">Voice Connected</p>
					<p className="truncate text-[11px] text-[#71717a]">General / WiChat server</p>
				</div>
				<WorkspaceHeaderIconButton
					className="shrink-0"
					aria-label="Disconnect voice"
					onClick={onDisconnect}
				>
					<PhoneOff className={workspaceHeaderIconClass} strokeWidth={2} />
				</WorkspaceHeaderIconButton>
			</div>
			<div className="mt-2 flex gap-1.5">
				<DockIconButton aria-label="Turn off camera">
					<VideoOff className={workspaceHeaderIconClass} strokeWidth={2} />
				</DockIconButton>
				<DockIconButton aria-label="Share screen">
					<MonitorUp className={workspaceHeaderIconClass} strokeWidth={2} />
				</DockIconButton>
				<DockIconButton aria-label="Activities">
					<Shapes className={workspaceHeaderIconClass} strokeWidth={2} />
				</DockIconButton>
				<DockIconButton aria-label="Soundboard">
					<Megaphone className={workspaceHeaderIconClass} strokeWidth={2} />
				</DockIconButton>
			</div>
		</div>
	);
}

function UserProfileBar({ inVoice }: { inVoice: boolean }) {
	return (
		<div className={cn(dockSurfaceClass, "flex h-[52px] min-w-0 items-center gap-1 px-2")}>
			<button
				type="button"
				className="flex min-h-9 min-w-[7.5rem] flex-1 items-center gap-2 rounded-lg px-1.5 py-1 text-left transition-colors hover:bg-sidebar-accent"
				aria-label="Open profile"
			>
				<div className="relative shrink-0">
					<Image
						src="/avatars/demo/marvell.png"
						alt=""
						width={32}
						height={32}
						className="h-8 w-8 rounded-full object-cover"
					/>
					<span
						className="absolute -bottom-0.5 -right-0.5 h-3.5 w-3.5 rounded-full border-2 border-chat-surface bg-[#ef4444]"
						aria-hidden
					/>
				</div>
				<div className="min-w-0 flex-1">
					<p className="truncate text-[13px] font-semibold text-foreground">ⱮąɾѵҽӀӀ</p>
					{inVoice ? (
						<p className="flex items-center gap-1 truncate text-[11px] text-muted-foreground">
							<Volume2 className="h-3 w-3 shrink-0 text-[#16a34a]" strokeWidth={2.2} />
							In voice
						</p>
					) : (
						<p className="truncate text-[11px] text-muted-foreground">Online</p>
					)}
				</div>
			</button>
			<div className="flex shrink-0 items-center gap-1">
				<WorkspaceSplitIconControl
					mainAriaLabel="Unmute microphone"
					optionsAriaLabel="Microphone options"
					mainActive
				>
					<MicOff className={workspaceHeaderIconClass} strokeWidth={2} />
				</WorkspaceSplitIconControl>
				<WorkspaceSplitIconControl mainAriaLabel="Deafen" optionsAriaLabel="Headphone options">
					<Headphones className={workspaceHeaderIconClass} strokeWidth={2} />
				</WorkspaceSplitIconControl>
				<WorkspaceHeaderIconButton aria-label="User settings">
					<Settings className={workspaceHeaderIconClass} strokeWidth={2} />
				</WorkspaceHeaderIconButton>
			</div>
		</div>
	);
}

export function ShellUserDock() {
	const { railOpen, sidebarWidth, voiceConnected, setVoiceConnected } = useShellChrome();
	const dockWidth = (railOpen ? ICON_RAIL_WIDTH_PX : 0) + sidebarWidth;

	return (
		<div
			className={cn(
				shellUserDockAnchorClass(),
				"pointer-events-none fixed bottom-0 left-0 z-50 box-border hidden p-2 lg:block",
			)}
			style={{ width: dockWidth }}
		>
			<div
				className={cn(
					"pointer-events-auto min-w-0 overflow-hidden rounded-xl border border-chat-surface-border shadow-[var(--shadow-elevated)]",
					dockSurfaceClass,
				)}
				style={{ minWidth: SHELL_USER_DOCK_MIN_INNER_WIDTH_PX }}
			>
				{voiceConnected ? (
					<VoiceConnectedPanel onDisconnect={() => setVoiceConnected(false)} />
				) : null}
				<UserProfileBar inVoice={voiceConnected} />
			</div>
		</div>
	);
}
