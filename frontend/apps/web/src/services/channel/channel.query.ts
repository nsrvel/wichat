import { queryOptions } from "@tanstack/react-query";
import { isPublicEntityId } from "@wichat/contracts";
import { Q_GROUP_CHANNELS } from "@/services/channel/channel.key";
import { listGroupChannels } from "@/services/channel/channel.service";

export function groupChannelsQueryOpts(groupId: string) {
	return queryOptions({
		queryKey: [Q_GROUP_CHANNELS, groupId],
		queryFn: () => listGroupChannels(groupId),
		enabled: isPublicEntityId(groupId),
		staleTime: 5 * 60_000,
		refetchOnWindowFocus: false,
	});
}
