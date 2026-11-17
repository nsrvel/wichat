import { z } from "zod";
import { publicEntityIdSchema } from "../common";

export const voiceOccupantSchema = z.object({
	user_id: z.string(),
	display_name: z.string(),
	username: z.string(),
	avatar_url: z.string().optional().default(""),
	is_muted: z.boolean().default(false),
	is_deafened: z.boolean().default(false),
	is_server_muted: z.boolean().default(false),
	is_server_deafened: z.boolean().default(false),
	is_camera_on: z.boolean().default(false),
	is_screen_sharing: z.boolean().default(false),
	joined_at: z.string(),
});
export type VoiceOccupant = z.infer<typeof voiceOccupantSchema>;

export const channelOccupancySchema = z.object({
	channel_id: z.string(),
	is_recording: z.boolean().default(false),
	occupants: z.array(voiceOccupantSchema).default([]),
});
export type ChannelOccupancy = z.infer<typeof channelOccupancySchema>;

export const GroupOccupancyResponseSchema = z.object({
	channels: z.array(channelOccupancySchema).default([]),
});
export type GroupOccupancyResponse = z.infer<
	typeof GroupOccupancyResponseSchema
>;

export const IssueVoiceTokenResponseSchema = z.object({
	token: z.string(),
	livekit_url: z.string(),
	room_name: z.string(),
	channel_id: z.string(),
	expires_at: z.string(),
	initial_server_muted: z.boolean().default(false),
	can_publish_audio: z.boolean().default(true),
	can_publish_video: z.boolean().default(true),
	use_voice_activity: z.boolean().default(true),
	bitrate_kbps: z.number().int().default(64),
	video_quality: z.string().default("auto"),
	is_recording: z.boolean().default(false),
	occupants: z.array(voiceOccupantSchema).default([]),
});
export type IssueVoiceTokenResponse = z.infer<
	typeof IssueVoiceTokenResponseSchema
>;

export const voiceModerationActionSchema = z.enum([
	"mute",
	"unmute",
	"deafen",
	"undeafen",
	"disconnect",
]);
export type VoiceModerationAction = z.infer<typeof voiceModerationActionSchema>;

export const ModerateParticipantRequestSchema = z.object({
	group_id: publicEntityIdSchema.optional(),
	target_user_id: z.string(),
	action: voiceModerationActionSchema,
	reason: z.string().max(240).optional(),
});
export type ModerateParticipantRequest = z.infer<
	typeof ModerateParticipantRequestSchema
>;

export const ModerateParticipantResponseSchema = z.object({
	ok: z.boolean(),
	occupant: voiceOccupantSchema,
});
export type ModerateParticipantResponse = z.infer<
	typeof ModerateParticipantResponseSchema
>;

export const CallRecordingSchema = z.object({
	id: z.string(),
	session_id: z.string(),
	channel_id: z.string(),
	egress_id: z.string(),
	status: z.string(),
	file_key: z.string().optional().default(""),
	file_size_bytes: z.number().default(0),
	duration_seconds: z.number().int().default(0),
	started_by: z.string(),
	stopped_by: z.string().optional().default(""),
	started_at: z.string(),
	ended_at: z.string().optional().default(""),
	download_url: z.string().optional().default(""),
});
export type CallRecording = z.infer<typeof CallRecordingSchema>;

export const ListChannelRecordingsResponseSchema = z.object({
	recordings: z.array(CallRecordingSchema).default([]),
	next_cursor: z.string().optional().default(""),
});
export type ListChannelRecordingsResponse = z.infer<
	typeof ListChannelRecordingsResponseSchema
>;

export const QuotaUsageResponseSchema = z.object({
	month: z.string(),
	used_participant_minutes: z.number(),
	limit_participant_minutes: z.number(),
	usage_percent: z.number(),
	warning_active: z.boolean(),
});
export type QuotaUsageResponse = z.infer<typeof QuotaUsageResponseSchema>;

export const VoiceOccupancyUpdatedEventSchema = z.object({
	type: z.literal("voice.occupancy.updated"),
	group_id: z.string(),
	channel_id: z.string(),
	is_recording: z.boolean().default(false),
	occupants: z.array(voiceOccupantSchema).default([]),
});
export type VoiceOccupancyUpdatedEvent = z.infer<
	typeof VoiceOccupancyUpdatedEventSchema
>;

export const VoiceRecordingUpdatedEventSchema = z.object({
	type: z.literal("voice.recording.updated"),
	group_id: z.string(),
	channel_id: z.string(),
	recording_id: z.string(),
	status: z.string(),
	is_recording: z.boolean(),
	started_by: z.string().optional().default(""),
	download_url: z.string().optional().default(""),
});
export type VoiceRecordingUpdatedEvent = z.infer<
	typeof VoiceRecordingUpdatedEventSchema
>;

export const VoiceModerationAppliedEventSchema = z.object({
	type: z.literal("voice.moderation.applied"),
	group_id: z.string(),
	channel_id: z.string(),
	target_user_id: z.string(),
	actor_user_id: z.string(),
	action: z.string(),
	reason: z.string().optional().default(""),
});
export type VoiceModerationAppliedEvent = z.infer<
	typeof VoiceModerationAppliedEventSchema
>;

export const VoiceIdleWarningEventSchema = z.object({
	type: z.literal("voice.idle.warning"),
	group_id: z.string(),
	channel_id: z.string(),
	user_id: z.string(),
	seconds_remaining: z.number().int(),
});
export type VoiceIdleWarningEvent = z.infer<typeof VoiceIdleWarningEventSchema>;

export interface VoiceUserSettings {
	inputDeviceId: string;
	outputDeviceId: string;
	cameraDeviceId: string;
	inputMode: "vad" | "ptt";
	vadThreshold: number; // 0..100
	pttKey: string; // KeyboardEvent.code e.g. "Space" or "KeyV"
	echoCancellation: boolean;
	noiseSuppression: boolean;
	autoGainControl: boolean;
	userVolumes: Record<string, number>; // userId -> 0..200
	locallyMutedUsers: Record<string, boolean>;
}
