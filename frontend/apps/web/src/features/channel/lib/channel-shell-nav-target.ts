import type { Channel } from "@wichat/contracts";
import type { ShellNavTarget } from "@/components/layout/shell-navigation";
import { formatChannelDisplayName } from "@/features/channel/lib/channel-display";

export function toChannelShellNavTarget(
	channelId: string,
	channels: readonly Channel[] | undefined,
): ShellNavTarget {
	const match = channels?.find((c) => c.id === channelId);
	const name = match?.name ?? "";
	return {
		id: channelId,
		label: formatChannelDisplayName(name),
		kind: "channel",
		channelType: match?.type,
		imageUrl: match?.image_url,
		channelName: name,
	};
}
