import { queryOptions } from "@tanstack/react-query";
import { isPublicEntityId } from "@wichat/contracts/common";
import { Q_CHANNEL_PERMISSIONS, Q_GROUP_CHANNELS } from "@/services/channel/channel.key";
import { listChannelPermissions, listGroupChannels } from "@/services/channel/channel.service";

export function groupChannelsQueryOpts(groupId: string) {
	return queryOptions({
		queryKey: [Q_GROUP_CHANNELS, groupId],
		queryFn: () => listGroupChannels(groupId),
		enabled: isPublicEntityId(groupId),
		staleTime: 5 * 60_000,
		refetchOnWindowFocus: false,
	});
}

export function channelPermissionsQueryOpts(groupId: string, channelId: string) {
	return queryOptions({
		queryKey: [Q_CHANNEL_PERMISSIONS, groupId, channelId],
		queryFn: () => listChannelPermissions(groupId, channelId),
		enabled: isPublicEntityId(groupId) && isPublicEntityId(channelId),
		staleTime: 60_000,
		refetchOnWindowFocus: false,
	});
}
