import type { InfiniteData } from "@tanstack/react-query";
import type {
	Message,
	MessageReactionSummary,
	MessageReplyPreview,
	MessagesList,
} from "@wichat/contracts/chat/message";

export type ChannelMessageReaction = {
	emoji: string;
	count?: number;
	active?: boolean;
	userIds?: string[];
};

export type ChannelMessage = {
	id: string;
	authorId: string;
	authorDisplayName: string;
	authorAvatarUrl: string;
	createdAt: string;
	editedAt?: string;
	isPinned?: boolean;
	replyTo?: MessageReplyPreview | null;
	bodyPlain: string;
	mine?: boolean;
	reactions?: ChannelMessageReaction[];
	timeLabel?: string;
};

export function appendOrUpdateInfiniteChannelMessage(
	prev: InfiniteData<MessagesList> | undefined,
	incoming: Message,
	currentUserId?: string | null,
): InfiniteData<MessagesList> {
	const isMine = incoming.mine || (Boolean(currentUserId) && incoming.author_id === currentUserId);

	const normalized: Message = {
		...incoming,
		mine: isMine,
	};

	if (!prev || !prev.pages || prev.pages.length === 0) {
		return {
			pages: [{ messages: [normalized], has_more: false }],
			pageParams: [undefined],
		};
	}

	let found = false;
	const nextPages = prev.pages.map((page) => {
		const existingIndex = page.messages.findIndex((m) => m.id === incoming.id);
		if (existingIndex >= 0) {
			found = true;
			const updated = [...page.messages];
			const existing = page.messages[existingIndex];
			updated[existingIndex] = {
				...normalized,
				mine: existing.mine || normalized.mine,
			};
			return { ...page, messages: updated };
		}
		return page;
	});

	if (found) {
		return {
			...prev,
			pages: nextPages,
		};
	}

	// If this is our own message arriving via WS before HTTP onSuccess, replace the oldest matching optimistic placeholder.
	if (isMine) {
		for (let pIdx = 0; pIdx < nextPages.length; pIdx++) {
			const page = nextPages[pIdx];
			const optIdx = page.messages.findIndex(
				(m) => m.id.startsWith("opt_") && m.body === incoming.body && (m.mine || isMine),
			);
			if (optIdx >= 0) {
				const updated = [...page.messages];
				updated[optIdx] = normalized;
				const replacedPages = [...nextPages];
				replacedPages[pIdx] = { ...page, messages: updated };
				return { ...prev, pages: replacedPages };
			}
		}
	}

	const page0 = nextPages[0] ?? { messages: [] };
	const updatedPage0: MessagesList = {
		...page0,
		messages: [...page0.messages, normalized],
	};

	return {
		...prev,
		pages: [updatedPage0, ...nextPages.slice(1)],
	};
}

export function replaceOptimisticChannelMessage(
	prev: InfiniteData<MessagesList> | undefined,
	tempId: string,
	realMessage: Message,
	currentUserId?: string | null,
): InfiniteData<MessagesList> {
	const isMine =
		realMessage.mine || (Boolean(currentUserId) && realMessage.author_id === currentUserId);
	const normalized: Message = { ...realMessage, mine: isMine };

	if (!prev || !prev.pages || prev.pages.length === 0) {
		return {
			pages: [{ messages: [normalized], has_more: false }],
			pageParams: [undefined],
		};
	}

	// Check if WS already delivered the real message; if so, remove tempId and update realMessage in place.
	const hasReal = prev.pages.some((p) => p.messages.some((m) => m.id === normalized.id));
	let replacedTemp = false;

	const nextPages = prev.pages.map((page) => {
		const filtered = hasReal
			? page.messages.filter((m) => m.id !== tempId)
			: page.messages.map((m) => {
					if (m.id === tempId) {
						replacedTemp = true;
						return normalized;
					}
					return m;
				});

		if (hasReal) {
			return {
				...page,
				messages: filtered.map((m) =>
					m.id === normalized.id ? { ...normalized, mine: m.mine || normalized.mine } : m,
				),
			};
		}

		return { ...page, messages: filtered };
	});

	if (hasReal || replacedTemp) {
		return { ...prev, pages: nextPages };
	}

	const page0 = nextPages[0] ?? { messages: [] };
	return {
		...prev,
		pages: [{ ...page0, messages: [...page0.messages, normalized] }, ...nextPages.slice(1)],
	};
}

export function mergeLatestPageIntoInfiniteMessages(
	prev: InfiniteData<MessagesList> | undefined,
	latestPage: MessagesList,
): InfiniteData<MessagesList> {
	if (!prev || !prev.pages || prev.pages.length === 0) {
		return {
			pages: [latestPage],
			pageParams: [undefined],
		};
	}

	const olderIds = new Set<string>();
	for (let i = 1; i < prev.pages.length; i++) {
		for (const msg of prev.pages[i].messages) {
			olderIds.add(msg.id);
		}
	}

	const page0 = prev.pages[0];
	const inFlightOptimistic = page0.messages.filter((m) => m.id.startsWith("opt_"));
	const latestById = new Map(latestPage.messages.map((m) => [m.id, m]));

	// Preserve existing page0 messages that aren't in latestPage (e.g. if page0 grew beyond PAGE_SIZE)
	// and merge updated/new messages from latestPage in chronological order.
	const merged: Message[] = [];
	const seen = new Set<string>();

	for (const msg of page0.messages) {
		if (msg.id.startsWith("opt_") || olderIds.has(msg.id)) {
			continue;
		}
		const updated = latestById.get(msg.id);
		if (updated) {
			merged.push({ ...updated, mine: msg.mine || updated.mine });
			seen.add(msg.id);
		} else if (
			latestPage.messages.length === 0 ||
			msg.created_at < latestPage.messages[0].created_at
		) {
			merged.push(msg);
			seen.add(msg.id);
		}
	}

	for (const msg of latestPage.messages) {
		if (!seen.has(msg.id) && !olderIds.has(msg.id)) {
			merged.push(msg);
			seen.add(msg.id);
		}
	}

	for (const opt of inFlightOptimistic) {
		if (!seen.has(opt.id)) {
			merged.push(opt);
		}
	}

	return {
		...prev,
		pages: [
			{
				...page0,
				messages: merged,
				has_more: prev.pages.length === 1 ? latestPage.has_more : page0.has_more,
			},
			...prev.pages.slice(1),
		],
	};
}

export function removeInfiniteChannelMessage(
	prev: InfiniteData<MessagesList> | undefined,
	messageId: string,
): InfiniteData<MessagesList> | undefined {
	if (!prev || !prev.pages) {
		return prev;
	}
	return {
		...prev,
		pages: prev.pages.map((page) => ({
			...page,
			messages: page.messages.filter((m) => m.id !== messageId),
		})),
	};
}

export function updateInfiniteChannelMessageReactions(
	prev: InfiniteData<MessagesList> | undefined,
	messageId: string,
	reactions: MessageReactionSummary[],
): InfiniteData<MessagesList> | undefined {
	if (!prev || !prev.pages) {
		return prev;
	}
	return {
		...prev,
		pages: prev.pages.map((page) => ({
			...page,
			messages: page.messages.map((m) => (m.id === messageId ? { ...m, reactions } : m)),
		})),
	};
}

export function flattenInfiniteChannelMessages(
	data: InfiniteData<MessagesList> | undefined,
): Message[] {
	if (!data || !data.pages || data.pages.length === 0) {
		return [];
	}

	const reversedPages = [...data.pages].reverse();
	const seen = new Set<string>();
	const result: Message[] = [];

	for (const page of reversedPages) {
		for (const msg of page.messages) {
			if (!seen.has(msg.id)) {
				seen.add(msg.id);
				result.push(msg);
			}
		}
	}

	return result;
}

export function appendOrUpdateChannelMessage(
	prev: MessagesList | undefined,
	incoming: Message,
	currentUserId?: string | null,
): MessagesList {
	const isMine = incoming.mine || (Boolean(currentUserId) && incoming.author_id === currentUserId);

	const normalized: Message = {
		...incoming,
		mine: isMine,
	};

	if (!prev || !prev.messages) {
		return { messages: [normalized] };
	}

	const existingIndex = prev.messages.findIndex(
		(m) =>
			m.id === incoming.id ||
			(m.id.startsWith("opt_") && m.body === incoming.body && (m.mine || isMine)),
	);
	if (existingIndex >= 0) {
		const existing = prev.messages[existingIndex];
		const updated = [...prev.messages];
		updated[existingIndex] = {
			...normalized,
			mine: existing.mine || normalized.mine,
		};
		return { messages: updated };
	}

	return { messages: [...prev.messages, normalized] };
}
