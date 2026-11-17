import {
	CallRecordingSchema,
	GroupOccupancyResponseSchema,
	IssueVoiceTokenResponseSchema,
	ListChannelRecordingsResponseSchema,
	ModerateParticipantRequestSchema,
	ModerateParticipantResponseSchema,
	QuotaUsageResponseSchema,
	type CallRecording,
	type GroupOccupancyResponse,
	type IssueVoiceTokenResponse,
	type ListChannelRecordingsResponse,
	type ModerateParticipantRequest,
	type ModerateParticipantResponse,
	type QuotaUsageResponse,
} from "@wichat/contracts/chat/voice";
import { OkSchema } from "@wichat/contracts/common";
import { api } from "@/integrations/api-client";

export async function issueVoiceToken(
	channelId: string,
	groupId?: string,
): Promise<IssueVoiceTokenResponse> {
	const path = groupId
		? `/groups/${encodeURIComponent(groupId)}/channels/${encodeURIComponent(channelId)}/voice/token`
		: `/channels/${encodeURIComponent(channelId)}/voice/token`;
	return api.post(path, IssueVoiceTokenResponseSchema, groupId ? { group_id: groupId } : {});
}

export async function listGroupOccupancy(groupId: string): Promise<GroupOccupancyResponse> {
	return api.get(
		`/groups/${encodeURIComponent(groupId)}/voice/occupancy`,
		GroupOccupancyResponseSchema,
	);
}

export async function moderateVoiceParticipant(
	channelId: string,
	payload: ModerateParticipantRequest,
): Promise<ModerateParticipantResponse> {
	const body = ModerateParticipantRequestSchema.parse(payload);
	return api.post(
		`/channels/${encodeURIComponent(channelId)}/voice/moderate`,
		ModerateParticipantResponseSchema,
		body,
	);
}

export async function acknowledgeVoiceIdleWarning(
	channelId: string,
	groupId?: string,
): Promise<{ ok: boolean }> {
	return api.post(
		`/channels/${encodeURIComponent(channelId)}/voice/keepalive`,
		OkSchema,
		groupId ? { group_id: groupId } : {},
	);
}

export async function startCallRecording(
	channelId: string,
	groupId?: string,
): Promise<CallRecording> {
	return api.post(
		`/channels/${encodeURIComponent(channelId)}/voice/recording/start`,
		CallRecordingSchema,
		groupId ? { group_id: groupId } : {},
	);
}

export async function stopCallRecording(
	channelId: string,
	groupId?: string,
	recordingId?: string,
): Promise<CallRecording> {
	return api.post(
		`/channels/${encodeURIComponent(channelId)}/voice/recording/stop`,
		CallRecordingSchema,
		{ group_id: groupId, recording_id: recordingId },
	);
}

export async function listChannelRecordings(
	channelId: string,
	groupId?: string,
	limit = 25,
): Promise<ListChannelRecordingsResponse> {
	const params = new URLSearchParams();
	if (groupId) params.set("group_id", groupId);
	params.set("limit", String(limit));
	return api.get(
		`/channels/${encodeURIComponent(channelId)}/voice/recordings?${params.toString()}`,
		ListChannelRecordingsResponseSchema,
	);
}

export async function getVoiceQuotaUsage(groupId?: string): Promise<QuotaUsageResponse> {
	const qs = groupId ? `?group_id=${encodeURIComponent(groupId)}` : "";
	return api.get(`/voice/quota${qs}`, QuotaUsageResponseSchema);
}
