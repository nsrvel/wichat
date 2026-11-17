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
