import { isPublicEntityId } from "@wichat/contracts";
import { isQuickViewId, type QuickViewId } from "@/lib/app-routes";

export type ChannelRouteTarget = {
	groupId?: string;
	channelId?: string;
	quickView?: QuickViewId;
};

/** Parse `/channel/...` pathname for a single persistent channel shell (rail stays mounted). */
export function parseChannelRoute(pathname: string): ChannelRouteTarget {
	const prefix = "/channel/";
	if (!pathname.startsWith(prefix)) {
		return {};
	}

	const rest = pathname.slice(prefix.length);
	const groupSegment = rest.split("/")[0] ?? "";
	if (!groupSegment) {
		return {};
	}

	const groupId = decodeURIComponent(groupSegment);
	if (!isPublicEntityId(groupId)) {
		return {};
	}

	const afterGroup = rest.slice(groupSegment.length);
	if (!afterGroup || afterGroup === "/") {
		return { groupId };
	}

	const sub = afterGroup.startsWith("/") ? afterGroup.slice(1) : afterGroup;
	if (sub.startsWith("view/")) {
		const view = sub.slice("view/".length).split("/")[0] ?? "";
		if (isQuickViewId(view)) {
			return { groupId, quickView: view };
		}
		return { groupId };
	}

	const channelSegment = sub.split("/")[0] ?? "";
	if (isPublicEntityId(channelSegment)) {
		return { groupId, channelId: decodeURIComponent(channelSegment) };
	}

	return { groupId };
}
