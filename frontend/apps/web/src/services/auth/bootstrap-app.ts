import type { User } from "@wichat/contracts/user";
import type { QueryClient } from "@tanstack/react-query";
import { isPublicEntityId } from "@wichat/contracts/common";
import { resolveChannelIconSrc } from "@/features/channel/lib/channel-preset-icon";
import { parseChannelRoute } from "@/features/shell/lib/parse-channel-route";
import { ensureAccessToken } from "@/integrations/api-client";
import { resolveGroupLogoUrl } from "@/lib/group-logo-url";
import { readLastChannel } from "@/lib/last-channel-storage";
import { preloadImages } from "@/lib/preload-images";
import { resolveAppDestination } from "@/lib/resolve-app-destination";
import * as authQuery from "@/services/auth/auth.query";
import * as channelQuery from "@/services/channel/channel.query";
import { myGroupsQueryOpts } from "@/services/group/group.query";
import { groupMembersQueryOpts } from "@/services/invite/invite.query";
import { channelMessagesQueryOpts } from "@/services/message/message.query";
import { myPresenceQueryOptions } from "@/services/user/user.query";

export const Q_APP_BOOTSTRAP = "app-bootstrap";

async function prefetchWorkspaceSlice(
	queryClient: QueryClient,
	groupId?: string,
	channelId?: string,
): Promise<void> {
	if (!groupId || !isPublicEntityId(groupId)) {
		return;
	}

	const tasks: Promise<unknown>[] = [
		queryClient.prefetchQuery(channelQuery.groupChannelsQueryOpts(groupId)),
		queryClient.prefetchQuery(groupMembersQueryOpts(groupId, "")),
	];

	if (channelId && isPublicEntityId(channelId)) {
		tasks.push(queryClient.prefetchInfiniteQuery(channelMessagesQueryOpts(groupId, channelId)));
	}

	await Promise.all(tasks);
}

/**
 * Full app + workspace preload: ensures auth token, fetches session, groups, presence,
 * and preloads the active (or destination) group's channels, members, and initial messages
 * in parallel so the UI is ready immediately when WorkspaceLoader dismisses.
 */
export async function bootstrapApp(
	queryClient: QueryClient,
	pathname?: string,
): Promise<User | null> {
	const token = await ensureAccessToken();
	if (!token) {
		return null;
	}

	const routeTarget = pathname ? parseChannelRoute(pathname) : {};

	const [user, groups] = await Promise.all([
		queryClient.fetchQuery(authQuery.sessionQueryOpts()),
		queryClient.fetchQuery(myGroupsQueryOpts()),
		queryClient.prefetchQuery(myPresenceQueryOptions()),
		prefetchWorkspaceSlice(queryClient, routeTarget.groupId, routeTarget.channelId),
	]);

	let activeGroupId = routeTarget.groupId;
	if (!activeGroupId && groups.length > 0) {
		const dest = resolveAppDestination(groups, readLastChannel());
		const destTarget = parseChannelRoute(dest);
		activeGroupId = destTarget.groupId;
		await prefetchWorkspaceSlice(queryClient, destTarget.groupId, destTarget.channelId);
	}

	// Priority 1: Critical chrome assets (own avatar + Group Rail logos using resolved URLs)
	const railImageUrls: (string | undefined)[] = [
		user?.avatar_url,
		...groups.map((g) => resolveGroupLogoUrl(g.logo_url)),
	];
	const railImagesPromise = preloadImages(railImageUrls, { timeoutMs: 2000, maxCount: 32 });

	// Priority 2: Secondary workspace assets (active group banner, channel icons, member avatars)
	const workspaceImageUrls: (string | undefined)[] = [];
	if (activeGroupId) {
		const activeGroup = groups.find((g) => g.id === activeGroupId);
		if (activeGroup?.banner_url) {
			workspaceImageUrls.push(resolveGroupLogoUrl(activeGroup.banner_url));
		}
		const channelsData = queryClient.getQueryData(
			channelQuery.groupChannelsQueryOpts(activeGroupId).queryKey,
		);
		if (channelsData?.channels) {
			for (const channel of channelsData.channels) {
				workspaceImageUrls.push(
					resolveChannelIconSrc({ type: channel.type, imageUrl: channel.image_url }),
				);
			}
		}
		const membersData = queryClient.getQueryData(groupMembersQueryOpts(activeGroupId, "").queryKey);
		if (membersData?.members) {
			for (const member of membersData.members) {
				workspaceImageUrls.push(member.avatar_url);
			}
		}
	}

	await Promise.all([
		railImagesPromise,
		preloadImages(workspaceImageUrls, { timeoutMs: 1500, maxCount: 32 }),
	]);

	return user;
}
