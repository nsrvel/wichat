import type { InfiniteData } from "@tanstack/react-query";
import type { Message, MessagesList } from "@wichat/contracts";

export type ChannelMessageReaction = {
	emoji: string;
	count?: number;
	active?: boolean;
};

export type ChannelMessage = {
	id: string;
	authorId: string;
	authorDisplayName: string;
	authorAvatarUrl: string;
	/** ISO-8601 from API later; dev fixtures may use display strings via `timeLabel`. */
	createdAt: string;
	bodyPlain: string;
	mine?: boolean;
	reactions?: ChannelMessageReaction[];
	/** Relative or formatted time for display when API formatting is not wired. */
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
			pages: [{ messages: [normalized] }],
			pageParams: [undefined],
		};
	}

	let found = false;
	const nextPages = prev.pages.map((page) => {
		const existingIndex = page.messages.findIndex(
			(m) =>
				m.id === incoming.id ||
				(m.id.startsWith("opt_") && m.body === incoming.body && (m.mine || isMine)),
		);
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

	// Match by real ID first, or by optimistic temp ID
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
