"use client";

import { useQuery } from "@tanstack/react-query";
import { Headphones, Mic, MicOff, Settings } from "lucide-react";
import { useState } from "react";
import { ICON_RAIL_WIDTH_PX, useShellChrome } from "@/components/layout/shell-chrome-context";
import {
	WorkspaceHeaderIconButton,
	workspaceHeaderIconClass,
} from "@/components/layout/workspace-header-icon-button";
import { WorkspaceSplitIconControl } from "@/components/layout/workspace-split-icon-control";
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
	const { railOpen, sidebarWidth } = useShellChrome();
	const { data: user } = useQuery(sessionQueryOpts());
	const [muted, setMuted] = useState(false);
	const [deafened, setDeafened] = useState(false);

	const dockWidth = (railOpen ? ICON_RAIL_WIDTH_PX : 0) + sidebarWidth;
	const displayName = user?.display_name?.trim() || user?.username?.trim() || "WiChat User";
	const secondaryText = user?.username ? `@${user.username}` : "Online";
	const avatarUrl = user?.avatar_url?.trim() || "";

	return (
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
					<button
						type="button"
						className="flex min-h-9 min-w-0 flex-1 items-center gap-2 rounded-lg px-1.5 py-1 text-left transition-colors hover:bg-sidebar-accent"
						aria-label="Open profile"
					>
						<div className="relative shrink-0">
							{avatarUrl ? (
								/* eslint-disable-next-line @next/next/no-img-element */
								<img
									src={avatarUrl}
									alt=""
									width={32}
									height={32}
									className="h-8 w-8 rounded-full object-cover"
								/>
							) : (
								<span className="flex h-8 w-8 items-center justify-center rounded-full bg-primary/15 text-xs font-semibold text-primary">
									{initialsForName(displayName)}
								</span>
							)}
							<span
								className="absolute -bottom-0.5 -right-0.5 h-3 w-3 rounded-full border-2 border-chat-surface bg-emerald-500"
								aria-hidden
							/>
						</div>
						<div className="min-w-0 flex-1">
							<p className="truncate text-[13px] font-semibold leading-tight text-foreground">
								{displayName}
							</p>
							<p className="truncate text-[11px] leading-tight text-muted-foreground">
								{secondaryText}
							</p>
						</div>
					</button>
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
						<WorkspaceHeaderIconButton aria-label="User settings">
							<Settings className={workspaceHeaderIconClass} strokeWidth={2} />
						</WorkspaceHeaderIconButton>
					</div>
				</div>
			</div>
		</div>
	);
}

