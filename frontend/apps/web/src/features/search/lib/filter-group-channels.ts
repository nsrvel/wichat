import type { Channel } from "@wichat/contracts/chat/channel";

export type ShellChannelSearchHit = {
	groupId: string;
	channelId: string;
	name: string;
	type: string;
};

const MAX_HITS = 12;

export function filterGroupChannels(
	channels: Channel[] | undefined,
	groupId: string,
	query: string,
): ShellChannelSearchHit[] {
	const q = query.trim().toLowerCase();
	if (!q || !groupId || !channels?.length) {
		return [];
	}

	const hits: ShellChannelSearchHit[] = [];
	for (const ch of channels) {
		if (ch.type === "dm") {
			continue;
		}
		const name = (ch.name ?? "").toLowerCase();
		if (!name.includes(q)) {
			continue;
		}
		hits.push({
			groupId,
			channelId: ch.id,
			name: ch.name,
			type: ch.type,
		});
		if (hits.length >= MAX_HITS) {
			break;
		}
	}
	return hits;
}
