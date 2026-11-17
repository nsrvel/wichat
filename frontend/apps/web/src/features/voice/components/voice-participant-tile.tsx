"use client";

import { useEffect, useRef, useState } from "react";
import type { Track } from "livekit-client";
import { MicOff, MonitorUp, PhoneOff, Volume2, VolumeX } from "lucide-react";
import { useTranslation } from "react-i18next";
import { UserAvatar } from "@/components/ui/user-avatar";
import { cn } from "@/lib/utils";
import type { VoiceParticipantSnapshot } from "@/features/voice/voice-session-provider";
import { readVoiceSettings, writeVoiceSettings } from "@/features/voice/lib/voice-settings-storage";

interface VoiceParticipantTileProps {
	participant: VoiceParticipantSnapshot;
	channelId: string;
	groupId?: string;
	volumePercent: number;
	locallyMuted: boolean;
	canMuteMembers: boolean;
	canDeafenMembers: boolean;
	canMoveMembers: boolean;
	focused?: boolean;
	onFocus?: () => void;
	onModerate?: (
		targetUserId: string,
		action: "mute" | "unmute" | "deafen" | "undeafen" | "disconnect",
	) => void;
	preferScreenTrack?: boolean;
}

function TrackVideo({ track, mirrored }: { track: Track; mirrored?: boolean }) {
	const videoRef = useRef<HTMLVideoElement | null>(null);

	useEffect(() => {
		const el = videoRef.current;
		if (!el) return;
		track.attach(el);
		return () => {
			track.detach(el);
		};
	}, [track]);

	return (
		<video
			ref={videoRef}
			autoPlay
			playsInline
			muted
			className={cn(
				"h-full w-full rounded-lg bg-black/80 object-contain",
				mirrored && "-scale-x-100",
			)}
		/>
	);
}

export function VoiceParticipantTile({
	participant,
	volumePercent,
	locallyMuted,
	canMuteMembers,
	canDeafenMembers,
	canMoveMembers,
	focused = false,
	onFocus,
	onModerate,
	preferScreenTrack = false,
}: VoiceParticipantTileProps) {
	const { t } = useTranslation();
	const [menuPos, setMenuPos] = useState<{ x: number; y: number } | null>(null);
	const activeTrack =
		preferScreenTrack && participant.screenTrack
			? participant.screenTrack
			: participant.cameraTrack;
	const showVideo = Boolean(activeTrack);

	useEffect(() => {
		if (!menuPos) return;
		const close = () => setMenuPos(null);
		window.addEventListener("click", close);
		window.addEventListener("keydown", close);
		return () => {
			window.removeEventListener("click", close);
			window.removeEventListener("keydown", close);
		};
	}, [menuPos]);

	return (
		<>
			<div
				role="button"
				tabIndex={0}
				onClick={onFocus}
				onContextMenu={(e) => {
					if (participant.isLocal) return;
					e.preventDefault();
					e.stopPropagation();
					setMenuPos({ x: e.clientX, y: e.clientY });
				}}
				onKeyDown={(e) => {
					if (e.key === "Enter" || e.key === " ") {
						e.preventDefault();
						onFocus?.();
					}
				}}
				className={cn(
					"group relative flex aspect-video min-h-[140px] items-center justify-center overflow-hidden rounded-xl border bg-chat-surface-elevated transition-all",
					participant.isSpeaking
						? "border-emerald-400 ring-2 ring-emerald-400/50"
						: "border-chat-surface-border hover:border-border",
					focused && "col-span-full aspect-[16/9] max-h-[520px]",
				)}
			>
				{showVideo && activeTrack ? (
					<TrackVideo
						track={activeTrack}
						mirrored={participant.isLocal && activeTrack === participant.cameraTrack}
					/>
				) : (
					<UserAvatar
						size="xl"
						src={participant.avatarUrl || undefined}
						alt={participant.name}
						initials={participant.name.slice(0, 2).toUpperCase()}
						className={cn(participant.isSpeaking && "ring-2 ring-emerald-400")}
					/>
				)}

				<div className="pointer-events-none absolute inset-x-2 bottom-2 flex items-center justify-between gap-2">
					<span className="inline-flex max-w-[75%] items-center gap-1.5 truncate rounded-md bg-black/65 px-2 py-0.5 text-xs font-medium text-white backdrop-blur-sm">
						<span className="truncate">
							{participant.name}
							{participant.isLocal ? ` (${t("voice.stage.you")})` : ""}
						</span>
					</span>
					<div className="flex items-center gap-1">
						{participant.isScreenSharing ? (
							<span className="rounded bg-emerald-500/80 p-1 text-white">
								<MonitorUp className="size-3" />
							</span>
						) : null}
						{participant.isMuted ? (
							<span className="rounded bg-red-500/85 p-1 text-white">
								<MicOff className="size-3" />
							</span>
						) : null}
					</div>
				</div>
			</div>

			{menuPos && !participant.isLocal ? (
				<div
					role="menu"
					className="fixed z-[90] w-56 rounded-lg border border-border bg-card p-1.5 text-xs shadow-elevated"
					style={{ left: menuPos.x, top: menuPos.y }}
					onClick={(e) => e.stopPropagation()}
				>
					<div className="px-2 py-1.5">
						<div className="mb-1 flex items-center justify-between text-text-muted">
							<span>{t("voice.moderation.user_volume", { percent: volumePercent })}</span>
							<Volume2 className="size-3.5" />
						</div>
						<input
							type="range"
							min={0}
							max={200}
							step={5}
							value={volumePercent}
							onChange={(e) => {
								const next = Number(e.target.value);
								const current = readVoiceSettings().userVolumes;
								writeVoiceSettings({
									userVolumes: { ...current, [participant.identity]: next },
								});
							}}
							className="w-full accent-primary"
						/>
					</div>

					<button
						type="button"
						className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left hover:bg-muted"
						onClick={() => {
							const current = readVoiceSettings().locallyMutedUsers;
							writeVoiceSettings({
								locallyMutedUsers: {
									...current,
									[participant.identity]: !locallyMuted,
								},
							});
							setMenuPos(null);
						}}
					>
						<VolumeX className="size-3.5" />
						<span>{t("voice.moderation.local_mute")}</span>
					</button>

					{(canMuteMembers || canDeafenMembers || canMoveMembers) && onModerate ? (
						<div className="my-1 border-t border-border pt-1">
							{canMuteMembers ? (
								<button
									type="button"
									className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left hover:bg-muted"
									onClick={() => {
										onModerate(participant.identity, "mute");
										setMenuPos(null);
									}}
								>
									<MicOff className="size-3.5" />
									<span>{t("voice.moderation.server_mute")}</span>
								</button>
							) : null}
							{canDeafenMembers ? (
								<button
									type="button"
									className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left hover:bg-muted"
									onClick={() => {
										onModerate(participant.identity, "deafen");
										setMenuPos(null);
									}}
								>
									<VolumeX className="size-3.5" />
									<span>{t("voice.moderation.server_deafen")}</span>
								</button>
							) : null}
							{canMoveMembers ? (
								<button
									type="button"
									className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-red-400 hover:bg-red-500/15"
									onClick={() => {
										onModerate(participant.identity, "disconnect");
										setMenuPos(null);
									}}
								>
									<PhoneOff className="size-3.5" />
									<span>{t("voice.moderation.disconnect_user")}</span>
								</button>
							) : null}
						</div>
					) : null}
				</div>
			) : null}
		</>
	);
}
