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

const BOTTOM_THRESHOLD_PX = 48;
const JUMP_BANNER_THRESHOLD_PX = 240;
const TOP_LOAD_THRESHOLD_PX = 140;

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

	const hasInitialSettledRef = useRef(false);
	const isAtBottomRef = useRef(true);
	const forceBottomRef = useRef(true);
	const isUserInteractingRef = useRef(false);
	const isJumpingToMessageRef = useRef(false);
	const isSmoothScrollingRef = useRef(false);

	const interactionTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
	const smoothScrollTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
	const jumpTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
	const highlightTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

	const prependAnchorRef = useRef<{
		id: string;
		topOffset: number;
		bottomOffset: number;
	} | null>(null);
	const prevChannelIdRef = useRef<string>(channelId);
	const prevFirstIdRef = useRef<string | undefined>(firstMessageId);
	const prevLastIdRef = useRef<string | undefined>(lastMessageId);

	const [showJumpToLatest, setShowJumpToLatest] = useState(false);
	const [unreadNewCount, setUnreadNewCount] = useState(0);
	const [highlightedMessageId, setHighlightedMessageId] = useState<string | null>(null);

	const snapToBottom = useCallback((behavior: ScrollBehavior = "auto") => {
		const el = scrollRef.current;
		if (!el) return;
		const targetTop = Math.max(0, el.scrollHeight - el.clientHeight);
		if (behavior === "smooth") {
			isSmoothScrollingRef.current = true;
			if (smoothScrollTimerRef.current) clearTimeout(smoothScrollTimerRef.current);
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

	// Reset state on channel switch (defensive even when keyed by channelId).
	useLayoutEffect(() => {
		if (prevChannelIdRef.current !== channelId) {
			prevChannelIdRef.current = channelId;
			prevFirstIdRef.current = undefined;
			prevLastIdRef.current = undefined;
			prependAnchorRef.current = null;
			hasInitialSettledRef.current = false;
			isAtBottomRef.current = true;
			forceBottomRef.current = true;
			isUserInteractingRef.current = false;
			isJumpingToMessageRef.current = false;
			setShowJumpToLatest(false);
			setUnreadNewCount(0);
			setHighlightedMessageId(null);
		}
	}, [channelId]);

	// Track explicit user scroll/pointer/wheel interaction so ResizeObserver never fights the user.
	useEffect(() => {
		const el = scrollRef.current;
		if (!el) return;

		const markUserInteracting = () => {
			isUserInteractingRef.current = true;
			isJumpingToMessageRef.current = false;
			if (interactionTimerRef.current) clearTimeout(interactionTimerRef.current);
			interactionTimerRef.current = setTimeout(() => {
				isUserInteractingRef.current = false;
				interactionTimerRef.current = null;
			}, 300);
		};

		const onWheel = (e: WheelEvent) => {
			markUserInteracting();
			if (e.deltaY < 0) {
				isAtBottomRef.current = false;
				forceBottomRef.current = false;
			}
		};

		el.addEventListener("wheel", onWheel, { passive: true });
		el.addEventListener("touchstart", markUserInteracting, { passive: true });
		el.addEventListener("pointerdown", markUserInteracting, { passive: true });
		return () => {
			el.removeEventListener("wheel", onWheel);
			el.removeEventListener("touchstart", markUserInteracting);
			el.removeEventListener("pointerdown", markUserInteracting);
		};
	}, []);

	// Deterministic scroll positioning on message list updates (before paint).
	useLayoutEffect(() => {
		const el = scrollRef.current;
		if (!el) return;

		// 1. Prepend anchoring when older messages are loaded at top.
		const anchor = prependAnchorRef.current;
		if (anchor && firstMessageId !== undefined && firstMessageId !== prevFirstIdRef.current) {
			const selector = `[data-message-id="${CSS.escape(anchor.id)}"]`;
			const anchorEl = el.querySelector(selector) as HTMLElement | null;
			if (anchorEl) {
				const containerTop = el.getBoundingClientRect().top;
				const newTopOffset = anchorEl.getBoundingClientRect().top - containerTop;
				el.scrollTop = Math.max(0, el.scrollTop + (newTopOffset - anchor.topOffset));
			} else {
				el.scrollTop = Math.max(0, el.scrollHeight - anchor.bottomOffset);
			}
			prependAnchorRef.current = null;
			prevFirstIdRef.current = firstMessageId;
			prevLastIdRef.current = lastMessageId;
			return;
		}

		// 2. Initial channel load or forced snap (e.g. user sent message).
		if (forceBottomRef.current && messageCount > 0) {
			forceBottomRef.current = false;
			hasInitialSettledRef.current = true;
			isAtBottomRef.current = true;
			setShowJumpToLatest(false);
			setUnreadNewCount(0);
			snapToBottom("auto");
		} else if (
			hasInitialSettledRef.current &&
			lastMessageId !== undefined &&
			lastMessageId !== prevLastIdRef.current &&
			isAtBottomRef.current &&
			!isUserInteractingRef.current &&
			!isJumpingToMessageRef.current
		) {
			// 3. New message appended while user is genuinely at bottom.
			snapToBottom("auto");
		}

		prevFirstIdRef.current = firstMessageId;
		prevLastIdRef.current = lastMessageId;
	}, [messageCount, firstMessageId, lastMessageId, snapToBottom]);

	// Keep bottom locked on image/codeblock resize ONLY if user is strictly at bottom and not interacting/jumping.
	useEffect(() => {
		const contentEl = contentRef.current;
		const scrollEl = scrollRef.current;
		if (!contentEl || !scrollEl || typeof ResizeObserver === "undefined") {
			return;
		}
		const observer = new ResizeObserver(() => {
			if (
				!hasInitialSettledRef.current ||
				!isAtBottomRef.current ||
				isUserInteractingRef.current ||
				isJumpingToMessageRef.current
			) {
				return;
			}
			snapToBottom(isSmoothScrollingRef.current ? "smooth" : "auto");
		});
		observer.observe(contentEl);
		return () => observer.disconnect();
	}, [snapToBottom]);

	const handleScroll = useCallback(() => {
		const el = scrollRef.current;
		if (!el || !hasInitialSettledRef.current) {
			return;
		}

		const distanceFromBottom = el.scrollHeight - el.scrollTop - el.clientHeight;
		const atBottom = distanceFromBottom <= BOTTOM_THRESHOLD_PX;
		isAtBottomRef.current = atBottom;

		if (atBottom) {
			setShowJumpToLatest(false);
			setUnreadNewCount(0);
		} else if (distanceFromBottom > JUMP_BANNER_THRESHOLD_PX) {
			setShowJumpToLatest(true);
		}

		if (
			el.scrollTop <= TOP_LOAD_THRESHOLD_PX &&
			hasOlderMessages &&
			!isFetchingOlderMessages &&
			prependAnchorRef.current === null &&
			firstMessageId
		) {
			const selector = `[data-message-id="${CSS.escape(firstMessageId)}"]`;
			const firstEl = el.querySelector(selector) as HTMLElement | null;
			const topOffset = firstEl
				? firstEl.getBoundingClientRect().top - el.getBoundingClientRect().top
				: 0;
			prependAnchorRef.current = {
				id: firstMessageId,
				topOffset,
				bottomOffset: el.scrollHeight - el.scrollTop,
			};
			fetchOlderMessages();
		}
	}, [firstMessageId, hasOlderMessages, isFetchingOlderMessages, fetchOlderMessages]);

	// Clear stale prepend anchor if fetch finishes without adding older items.
	useEffect(() => {
		if (!isFetchingOlderMessages && prependAnchorRef.current !== null) {
			const id = requestAnimationFrame(() => {
				prependAnchorRef.current = null;
			});
			return () => cancelAnimationFrame(id);
		}
	}, [isFetchingOlderMessages]);

	const scrollToBottom = useCallback(
		(behavior: ScrollBehavior = "smooth") => {
			isJumpingToMessageRef.current = false;
			isUserInteractingRef.current = false;
			isAtBottomRef.current = true;
			setShowJumpToLatest(false);
			setUnreadNewCount(0);
			snapToBottom(behavior);
		},
		[snapToBottom],
	);

	const prepareForSend = useCallback(() => {
		isJumpingToMessageRef.current = false;
		isUserInteractingRef.current = false;
		forceBottomRef.current = true;
		isAtBottomRef.current = true;
		setShowJumpToLatest(false);
		setUnreadNewCount(0);
	}, []);

	const notifyIncomingMessage = useCallback((isMine: boolean) => {
		if (isMine) {
			isJumpingToMessageRef.current = false;
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
			if (highlightTimerRef.current) clearTimeout(highlightTimerRef.current);
			if (smoothScrollTimerRef.current) clearTimeout(smoothScrollTimerRef.current);
			if (interactionTimerRef.current) clearTimeout(interactionTimerRef.current);
			if (jumpTimerRef.current) clearTimeout(jumpTimerRef.current);
		};
	}, []);

	const scrollToMessage = useCallback((messageId: string): boolean => {
		const container = scrollRef.current;
		if (!container || !messageId) return false;
		const selector = `[data-message-id="${CSS.escape(messageId)}"]`;
		const target = container.querySelector(selector) as HTMLElement | null;
		if (!target) return false;

		isAtBottomRef.current = false;
		forceBottomRef.current = false;
		isJumpingToMessageRef.current = true;
		if (jumpTimerRef.current) clearTimeout(jumpTimerRef.current);
		jumpTimerRef.current = setTimeout(() => {
			isJumpingToMessageRef.current = false;
			jumpTimerRef.current = null;
		}, 900);

		target.scrollIntoView({ behavior: "smooth", block: "center" });
		setHighlightedMessageId(messageId);
		if (highlightTimerRef.current) clearTimeout(highlightTimerRef.current);
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
