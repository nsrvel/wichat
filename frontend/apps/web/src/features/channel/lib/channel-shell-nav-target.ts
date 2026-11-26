import type { Channel } from "@wichat/contracts/chat/channel";
import type { GroupMember } from "@wichat/contracts/group/member";
import type { ShellNavTarget } from "@/components/layout/shell-navigation";
import { formatChannelDisplayName } from "@/features/channel/lib/channel-display";
import { dmPeerDisplayLabel } from "@/features/channel/lib/dm-sidebar-peer";
import { memberPresenceDot } from "@/features/presence/presence-dot";
import { userInitials } from "@/lib/user-initials";

export function toChannelShellNavTarget(
	channelId: string,
	channels: readonly Channel[] | undefined,
	peer?: GroupMember,
): ShellNavTarget {
	const match = channels?.find((c) => c.id === channelId);
	const name = match?.name ?? "";
	if (match?.type === "dm") {
		const label = dmPeerDisplayLabel(name, peer);
		return {
			id: channelId,
			label,
			kind: "dm",
			dmAvatarSrc: peer?.avatar_url,
			dmAvatarInitials: peer
				? userInitials(peer.display_name, peer.username)
				: label.slice(0, 2).toUpperCase(),
			dmPresence: peer ? memberPresenceDot(peer) : undefined,
		};
	}
	return {
		id: channelId,
		label: formatChannelDisplayName(name),
		kind: "channel",
		channelType: match?.type,
		imageUrl: match?.image_url,
		channelName: name,
	};
}
