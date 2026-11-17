"use client";

import {
	HeadphoneOff,
	Headphones,
	Mic,
	MicOff,
	MonitorUp,
	PhoneOff,
	Radio,
	Video,
	VideoOff,
} from "lucide-react";
import Link from "next/link";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { channelPath } from "@/lib/app-routes";
import { cn } from "@/lib/utils";
import { useVoiceSession } from "@/features/voice/voice-session-provider";

export function VoiceCallControls() {
	const { t } = useTranslation();
	const {
		phase,
		activeChannel,
		isMuted,
		isDeafened,
		isServerMuted,
		isServerDeafened,
		isCameraOn,
		isScreenSharing,
		isRecording,
		canPublishAudio,
		canPublishVideo,
		leaveChannel,
		toggleMute,
		toggleDeafen,
		toggleCamera,
		toggleScreenShare,
	} = useVoiceSession();

	if (phase === "disconnected" || !activeChannel) {
		return null;
	}

	const statusLabel =
		phase === "connecting"
			? t("voice.status.connecting")
			: phase === "reconnecting"
				? t("voice.status.reconnecting")
				: t("voice.status.connected");

	const statusColor =
		phase === "connected"
			? "text-emerald-400"
			: phase === "reconnecting"
				? "text-amber-400"
				: "text-text-muted";

	const channelHref = activeChannel.groupId
		? channelPath(activeChannel.groupId, activeChannel.channelId)
		: `/channels/${activeChannel.channelId}`;

	return (
		<div
			className="pointer-events-auto mb-1.5 w-full rounded-xl border border-chat-surface-border bg-chat-surface-elevated p-2.5 shadow-[var(--shadow-elevated)]"
			data-testid="voice-call-controls"
		>
			<div className="mb-2 flex items-center justify-between gap-2">
				<div className="min-w-0">
					<div className="flex items-center gap-1.5">
						<span className={cn("text-xs font-semibold", statusColor)}>{statusLabel}</span>
						{isRecording ? (
							<span className="inline-flex items-center gap-1 rounded-full bg-red-500/15 px-1.5 py-0.5 text-[10px] font-medium text-red-400">
								<Radio className="size-2.5 animate-pulse" />
								{t("voice.recording.badge")}
							</span>
						) : null}
					</div>
					<Link
						href={channelHref}
						className="block truncate text-xs text-text-muted hover:text-text-primary hover:underline"
					>
						{activeChannel.channelName}
					</Link>
				</div>
				<Button
					type="button"
					variant="ghost"
					size="icon"
					className="size-7 shrink-0 text-red-400 hover:bg-red-500/15 hover:text-red-300"
					aria-label={t("voice.controls.disconnect")}
					onClick={() => void leaveChannel()}
				>
					<PhoneOff className="size-4" />
				</Button>
			</div>

			<div className="grid grid-cols-4 gap-1">
				<Button
					type="button"
					variant="ghost"
					size="sm"
					disabled={isServerMuted || !canPublishAudio}
					className={cn(
						"h-8 px-0",
						(isMuted || isServerMuted) && "bg-red-500/15 text-red-400 hover:bg-red-500/20",
					)}
					aria-label={
						isServerMuted
							? t("voice.controls.server_muted")
							: isMuted
								? t("voice.controls.unmute")
								: t("voice.controls.mute")
					}
					onClick={() => void toggleMute()}
				>
					{isMuted || isServerMuted ? <MicOff className="size-4" /> : <Mic className="size-4" />}
				</Button>
				<Button
					type="button"
					variant="ghost"
					size="sm"
					disabled={isServerDeafened}
					className={cn(
						"h-8 px-0",
						(isDeafened || isServerDeafened) && "bg-red-500/15 text-red-400 hover:bg-red-500/20",
					)}
					aria-label={
						isServerDeafened
							? t("voice.controls.server_deafened")
							: isDeafened
								? t("voice.controls.undeafen")
								: t("voice.controls.deafen")
					}
					onClick={() => void toggleDeafen()}
				>
					{isDeafened || isServerDeafened ? (
						<HeadphoneOff className="size-4" />
					) : (
						<Headphones className="size-4" />
					)}
				</Button>
				<Button
					type="button"
					variant="ghost"
					size="sm"
					disabled={!canPublishVideo}
					className={cn(
						"h-8 px-0",
						isCameraOn && "bg-emerald-500/15 text-emerald-400 hover:bg-emerald-500/20",
					)}
					aria-label={isCameraOn ? t("voice.controls.camera_off") : t("voice.controls.camera_on")}
					onClick={() => void toggleCamera()}
				>
					{isCameraOn ? <Video className="size-4" /> : <VideoOff className="size-4" />}
				</Button>
				<Button
					type="button"
					variant="ghost"
					size="sm"
					disabled={!canPublishVideo}
					className={cn(
						"h-8 px-0",
						isScreenSharing && "bg-emerald-500/15 text-emerald-400 hover:bg-emerald-500/20",
					)}
					aria-label={
						isScreenSharing ? t("voice.controls.stop_share") : t("voice.controls.share_screen")
					}
					onClick={() => void toggleScreenShare()}
				>
					<MonitorUp className="size-4" />
				</Button>
			</div>
		</div>
	);
}
