import type { Channel } from "@wichat/contracts";

/** Group channels by `sidebar_section` from ms-user (see channels.sidebar_section). */
export function partitionChannelsBySidebarSection(channels: readonly Channel[]) {
	const favorites = channels.filter((c) => c.sidebar_section === "favorites");
	const listChannels = channels.filter((c) => c.sidebar_section !== "favorites");

	return { favorites, listChannels };
}
