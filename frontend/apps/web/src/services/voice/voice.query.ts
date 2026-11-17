import { queryOptions } from "@tanstack/react-query";
import { isPublicEntityId } from "@wichat/contracts/common";
import { Q_VOICE_OCCUPANCY, Q_VOICE_QUOTA, Q_VOICE_RECORDINGS } from "@/services/voice/voice.key";
import {
	getVoiceQuotaUsage,
	listChannelRecordings,
	listGroupOccupancy,
} from "@/services/voice/voice.service";

export function groupVoiceOccupancyQueryOpts(groupId: string) {
	return queryOptions({
		queryKey: [Q_VOICE_OCCUPANCY, groupId],
		queryFn: () => listGroupOccupancy(groupId),
		enabled: isPublicEntityId(groupId),
		staleTime: 15_000,
		refetchOnWindowFocus: false,
	});
}

export function channelRecordingsQueryOpts(channelId: string, groupId?: string) {
	return queryOptions({
		queryKey: [Q_VOICE_RECORDINGS, channelId, groupId ?? ""],
		queryFn: () => listChannelRecordings(channelId, groupId),
		enabled: isPublicEntityId(channelId),
		staleTime: 30_000,
		refetchOnWindowFocus: false,
	});
}

export function voiceQuotaQueryOpts(groupId?: string) {
	return queryOptions({
		queryKey: [Q_VOICE_QUOTA, groupId ?? ""],
		queryFn: () => getVoiceQuotaUsage(groupId),
		staleTime: 60_000,
		refetchOnWindowFocus: false,
	});
}
