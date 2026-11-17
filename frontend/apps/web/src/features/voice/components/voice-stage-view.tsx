"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import {
	AlertTriangle,
	Download,
	LayoutGrid,
	Maximize2,
	Phone,
	PhoneOff,
	Radio,
	Square,
} from "lucide-react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { translateError } from "@/integrations/i18n/translate-error";
import { cn } from "@/lib/utils";
import { channelRecordingsQueryOpts, voiceQuotaQueryOpts } from "@/services/voice/voice.query";
import {
	useModerateVoiceParticipantMutation,
	useStartRecordingMutation,
	useStopRecordingMutation,
} from "@/services/voice/voice.mutation";
import { useVoiceSession } from "@/features/voice/voice-session-provider";
import { VoiceParticipantTile } from "@/features/voice/components/voice-participant-tile";
import { VoiceRecordingConsentDialog } from "@/features/voice/components/voice-recording-consent-dialog";

interface VoiceStageViewProps {
	channelId: string;
	channelName: string;
	groupId?: string;
	groupName?: string;
	canRecordCall?: boolean;
	canMuteMembers?: boolean;
	canDeafenMembers?: boolean;
	canMoveMembers?: boolean;
	isAdmin?: boolean;
	compact?: boolean;
}

export function VoiceStageView({
	channelId,
	channelName,
	groupId,
	groupName,
	canRecordCall = false,
	canMuteMembers = false,
	canDeafenMembers = false,
	canMoveMembers = false,
	isAdmin = false,
	compact = false,
}: VoiceStageViewProps) {
	const { t } = useTranslation();
	const {
		phase,
		activeChannel,
		participants,
		isRecording,
		isScreenSharing,
		idleWarningSeconds,
		recordingConsentOpen,
		settings,
		joinChannel,
		leaveChannel,
		acknowledgeIdle,
		acceptRecordingConsent,
		declineRecordingConsent,
	} = useVoiceSession();

	const isThisChannelActive = activeChannel?.channelId === channelId && phase !== "disconnected";

	const [layoutMode, setLayoutMode] = useState<"grid" | "focus">("grid");
	const [focusedIdentity, setFocusedIdentity] = useState<string | null>(null);
	const [showRecordings, setShowRecordings] = useState(false);

	const moderateMutation = useModerateVoiceParticipantMutation(channelId, groupId);
	const startRecMutation = useStartRecordingMutation(channelId, groupId);
	const stopRecMutation = useStopRecordingMutation(channelId, groupId);

	const recordingsQuery = useQuery({
		...channelRecordingsQueryOpts(channelId, groupId),
		enabled: showRecordings,
	});
	const quotaQuery = useQuery({
		...voiceQuotaQueryOpts(groupId),
		enabled: isAdmin,
	});

	const screenSharer = participants.find((p) => p.isScreenSharing);
	const effectiveFocusedId = focusedIdentity ?? (screenSharer ? screenSharer.identity : null);
	const effectiveLayout = layoutMode === "focus" || Boolean(screenSharer) ? "focus" : "grid";

	const handleModerate = (
		targetUserId: string,
		action: "mute" | "unmute" | "deafen" | "undeafen" | "disconnect",
	) => {
		moderateMutation.mutate(
			{ group_id: groupId, target_user_id: targetUserId, action },
			{
				onError: (err) => toast.error(translateError(t, err)),
			},
		);
	};

	const handleToggleRecording = () => {
		if (isRecording) {
			stopRecMutation.mutate(undefined, {
				onError: (err) => toast.error(translateError(t, err)),
			});
		} else {
			startRecMutation.mutate(undefined, {
				onError: (err) => toast.error(translateError(t, err)),
			});
		}
	};

	return (
		<div
			className={cn(
				"flex flex-col bg-chat-surface",
				compact ? "border-b border-chat-surface-border p-3" : "min-h-0 flex-1 p-4",
			)}
		>
			{quotaQuery.data?.warning_active ? (
				<div className="mb-3 flex items-center gap-2 rounded-lg border border-amber-500/30 bg-amber-500/10 px-3 py-2 text-xs text-amber-300">
					<AlertTriangle className="size-4 shrink-0" />
					<span>
						{t("voice.quota.warning_banner", {
							percent: Math.round(quotaQuery.data.usage_percent),
							used: quotaQuery.data.used_participant_minutes,
							limit: quotaQuery.data.limit_participant_minutes,
						})}
					</span>
				</div>
			) : null}

			{isThisChannelActive && idleWarningSeconds !== null ? (
				<div className="mb-3 flex items-center justify-between gap-3 rounded-lg border border-amber-500/40 bg-amber-500/15 px-3 py-2 text-xs">
					<div>
						<p className="font-semibold text-amber-200">{t("voice.idle.warning_title")}</p>
						<p className="text-amber-300/90">{t("voice.idle.warning_description")}</p>
					</div>
					<Button
						type="button"
						size="sm"
						variant="secondary"
						onClick={() => void acknowledgeIdle()}
					>
						{t("voice.idle.stay_connected")}
					</Button>
				</div>
			) : null}

			<div className="mb-3 flex flex-wrap items-center justify-between gap-2">
				<div className="flex items-center gap-2">
					<h2 className="text-sm font-semibold text-text-primary">{channelName}</h2>
					{isThisChannelActive && isRecording ? (
						<span
							className="inline-flex items-center gap-1.5 rounded-full bg-red-500/15 px-2.5 py-0.5 text-xs font-medium text-red-400"
							data-testid="voice-recording-badge"
						>
							<Radio className="size-3 animate-pulse" />
							{t("voice.recording.badge")}
						</span>
					) : null}
					{isThisChannelActive && isScreenSharing ? (
						<span className="text-xs text-text-muted">{t("voice.stage.screen_share_hint")}</span>
					) : null}
				</div>

				<div className="flex items-center gap-1.5">
					{isThisChannelActive ? (
						<>
							<Button
								type="button"
								size="sm"
								variant="ghost"
								onClick={() => setLayoutMode((prev) => (prev === "grid" ? "focus" : "grid"))}
							>
								{effectiveLayout === "grid" ? (
									<>
										<Maximize2 className="mr-1.5 size-3.5" />
										{t("voice.controls.focus_view")}
									</>
								) : (
									<>
										<LayoutGrid className="mr-1.5 size-3.5" />
										{t("voice.controls.grid_view")}
									</>
								)}
							</Button>

							{canRecordCall ? (
								<Button
									type="button"
									size="sm"
									variant={isRecording ? "destructive" : "outline"}
									disabled={startRecMutation.isPending || stopRecMutation.isPending}
									onClick={handleToggleRecording}
								>
									{isRecording ? (
										<>
											<Square className="mr-1.5 size-3.5" />
											{t("voice.controls.stop_recording")}
										</>
									) : (
										<>
											<Radio className="mr-1.5 size-3.5 text-red-400" />
											{t("voice.controls.start_recording")}
										</>
									)}
								</Button>
							) : null}
						</>
					) : null}

					<Button
						type="button"
						size="sm"
						variant="outline"
						onClick={() => setShowRecordings((prev) => !prev)}
					>
						{t("voice.stage.recordings_title")}
					</Button>

					{isThisChannelActive ? (
						<Button
							type="button"
							size="sm"
							variant="destructive"
							onClick={() => void leaveChannel()}
						>
							<PhoneOff className="mr-1.5 size-3.5" />
							{t("voice.controls.leave_voice")}
						</Button>
					) : (
						<Button
							type="button"
							size="sm"
							onClick={() =>
								void joinChannel({
									channelId,
									channelName,
									groupId,
									groupName,
								})
							}
						>
							<Phone className="mr-1.5 size-3.5" />
							{t("voice.controls.join_voice")}
						</Button>
					)}
				</div>
			</div>

			{showRecordings ? (
				<div className="mb-3 rounded-xl border border-chat-surface-border bg-chat-surface-elevated p-3">
					<p className="mb-2 text-xs font-semibold text-text-secondary">
						{t("voice.stage.recordings_title")}
					</p>
					{(recordingsQuery.data?.recordings ?? []).length === 0 ? (
						<p className="text-xs text-text-muted">{t("voice.stage.no_recordings")}</p>
					) : (
						<div className="flex flex-col gap-1.5">
							{recordingsQuery.data?.recordings.map((rec) => (
								<div
									key={rec.id}
									className="flex items-center justify-between rounded-md border border-chat-surface-border px-2.5 py-1.5 text-xs"
								>
									<div className="min-w-0">
										<span className="font-medium text-text-primary">
											{new Date(rec.started_at).toLocaleString()}
										</span>
										<span className="ml-2 text-text-muted">
											({rec.status}, {rec.duration_seconds}s)
										</span>
									</div>
									{rec.download_url ? (
										<a
											href={rec.download_url}
											target="_blank"
											rel="noreferrer"
											className="inline-flex items-center gap-1 text-primary hover:underline"
										>
											<Download className="size-3.5" />
											{t("voice.stage.download_recording")}
										</a>
									) : null}
								</div>
							))}
						</div>
					)}
				</div>
			) : null}

			{!isThisChannelActive ? (
				<div className="flex flex-1 flex-col items-center justify-center rounded-xl border border-dashed border-chat-surface-border p-8 text-center">
					<p className="text-sm font-medium text-text-primary">{t("voice.stage.empty_title")}</p>
					<p className="mt-1 text-xs text-text-muted">{t("voice.stage.empty_description")}</p>
				</div>
			) : (
				<div
					className={cn(
						"grid gap-3 overflow-y-auto",
						effectiveLayout === "focus"
							? "grid-cols-2 sm:grid-cols-3 lg:grid-cols-4"
							: participants.length <= 1
								? "grid-cols-1"
								: participants.length <= 4
									? "grid-cols-1 sm:grid-cols-2"
									: "grid-cols-2 lg:grid-cols-3",
					)}
				>
					{participants.map((p) => (
						<VoiceParticipantTile
							key={p.identity}
							participant={p}
							channelId={channelId}
							groupId={groupId}
							volumePercent={settings.userVolumes[p.identity] ?? 100}
							locallyMuted={Boolean(settings.locallyMutedUsers[p.identity])}
							canMuteMembers={canMuteMembers || isAdmin}
							canDeafenMembers={canDeafenMembers || isAdmin}
							canMoveMembers={canMoveMembers || isAdmin}
							focused={effectiveLayout === "focus" && p.identity === effectiveFocusedId}
							preferScreenTrack={p.identity === effectiveFocusedId && p.isScreenSharing}
							onFocus={() =>
								setFocusedIdentity((prev) => (prev === p.identity ? null : p.identity))
							}
							onModerate={handleModerate}
						/>
					))}
				</div>
			)}

			<VoiceRecordingConsentDialog
				open={isThisChannelActive && recordingConsentOpen}
				onAccept={acceptRecordingConsent}
				onLeave={() => void declineRecordingConsent()}
			/>
		</div>
	);
}
