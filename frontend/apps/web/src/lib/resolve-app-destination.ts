import type { Group } from "@wichat/contracts";
import { APP_HOME_PATH, channelPath } from "@/lib/app-routes";
import type { LastChannelRef } from "@/lib/last-channel-storage";

export function resolveAppDestination(
	groups: readonly Group[],
	lastChannel: LastChannelRef | null,
): string {
	if (!groups.length) {
		return APP_HOME_PATH;
	}

	const groupIds = new Set(groups.map((g) => g.id));
	if (lastChannel && groupIds.has(lastChannel.groupId)) {
		return channelPath(lastChannel.groupId, lastChannel.channelId);
	}

	const first = groups[0];
	if (first?.default_channel_id) {
		return channelPath(first.id, first.default_channel_id);
	}

	return APP_HOME_PATH;
}
