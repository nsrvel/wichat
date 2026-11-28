"use client";

import type { Channel } from "@wichat/contracts/chat/channel";
import type { GroupMember } from "@wichat/contracts/group/member";
import { ChannelListIcon } from "@/components/layout/channel-list-icon";
import { UserAvatar } from "@/components/ui/user-avatar";
import { memberPresenceDot } from "@/features/presence/presence-dot";
import { userInitials } from "@/lib/user-initials";

type DmSidebarChannelIconProps = {
	channel: Channel;
	peer?: GroupMember;
};

export function DmSidebarChannelIcon({ channel, peer }: DmSidebarChannelIconProps) {
	if (!peer) {
		return (
			<ChannelListIcon
				type={channel.type}
				channelName={channel.name}
				imageUrl={channel.image_url}
			/>
		);
	}

	return (
		<UserAvatar
			size="xs"
			src={peer.avatar_url}
			initials={userInitials(peer.display_name, peer.username)}
			status={memberPresenceDot(peer)}
			statusRingClassName="ring-sidebar"
			alt=""
		/>
	);
}
