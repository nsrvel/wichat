import type { Channel } from "@wichat/contracts";
import type { ShellNavTarget } from "@/components/layout/shell-navigation";

export function toChannelShellNavTarget(
	channelId: string,
	channels: readonly Channel[] | undefined,
): ShellNavTarget {
	const match = channels?.find((c) => c.id === channelId);
	return {
		id: channelId,
		label: match?.name ?? "",
		kind: "channel",
	};
}
