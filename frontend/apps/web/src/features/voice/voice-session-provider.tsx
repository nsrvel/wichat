"use client";

import {
	createContext,
	useCallback,
	useContext,
	useEffect,
	useMemo,
	useRef,
	useState,
	useSyncExternalStore,
	type ReactNode,
} from "react";
import {
	ConnectionState,
	DisconnectReason,
	Room,
	RoomEvent,
	Track,
	VideoPresets,
	type LocalTrackPublication,
	type Participant,
	type RemoteParticipant,
	type RemoteTrack,
	type RemoteTrackPublication,
} from "livekit-client";
import type { IssueVoiceTokenResponse, VoiceUserSettings } from "@wichat/contracts/chat/voice";
import { useQueryClient } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { translateError } from "@/integrations/i18n/translate-error";
import { subscribeVoiceRealtime } from "@/integrations/realtime/voice-realtime";
import { Q_VOICE_OCCUPANCY } from "@/services/voice/voice.key";
import { acknowledgeVoiceIdleWarning, issueVoiceToken } from "@/services/voice/voice.service";
import {
	playVoiceDeafenCue,
	playVoiceJoinCue,
	playVoiceLeaveCue,
	playVoiceMuteCue,
} from "@/features/voice/lib/voice-audio-cues";
import {
	DEFAULT_VOICE_SETTINGS,
	readVoiceSettings,
	subscribeVoiceSettings,
} from "@/features/voice/lib/voice-settings-storage";

export type VoiceConnectionPhase = "disconnected" | "connecting" | "connected" | "reconnecting";

export interface ActiveVoiceChannel {
	channelId: string;
	channelName: string;
	groupId?: string;
	groupName?: string;
}

export interface VoiceParticipantSnapshot {
	identity: string;
	name: string;
	avatarUrl: string;
	isLocal: boolean;
	isSpeaking: boolean;
	isMuted: boolean;
	isCameraOn: boolean;
	isScreenSharing: boolean;
	cameraTrack?: Track;
	screenTrack?: Track;
}

interface VoiceSessionContextValue {
	phase: VoiceConnectionPhase;
	activeChannel: ActiveVoiceChannel | null;
	participants: VoiceParticipantSnapshot[];
	isMuted: boolean;
	isDeafened: boolean;
	isServerMuted: boolean;
	isServerDeafened: boolean;
	isCameraOn: boolean;
	isScreenSharing: boolean;
	isRecording: boolean;
	canPublishAudio: boolean;
	canPublishVideo: boolean;
	idleWarningSeconds: number | null;
	recordingConsentOpen: boolean;
	settings: VoiceUserSettings;
	joinChannel: (target: ActiveVoiceChannel) => Promise<void>;
	leaveChannel: () => Promise<void>;
	toggleMute: () => Promise<void>;
	toggleDeafen: () => Promise<void>;
	toggleCamera: () => Promise<void>;
	toggleScreenShare: () => Promise<void>;
	acknowledgeIdle: () => Promise<void>;
	acceptRecordingConsent: () => void;
	declineRecordingConsent: () => Promise<void>;
}

const VoiceSessionContext = createContext<VoiceSessionContextValue | null>(null);

function parseParticipantMeta(participant: Participant): {
	displayName: string;
	avatarUrl: string;
} {
	let displayName = participant.name || participant.identity;
	let avatarUrl = "";
	if (participant.metadata) {
		try {
			const parsed = JSON.parse(participant.metadata) as {
				display_name?: string;
				avatar_url?: string;
			};
			if (parsed.display_name) displayName = parsed.display_name;
			if (parsed.avatar_url) avatarUrl = parsed.avatar_url;
		} catch {
			// ignore malformed metadata
		}
	}
	return { displayName, avatarUrl };
}

function buildParticipantSnapshots(room: Room | null): VoiceParticipantSnapshot[] {
	if (!room) return [];
	const list: Participant[] = [
		room.localParticipant,
		...Array.from(room.remoteParticipants.values()),
	];
	return list.map((p) => {
		const meta = parseParticipantMeta(p);
		const micPub = p.getTrackPublication(Track.Source.Microphone);
		const camPub = p.getTrackPublication(Track.Source.Camera);
		const screenPub = p.getTrackPublication(Track.Source.ScreenShare);
		return {
			identity: p.identity,
			name: meta.displayName,
			avatarUrl: meta.avatarUrl,
			isLocal: p.isLocal,
			isSpeaking: p.isSpeaking,
			isMuted: micPub ? micPub.isMuted : !p.isMicrophoneEnabled,
			isCameraOn: Boolean(camPub && !camPub.isMuted && camPub.track),
			isScreenSharing: Boolean(screenPub && !screenPub.isMuted && screenPub.track),
			cameraTrack: camPub?.track,
			screenTrack: screenPub?.track,
		};
	});
}

export function VoiceSessionProvider({
	currentUserId,
	children,
}: {
	currentUserId?: string;
	children: ReactNode;
}) {
	const { t } = useTranslation();
	const qc = useQueryClient();
	const settings = useSyncExternalStore(
		subscribeVoiceSettings,
		readVoiceSettings,
		() => DEFAULT_VOICE_SETTINGS,
	);

	const roomRef = useRef<Room | null>(null);
	const joiningChannelIdRef = useRef<string | null>(null);
	const refreshTimerRef = useRef<number | null>(null);
	const audioNodesRef = useRef<
		Map<
			string,
			{
				audioEl: HTMLAudioElement;
				ctx?: AudioContext;
				gain?: GainNode;
			}
		>
	>(new Map());

	const [phase, setPhase] = useState<VoiceConnectionPhase>("disconnected");
	const [activeChannel, setActiveChannel] = useState<ActiveVoiceChannel | null>(null);
	const [participants, setParticipants] = useState<VoiceParticipantSnapshot[]>([]);
	const [isMuted, setIsMuted] = useState(false);
	const [isDeafened, setIsDeafened] = useState(false);
	const [isServerMuted, setIsServerMuted] = useState(false);
	const [isServerDeafened, setIsServerDeafened] = useState(false);
	const [isCameraOn, setIsCameraOn] = useState(false);
	const [isScreenSharing, setIsScreenSharing] = useState(false);
	const [isRecording, setIsRecording] = useState(false);
	const [canPublishAudio, setCanPublishAudio] = useState(true);
	const [canPublishVideo, setCanPublishVideo] = useState(true);
	const [idleWarningSeconds, setIdleWarningSeconds] = useState<number | null>(null);
	const [recordingConsentOpen, setRecordingConsentOpen] = useState(false);

	const isDeafenedRef = useRef(isDeafened);
	const settingsRef = useRef(settings);
	const activeChannelRef = useRef(activeChannel);

	useEffect(() => {
		isDeafenedRef.current = isDeafened;
		settingsRef.current = settings;
		activeChannelRef.current = activeChannel;
	}, [isDeafened, settings, activeChannel]);

	const syncParticipants = useCallback(() => {
		setParticipants(buildParticipantSnapshots(roomRef.current));
	}, []);

	const applyAudioRoutingForParticipant = useCallback(
		(identity: string) => {
			const node = audioNodesRef.current.get(identity);
			if (!node) return;
			const curSettings = settingsRef.current;
			const locallyMuted = Boolean(curSettings.locallyMutedUsers[identity]);
			const rawPercent = curSettings.userVolumes[identity] ?? 100;
			const effectiveGain =
				isDeafenedRef.current || isServerDeafened || locallyMuted
					? 0
					: Math.max(0, Math.min(2, rawPercent / 100));

			if (node.gain) {
				node.gain.gain.value = effectiveGain;
			} else {
				node.audioEl.volume = Math.min(1, effectiveGain);
				node.audioEl.muted = effectiveGain === 0;
			}
			const elWithSink = node.audioEl as HTMLAudioElement & {
				setSinkId?: (id: string) => Promise<void>;
			};
			if (curSettings.outputDeviceId && typeof elWithSink.setSinkId === "function") {
				void elWithSink.setSinkId(curSettings.outputDeviceId).catch(() => {});
			}
		},
		[isServerDeafened],
	);

	// Re-apply GainNode routing whenever deafen or per-user volumes change
	useEffect(() => {
		audioNodesRef.current.forEach((_val, identity) => {
			applyAudioRoutingForParticipant(identity);
		});
	}, [
		isDeafened,
		isServerDeafened,
		settings.userVolumes,
		settings.locallyMutedUsers,
		settings.outputDeviceId,
		applyAudioRoutingForParticipant,
	]);

	const cleanupAudioNodes = useCallback(() => {
		audioNodesRef.current.forEach((node) => {
			try {
				node.audioEl.pause();
				node.audioEl.srcObject = null;
				node.audioEl.remove();
				void node.ctx?.close();
			} catch {
				// ignore cleanup failure
			}
		});
		audioNodesRef.current.clear();
	}, []);

	const clearRefreshTimer = useCallback(() => {
		if (refreshTimerRef.current !== null) {
			window.clearTimeout(refreshTimerRef.current);
			refreshTimerRef.current = null;
		}
	}, []);

	const teardownRoom = useCallback(
		async (playCue: boolean) => {
			clearRefreshTimer();
			const existing = roomRef.current;
			roomRef.current = null;
			if (existing) {
				try {
					await existing.disconnect(true);
				} catch {
					// ignore disconnect error
				}
			}
			cleanupAudioNodes();
			setPhase("disconnected");
			setActiveChannel(null);
			setParticipants([]);
			setIsCameraOn(false);
			setIsScreenSharing(false);
			setIsRecording(false);
			setIsServerMuted(false);
			setIsServerDeafened(false);
			setIdleWarningSeconds(null);
			setRecordingConsentOpen(false);
			if (playCue) {
				playVoiceLeaveCue();
			}
		},
		[clearRefreshTimer, cleanupAudioNodes],
	);

	const scheduleTokenRefresh = useCallback(
		(expiresAtIso: string, target: ActiveVoiceChannel) => {
			const scheduleNext = (iso: string) => {
				clearRefreshTimer();
				const expMs = Date.parse(iso);
				if (!Number.isFinite(expMs)) return;
				const delayMs = Math.max(30_000, expMs - Date.now() - 90_000);
				refreshTimerRef.current = window.setTimeout(() => {
					if (!roomRef.current || activeChannelRef.current?.channelId !== target.channelId) {
						return;
					}
					void issueVoiceToken(target.channelId, target.groupId)
						.then((fresh) => {
							scheduleNext(fresh.expires_at);
						})
						.catch(() => {
							// LiveKit connection remains valid until expiry; retry shortly
						});
				}, delayMs);
			};
			scheduleNext(expiresAtIso);
		},
		[clearRefreshTimer],
	);

	const attachRemoteAudioTrack = useCallback(
		(track: RemoteTrack, participant: RemoteParticipant) => {
			if (track.kind !== Track.Kind.Audio) return;
			const audioEl = track.attach() as HTMLAudioElement;
			audioEl.autoplay = true;
			audioEl.style.display = "none";
			document.body.appendChild(audioEl);

			let ctx: AudioContext | undefined;
			let gain: GainNode | undefined;
			try {
				const AudioCtx = window.AudioContext;
				if (AudioCtx && audioEl.srcObject instanceof MediaStream) {
					ctx = new AudioCtx();
					const source = ctx.createMediaStreamSource(audioEl.srcObject);
					gain = ctx.createGain();
					source.connect(gain);
					gain.connect(ctx.destination);
					// Mute raw element so only GainNode outputs (allows 0–200% amplification)
					audioEl.muted = true;
				}
			} catch {
				ctx = undefined;
				gain = undefined;
			}

			audioNodesRef.current.set(participant.identity, { audioEl, ctx, gain });
			applyAudioRoutingForParticipant(participant.identity);
		},
		[applyAudioRoutingForParticipant],
	);

	const detachRemoteAudioTrack = useCallback(
		(track: RemoteTrack, participant: RemoteParticipant) => {
			if (track.kind !== Track.Kind.Audio) return;
			track.detach();
			const node = audioNodesRef.current.get(participant.identity);
			if (node) {
				node.audioEl.remove();
				void node.ctx?.close().catch(() => {});
				audioNodesRef.current.delete(participant.identity);
			}
		},
		[],
	);

	const connectWithToken = useCallback(
		async (target: ActiveVoiceChannel, tokenRes: IssueVoiceTokenResponse) => {
			const room = new Room({
				adaptiveStream: true,
				dynacast: true,
				audioCaptureDefaults: {
					deviceId: settingsRef.current.inputDeviceId || undefined,
					echoCancellation: settingsRef.current.echoCancellation,
					noiseSuppression: settingsRef.current.noiseSuppression,
					autoGainControl: settingsRef.current.autoGainControl,
				},
				videoCaptureDefaults: {
					deviceId: settingsRef.current.cameraDeviceId || undefined,
					resolution: VideoPresets.h720.resolution,
				},
				publishDefaults: {
					simulcast: true,
					screenShareEncoding: {
						maxBitrate: 3_000_000,
						maxFramerate: 30,
					},
				},
			});

			roomRef.current = room;

			room
				.on(RoomEvent.ConnectionStateChanged, (state) => {
					if (state === ConnectionState.Connected) setPhase("connected");
					else if (state === ConnectionState.Reconnecting) setPhase("reconnecting");
					else if (state === ConnectionState.Connecting) setPhase("connecting");
				})
				.on(RoomEvent.ParticipantConnected, syncParticipants)
				.on(RoomEvent.ParticipantDisconnected, syncParticipants)
				.on(RoomEvent.ActiveSpeakersChanged, syncParticipants)
				.on(RoomEvent.LocalTrackPublished, (pub: LocalTrackPublication) => {
					if (pub.source === Track.Source.ScreenShare && pub.track?.mediaStreamTrack) {
						pub.track.mediaStreamTrack.contentHint = "detail";
					}
					syncParticipants();
				})
				.on(RoomEvent.LocalTrackUnpublished, syncParticipants)
				.on(RoomEvent.TrackMuted, syncParticipants)
				.on(RoomEvent.TrackUnmuted, syncParticipants)
				.on(
					RoomEvent.TrackSubscribed,
					(track: RemoteTrack, _pub: RemoteTrackPublication, participant: RemoteParticipant) => {
						attachRemoteAudioTrack(track, participant);
						syncParticipants();
					},
				)
				.on(
					RoomEvent.TrackUnsubscribed,
					(track: RemoteTrack, _pub: RemoteTrackPublication, participant: RemoteParticipant) => {
						detachRemoteAudioTrack(track, participant);
						syncParticipants();
					},
				)
				.on(RoomEvent.Disconnected, (reason) => {
					if (reason === DisconnectReason.PARTICIPANT_REMOVED) {
						toast.info(t("voice.moderation.kicked_notice"));
					}
					void teardownRoom(false);
				});

			await room.connect(tokenRes.livekit_url, tokenRes.token);

			setCanPublishAudio(tokenRes.can_publish_audio);
			setCanPublishVideo(tokenRes.can_publish_video);
			setIsServerMuted(tokenRes.initial_server_muted);
			setIsRecording(tokenRes.is_recording);
			if (tokenRes.is_recording) {
				setRecordingConsentOpen(true);
			}

			const shouldMuteInitially =
				!tokenRes.can_publish_audio ||
				tokenRes.initial_server_muted ||
				isMuted ||
				isDeafenedRef.current ||
				settingsRef.current.inputMode === "ptt" ||
				!tokenRes.use_voice_activity;

			try {
				await room.localParticipant.setMicrophoneEnabled(!shouldMuteInitially);
				setIsMuted(shouldMuteInitially);
			} catch {
				// Microphone permission might not be granted yet; user stays listen-only
				setIsMuted(true);
			}

			syncParticipants();
			scheduleTokenRefresh(tokenRes.expires_at, target);
			playVoiceJoinCue();
		},
		[
			attachRemoteAudioTrack,
			detachRemoteAudioTrack,
			isMuted,
			scheduleTokenRefresh,
			syncParticipants,
			t,
			teardownRoom,
		],
	);

	const joinChannel = useCallback(
		async (target: ActiveVoiceChannel) => {
			if (joiningChannelIdRef.current === target.channelId) {
				return;
			}
			if (activeChannelRef.current?.channelId === target.channelId && roomRef.current) {
				return;
			}
			joiningChannelIdRef.current = target.channelId;
			if (roomRef.current) {
				await teardownRoom(false);
			}
			setPhase("connecting");
			setActiveChannel(target);
			try {
				const tokenRes = await issueVoiceToken(target.channelId, target.groupId);
				await connectWithToken(target, tokenRes);
				if (target.groupId) {
					void qc.invalidateQueries({ queryKey: [Q_VOICE_OCCUPANCY, target.groupId] });
				}
			} catch (err) {
				await teardownRoom(false);
				toast.error(translateError(t, err));
			} finally {
				if (joiningChannelIdRef.current === target.channelId) {
					joiningChannelIdRef.current = null;
				}
			}
		},
		[connectWithToken, qc, t, teardownRoom],
	);

	const leaveChannel = useCallback(async () => {
		const groupId = activeChannelRef.current?.groupId;
		await teardownRoom(true);
		if (groupId) {
			void qc.invalidateQueries({ queryKey: [Q_VOICE_OCCUPANCY, groupId] });
		}
	}, [qc, teardownRoom]);

	const toggleMute = useCallback(async () => {
		if (isServerMuted || !canPublishAudio) return;
		const nextMuted = !isMuted;
		setIsMuted(nextMuted);
		playVoiceMuteCue(nextMuted);
		if (roomRef.current) {
			await roomRef.current.localParticipant.setMicrophoneEnabled(!nextMuted).catch(() => {});
			syncParticipants();
		}
	}, [canPublishAudio, isMuted, isServerMuted, syncParticipants]);

	const toggleDeafen = useCallback(async () => {
		const nextDeafened = !isDeafened;
		setIsDeafened(nextDeafened);
		playVoiceDeafenCue(nextDeafened);
		if (nextDeafened && !isMuted) {
			setIsMuted(true);
			if (roomRef.current) {
				await roomRef.current.localParticipant.setMicrophoneEnabled(false).catch(() => {});
			}
		}
		syncParticipants();
	}, [isDeafened, isMuted, syncParticipants]);

	const toggleCamera = useCallback(async () => {
		if (!roomRef.current || !canPublishVideo) return;
		const next = !isCameraOn;
		try {
			await roomRef.current.localParticipant.setCameraEnabled(next);
			setIsCameraOn(next);
			syncParticipants();
		} catch (err) {
			toast.error(translateError(t, err));
		}
	}, [canPublishVideo, isCameraOn, syncParticipants, t]);

	const toggleScreenShare = useCallback(async () => {
		if (!roomRef.current || !canPublishVideo) return;
		const next = !isScreenSharing;
		try {
			await roomRef.current.localParticipant.setScreenShareEnabled(next, {
				audio: true,
				resolution: VideoPresets.h1080.resolution,
			});
			setIsScreenSharing(next);
			syncParticipants();
		} catch (err) {
			toast.error(translateError(t, err));
		}
	}, [canPublishVideo, isScreenSharing, syncParticipants, t]);

	const acknowledgeIdle = useCallback(async () => {
		const cur = activeChannelRef.current;
		if (!cur) return;
		await acknowledgeVoiceIdleWarning(cur.channelId, cur.groupId);
		setIdleWarningSeconds(null);
	}, []);

	const acceptRecordingConsent = useCallback(() => {
		setRecordingConsentOpen(false);
	}, []);

	const declineRecordingConsent = useCallback(async () => {
		setRecordingConsentOpen(false);
		await leaveChannel();
	}, [leaveChannel]);

	// Push-to-Talk keybind listener when in PTT mode
	useEffect(() => {
		if (
			phase !== "connected" ||
			settings.inputMode !== "ptt" ||
			isServerMuted ||
			isServerDeafened ||
			!canPublishAudio
		) {
			return;
		}
		const targetKey = settings.pttKey || "Space";
		const onKeyDown = (e: KeyboardEvent) => {
			if (e.code !== targetKey || e.repeat) return;
			const tag = (e.target as HTMLElement | null)?.tagName;
			if (
				tag === "INPUT" ||
				tag === "TEXTAREA" ||
				(e.target as HTMLElement | null)?.isContentEditable
			) {
				return;
			}
			setIsMuted(false);
			void roomRef.current?.localParticipant.setMicrophoneEnabled(true).catch(() => {});
		};
		const onKeyUp = (e: KeyboardEvent) => {
			if (e.code !== targetKey) return;
			setIsMuted(true);
			void roomRef.current?.localParticipant.setMicrophoneEnabled(false).catch(() => {});
		};
		window.addEventListener("keydown", onKeyDown);
		window.addEventListener("keyup", onKeyUp);
		return () => {
			window.removeEventListener("keydown", onKeyDown);
			window.removeEventListener("keyup", onKeyUp);
		};
	}, [
		canPublishAudio,
		isServerDeafened,
		isServerMuted,
		phase,
		settings.inputMode,
		settings.pttKey,
	]);

	// Gateway WS listener for moderation, recording, and idle warning
	useEffect(() => {
		return subscribeVoiceRealtime((ev) => {
			const cur = activeChannelRef.current;
			if (!cur || ev.channel_id !== cur.channelId) return;

			if (ev.type === "voice.recording.updated") {
				setIsRecording(ev.is_recording);
				if (ev.is_recording) {
					setRecordingConsentOpen(true);
				}
			} else if (
				ev.type === "voice.moderation.applied" &&
				currentUserId &&
				ev.target_user_id === currentUserId
			) {
				if (ev.action === "mute") {
					setIsServerMuted(true);
					setIsMuted(true);
					void roomRef.current?.localParticipant.setMicrophoneEnabled(false).catch(() => {});
				} else if (ev.action === "unmute") {
					setIsServerMuted(false);
				} else if (ev.action === "deafen") {
					setIsServerDeafened(true);
					setIsDeafened(true);
				} else if (ev.action === "undeafen") {
					setIsServerDeafened(false);
				} else if (ev.action === "disconnect") {
					toast.info(t("voice.moderation.kicked_notice"));
					void teardownRoom(false);
				}
			} else if (
				ev.type === "voice.idle.warning" &&
				currentUserId &&
				ev.user_id === currentUserId
			) {
				setIdleWarningSeconds(ev.seconds_remaining);
			}
		});
	}, [currentUserId, t, teardownRoom]);

	const value = useMemo<VoiceSessionContextValue>(
		() => ({
			phase,
			activeChannel,
			participants,
			isMuted,
			isDeafened,
			isServerMuted,
			isServerDeafened,
			isCameraOn,
			isScreenSharing,
			isRecording,
			canPublishAudio,
			canPublishVideo,
			idleWarningSeconds,
			recordingConsentOpen,
			settings,
			joinChannel,
			leaveChannel,
			toggleMute,
			toggleDeafen,
			toggleCamera,
			toggleScreenShare,
			acknowledgeIdle,
			acceptRecordingConsent,
			declineRecordingConsent,
		}),
		[
			phase,
			activeChannel,
			participants,
			isMuted,
			isDeafened,
			isServerMuted,
			isServerDeafened,
			isCameraOn,
			isScreenSharing,
			isRecording,
			canPublishAudio,
			canPublishVideo,
			idleWarningSeconds,
			recordingConsentOpen,
			settings,
			joinChannel,
			leaveChannel,
			toggleMute,
			toggleDeafen,
			toggleCamera,
			toggleScreenShare,
			acknowledgeIdle,
			acceptRecordingConsent,
			declineRecordingConsent,
		],
	);

	return <VoiceSessionContext.Provider value={value}>{children}</VoiceSessionContext.Provider>;
}

export function useVoiceSession(): VoiceSessionContextValue {
	const ctx = useContext(VoiceSessionContext);
	if (!ctx) {
		throw new Error("useVoiceSession must be used within VoiceSessionProvider");
	}
	return ctx;
}
