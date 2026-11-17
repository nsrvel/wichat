"use client";

import { useCallback, useEffect, useLayoutEffect, useRef, useState, type RefObject } from "react";

type UseChatScrollerOptions = {
	channelId: string;
	messageCount: number;
	firstMessageId: string | undefined;
	lastMessageId: string | undefined;
	hasOlderMessages: boolean;
	isFetchingOlderMessages: boolean;
	fetchOlderMessages: () => void;
};

export type UseChatScrollerResult = {
	scrollRef: RefObject<HTMLDivElement | null>;
	contentRef: RefObject<HTMLDivElement | null>;
	showJumpToLatest: boolean;
	unreadNewCount: number;
	highlightedMessageId: string | null;
	handleScroll: () => void;
	scrollToBottom: (behavior?: ScrollBehavior) => void;
	scrollToMessage: (messageId: string) => boolean;
	notifyIncomingMessage: (isMine: boolean) => void;
	prepareForSend: () => void;
};

const BOTTOM_THRESHOLD_PX = 64;
const TOP_LOAD_THRESHOLD_PX = 160;

export function useChatScroller({
	channelId,
	messageCount,
	firstMessageId,
	lastMessageId,
	hasOlderMessages,
	isFetchingOlderMessages,
	fetchOlderMessages,
}: UseChatScrollerOptions): UseChatScrollerResult {
	const scrollRef = useRef<HTMLDivElement>(null);
	const contentRef = useRef<HTMLDivElement>(null);

	const isAtBottomRef = useRef(true);
	const forceBottomRef = useRef(true);
	const isSmoothScrollingRef = useRef(false);
	const smoothScrollTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
	const prependBottomOffsetRef = useRef<number | null>(null);
	const prevChannelIdRef = useRef<string>(channelId);
	const prevFirstIdRef = useRef<string | undefined>(firstMessageId);
	const prevLastIdRef = useRef<string | undefined>(lastMessageId);

	const [showJumpToLatest, setShowJumpToLatest] = useState(false);
	const [unreadNewCount, setUnreadNewCount] = useState(0);
	const [highlightedMessageId, setHighlightedMessageId] = useState<string | null>(null);
	const highlightTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

	const snapToBottom = useCallback((behavior: ScrollBehavior = "auto") => {
		const el = scrollRef.current;
		if (!el) {
			return;
		}
		const targetTop = Math.max(0, el.scrollHeight - el.clientHeight);
		if (behavior === "smooth") {
			isSmoothScrollingRef.current = true;
			if (smoothScrollTimerRef.current) {
				clearTimeout(smoothScrollTimerRef.current);
			}
			smoothScrollTimerRef.current = setTimeout(() => {
				isSmoothScrollingRef.current = false;
				smoothScrollTimerRef.current = null;
			}, 650);
			el.scrollTo({ top: targetTop, behavior: "smooth" });
		} else {
			isSmoothScrollingRef.current = false;
			if (smoothScrollTimerRef.current) {
				clearTimeout(smoothScrollTimerRef.current);
				smoothScrollTimerRef.current = null;
			}
			el.scrollTop = targetTop;
		}
	}, []);

	// Reset state on channel switch.
	useLayoutEffect(() => {
		if (prevChannelIdRef.current !== channelId) {
			prevChannelIdRef.current = channelId;
			prevFirstIdRef.current = undefined;
			prevLastIdRef.current = undefined;
			prependBottomOffsetRef.current = null;
			isAtBottomRef.current = true;
			forceBottomRef.current = true;
			setShowJumpToLatest(false);
			setUnreadNewCount(0);
			setHighlightedMessageId(null);
			if (highlightTimerRef.current) {
				clearTimeout(highlightTimerRef.current);
				highlightTimerRef.current = null;
			}
		}
	}, [channelId]);

	// Deterministic scroll positioning on message list updates (before paint).
	useLayoutEffect(() => {
		const el = scrollRef.current;
		if (!el) {
			return;
		}

		// 1. Prepend anchoring when older messages are loaded at top.
		if (
			prependBottomOffsetRef.current !== null &&
			firstMessageId !== undefined &&
			firstMessageId !== prevFirstIdRef.current
		) {
			el.scrollTop = Math.max(0, el.scrollHeight - prependBottomOffsetRef.current);
			prependBottomOffsetRef.current = null;
			prevFirstIdRef.current = firstMessageId;
			prevLastIdRef.current = lastMessageId;
			return;
		}

		// 2. Initial channel load or forced snap (e.g. user sent message).
		if (forceBottomRef.current && messageCount > 0) {
			forceBottomRef.current = false;
			isAtBottomRef.current = true;
			setShowJumpToLatest(false);
			setUnreadNewCount(0);
			snapToBottom("auto");
		} else if (lastMessageId !== prevLastIdRef.current && isAtBottomRef.current) {
			// 3. New message appended while user is locked at bottom.
			snapToBottom("auto");
		}

		prevFirstIdRef.current = firstMessageId;
		prevLastIdRef.current = lastMessageId;
	}, [messageCount, firstMessageId, lastMessageId, snapToBottom]);

	// Keep bottom locked whenever inner content resizes (images, code blocks, AI streaming).
	useEffect(() => {
		const contentEl = contentRef.current;
		if (!contentEl || typeof ResizeObserver === "undefined") {
			return;
		}
		const observer = new ResizeObserver(() => {
			if (isAtBottomRef.current) {
				snapToBottom(isSmoothScrollingRef.current ? "smooth" : "auto");
			}
		});
		observer.observe(contentEl);
		return () => observer.disconnect();
	}, [snapToBottom]);

	const handleScroll = useCallback(() => {
		const el = scrollRef.current;
		if (!el) {
			return;
		}

		const distanceFromBottom = el.scrollHeight - el.scrollTop - el.clientHeight;
		const atBottom = distanceFromBottom <= BOTTOM_THRESHOLD_PX;
		isAtBottomRef.current = atBottom;

		if (atBottom) {
			setShowJumpToLatest(false);
			setUnreadNewCount(0);
		} else if (distanceFromBottom > 240) {
			setShowJumpToLatest(true);
		}

		if (
			el.scrollTop <= TOP_LOAD_THRESHOLD_PX &&
			hasOlderMessages &&
			!isFetchingOlderMessages &&
			prependBottomOffsetRef.current === null
		) {
			prependBottomOffsetRef.current = el.scrollHeight - el.scrollTop;
			fetchOlderMessages();
		}
	}, [hasOlderMessages, isFetchingOlderMessages, fetchOlderMessages]);

	// Clear stale prepend offset if fetch finishes without adding older items.
	useEffect(() => {
		if (!isFetchingOlderMessages && prependBottomOffsetRef.current !== null) {
			const id = requestAnimationFrame(() => {
				prependBottomOffsetRef.current = null;
			});
			return () => cancelAnimationFrame(id);
		}
	}, [isFetchingOlderMessages]);

	const scrollToBottom = useCallback(
		(behavior: ScrollBehavior = "smooth") => {
			isAtBottomRef.current = true;
			setShowJumpToLatest(false);
			setUnreadNewCount(0);
			snapToBottom(behavior);
		},
		[snapToBottom],
	);

	const prepareForSend = useCallback(() => {
		forceBottomRef.current = true;
		isAtBottomRef.current = true;
		setShowJumpToLatest(false);
		setUnreadNewCount(0);
	}, []);

	const notifyIncomingMessage = useCallback((isMine: boolean) => {
		if (isMine) {
			forceBottomRef.current = true;
			isAtBottomRef.current = true;
			setShowJumpToLatest(false);
			setUnreadNewCount(0);
			return;
		}
		if (!isAtBottomRef.current) {
			setShowJumpToLatest(true);
			setUnreadNewCount((prev) => prev + 1);
		}
	}, []);

	useEffect(() => {
		return () => {
			if (highlightTimerRef.current) {
				clearTimeout(highlightTimerRef.current);
			}
			if (smoothScrollTimerRef.current) {
				clearTimeout(smoothScrollTimerRef.current);
			}
		};
	}, []);

	const scrollToMessage = useCallback((messageId: string): boolean => {
		const container = scrollRef.current;
		if (!container || !messageId) {
			return false;
		}
		const selector = `[data-message-id="${CSS.escape(messageId)}"]`;
		const target = container.querySelector(selector) as HTMLElement | null;
		if (!target) {
			return false;
		}
		target.scrollIntoView({ behavior: "smooth", block: "center" });
		setHighlightedMessageId(messageId);
		if (highlightTimerRef.current) {
			clearTimeout(highlightTimerRef.current);
		}
		highlightTimerRef.current = setTimeout(() => {
			setHighlightedMessageId((current) => (current === messageId ? null : current));
		}, 1600);
		return true;
	}, []);

	return {
		scrollRef,
		contentRef,
		showJumpToLatest,
		unreadNewCount,
		highlightedMessageId,
		handleScroll,
		scrollToBottom,
		scrollToMessage,
		notifyIncomingMessage,
		prepareForSend,
	};
}
