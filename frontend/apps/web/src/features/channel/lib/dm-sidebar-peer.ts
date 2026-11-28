import type { Channel } from "@wichat/contracts/chat/channel";
import type { GroupMember } from "@wichat/contracts/group/member";

type DmPeerNameFields = Pick<GroupMember, "display_name" | "username">;

/** Sidebar / header DM title — display name as stored, no channel-name formatting. */
export function dmPeerDisplayLabel(
	fallbackChannelName: string,
	peer?: DmPeerNameFields | null,
): string {
	const display = peer?.display_name?.trim();
	if (display) {
		return display;
	}
	const username = peer?.username?.trim();
	if (username) {
		return username;
	}
	return fallbackChannelName.trim();
}

/** Resolve the other participant in a group DM for sidebar avatar/presence. */
export function dmSidebarPeerMember(
	channel: Channel,
	members: GroupMember[] | undefined,
	selfUserId?: string,
): GroupMember | undefined {
	if (!members?.length) {
		return undefined;
	}
	const peerId = channel.dm_peer_user_id?.trim();
	if (peerId) {
		const byId = members.find((m) => m.user_id === peerId);
		if (byId) {
			return byId;
		}
	}
	const label = channel.name.trim().toLowerCase();
	if (!label) {
		return undefined;
	}
	return members.find(
		(m) =>
			m.user_id !== selfUserId &&
			(m.display_name.trim().toLowerCase() === label || m.username.trim().toLowerCase() === label),
	);
}
