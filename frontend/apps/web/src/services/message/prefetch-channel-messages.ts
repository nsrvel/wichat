import type { QueryClient } from "@tanstack/react-query";
import { isPublicEntityId } from "@wichat/contracts/common";
import { channelMessagesQueryOpts } from "@/services/message/message.query";

const MAX_IN_FLIGHT = 2;
let inFlight = 0;
const waiters: Array<() => void> = [];

function drainQueue(): void {
	while (inFlight < MAX_IN_FLIGHT && waiters.length > 0) {
		const next = waiters.shift();
		if (!next) {
			break;
		}
		next();
	}
}

export type PrefetchChannelMessagesOptions = {
	/** Navigation click / pointerdown — runs immediately, not behind hover queue. */
	priority?: boolean;
};

function runPrefetch(queryClient: QueryClient, groupId: string, channelId: string): void {
	const opts = channelMessagesQueryOpts(groupId, channelId);
	const existing = queryClient.getQueryData(opts.queryKey);
	if (existing) {
		return;
	}

	const state = queryClient.getQueryState(opts.queryKey);
	if (state?.fetchStatus === "fetching") {
		return;
	}

	inFlight += 1;
	void queryClient.prefetchInfiniteQuery(opts).finally(() => {
		inFlight -= 1;
		drainQueue();
	});
}

/**
 * Best-effort first-page prefetch for a channel (hover / intent). Skips when data is already cached.
 * Concurrency capped so hover-spam does not crowd out the active channel fetch.
 */
export function prefetchChannelMessages(
	queryClient: QueryClient,
	groupId: string,
	channelId: string,
	options?: PrefetchChannelMessagesOptions,
): void {
	if (!isPublicEntityId(groupId) || !isPublicEntityId(channelId)) {
		return;
	}

	if (options?.priority) {
		runPrefetch(queryClient, groupId, channelId);
		return;
	}

	const opts = channelMessagesQueryOpts(groupId, channelId);
	const existing = queryClient.getQueryData(opts.queryKey);
	if (existing) {
		return;
	}

	const state = queryClient.getQueryState(opts.queryKey);
	if (state?.fetchStatus === "fetching") {
		return;
	}

	waiters.push(() => {
		runPrefetch(queryClient, groupId, channelId);
	});
	drainQueue();
}
