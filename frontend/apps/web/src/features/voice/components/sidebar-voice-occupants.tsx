"use client";

import type { VoiceOccupant } from "@wichat/contracts/chat/voice";
import { HeadphoneOff, MicOff, MonitorUp, Video } from "lucide-react";
import { UserAvatar } from "@/components/ui/user-avatar";
import { cn } from "@/lib/utils";

interface SidebarVoiceOccupantsProps {
	occupants: VoiceOccupant[];
	isRecording?: boolean;
	speakingIds?: Set<string>;
}

export function SidebarVoiceOccupants({
	occupants,
	isRecording = false,
	speakingIds,
}: SidebarVoiceOccupantsProps) {
	if (occupants.length === 0 && !isRecording) {
		return null;
	}

	return (
		<div className="mt-0.5 ml-6 flex flex-col gap-0.5 pr-1">
			{occupants.map((occ) => {
				const speaking = speakingIds?.has(occ.user_id) ?? false;
				const displayName = occ.display_name || occ.username || "?";
				return (
					<div
						key={occ.user_id}
						className="flex h-6 items-center justify-between gap-1.5 rounded-md px-1.5 text-xs text-text-muted"
					>
						<div className="flex min-w-0 items-center gap-1.5">
							<UserAvatar
								size="xs"
								src={occ.avatar_url || undefined}
								alt={displayName}
								initials={displayName.slice(0, 2).toUpperCase()}
								className={cn("size-4 shrink-0", speaking && "ring-2 ring-emerald-400")}
							/>
							<span className="truncate text-[12px] text-text-secondary">{displayName}</span>
						</div>
						<div className="flex shrink-0 items-center gap-1 text-text-muted">
							{occ.is_screen_sharing ? <MonitorUp className="size-3 text-emerald-400" /> : null}
							{occ.is_camera_on ? <Video className="size-3 text-emerald-400" /> : null}
							{occ.is_deafened || occ.is_server_deafened ? (
								<HeadphoneOff className="size-3 text-red-400" />
							) : occ.is_muted || occ.is_server_muted ? (
								<MicOff className="size-3 text-red-400" />
							) : null}
						</div>
					</div>
				);
			})}
		</div>
	);
}
