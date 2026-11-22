"use client";

import { useVirtualizer, type Virtualizer } from "@tanstack/react-virtual";
import { useCallback, useEffect, useLayoutEffect, useRef, useState, type RefObject } from "react";
import type { TimelineItem } from "@/features/chat/lib/group-timeline-items";

type ScrollMode = "sticky-bottom" | "free-scroll" | "jump-target";

type SavedChannelScroll = { mode: "sticky-bottom" } | { mode: "free-scroll"; scrollTop: number };

const channelScrollMemory = new Map<string, SavedChannelScroll>();
const MAX_SCROLL_MEMORY_ENTRIES = 50;

function saveChannelScroll(channelId: string, state: SavedChannelScroll) {
	if (channelScrollMemory.has(channelId)) {
		channelScrollMemory.delete(channelId);
	}
	channelScrollMemory.set(channelId, state);
	if (channelScrollMemory.size > MAX_SCROLL_MEMORY_ENTRIES) {
		const oldestKey = channelScrollMemory.keys().next().value;
		if (oldestKey) {
			channelScrollMemory.delete(oldestKey);
		}
	}
}

type UseChatScrollerOptions = {
	channelId: string;
	items: TimelineItem[];
	hasOlderMessages: boolean;
	isFetchingOlderMessages: boolean;
	fetchOlderMessages: () => void;
};

export type UseChatScrollerResult = {
	scrollRef: RefObject<HTMLDivElement | null>;
	contentRef: RefObject<HTMLDivElement | null>;
	virtualizer: Virtualizer<HTMLDivElement, HTMLDivElement>;
	showJumpToLatest: boolean;
	unreadNewCount: number;
	highlightedMessageId: string | null;
	handleScroll: () => void;
	scrollToBottom: (behavior?: ScrollBehavior) => void;
	scrollToMessage: (messageId: string) => boolean;
	notifyIncomingMessage: (isMine: boolean) => void;
	prepareForSend: () => void;
};

const BOTTOM_THRESHOLD_PX = 56;
const JUMP_PILL_THRESHOLD_PX = 240;
const TOP_LOAD_THRESHOLD_PX = 200;

function estimateTimelineItemSize(item: TimelineItem | undefined): number {
	if (!item) {
		return 76;
	}
	if (item.kind === "date-divider") {
		return 44;
	}
	if (item.message.bodyPlain.includes("```wichat-attachment\n")) {
		return 220;
	}
	if (item.message.bodyPlain.includes("```")) {
		return 140;
	}
	if (item.message.replyTo) {
		return 108;
	}
	return item.compact ? 44 : 76;
}

export function useChatScroller({
	channelId,
	items,
	hasOlderMessages,
	isFetchingOlderMessages,
	fetchOlderMessages,
}: UseChatScrollerOptions): UseChatScrollerResult {
	const scrollRef = useRef<HTMLDivElement>(null);
	const contentRef = useRef<HTMLDivElement>(null);

	const savedScroll = channelScrollMemory.get(channelId);
	const modeRef = useRef<ScrollMode>(savedScroll?.mode ?? "sticky-bottom");
	const initialRestoredRef = useRef(false);
	const prevFirstKeyRef = useRef<string | undefined>(undefined);
	const prevLastKeyRef = useRef<string | undefined>(undefined);
	const prevTotalSizeRef = useRef<number>(0);
	const prevScrollTopRef = useRef<number>(0);

	const [showJumpToLatest, setShowJumpToLatest] = useState(savedScroll?.mode === "free-scroll");
	const [unreadNewCount, setUnreadNewCount] = useState(0);
	const [highlightedMessageId, setHighlightedMessageId] = useState<string | null>(null);
	const highlightTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

	// eslint-disable-next-line react-hooks/incompatible-library -- TanStack Virtual returns non-memoizable imperative methods.
	const virtualizer = useVirtualizer<HTMLDivElement, HTMLDivElement>({
		count: items.length,
		getScrollElement: () => scrollRef.current,
		estimateSize: (index) => estimateTimelineItemSize(items[index]),
		getItemKey: (index) => items[index]?.key ?? index,
		overscan: 8,
	});

	const itemsRef = useRef(items);
	const paginationRef = useRef({ hasOlderMessages, isFetchingOlderMessages, fetchOlderMessages });
	const virtualizerRef = useRef(virtualizer);

	useLayoutEffect(() => {
		itemsRef.current = items;
		paginationRef.current = { hasOlderMessages, isFetchingOlderMessages, fetchOlderMessages };
		virtualizerRef.current = virtualizer;
	});

	const totalSize = virtualizer.getTotalSize();

	const snapToBottomNow = useCallback((behavior: ScrollBehavior = "auto") => {
		const el = scrollRef.current;
		if (!el) {
			return;
		}
		const targetTop = Math.max(0, el.scrollHeight - el.clientHeight);
		if (behavior === "smooth") {
			el.scrollTo({ top: targetTop, behavior: "smooth" });
		} else {
			el.scrollTop = targetTop;
		}
		prevScrollTopRef.current = el.scrollTop;
	}, []);

	// Persist scroll state on unmount (channel switch).
	useEffect(() => {
		const el = scrollRef.current;
		return () => {
			if (!el) {
				return;
			}
			if (modeRef.current === "sticky-bottom") {
				saveChannelScroll(channelId, { mode: "sticky-bottom" });
			} else {
				saveChannelScroll(channelId, { mode: "free-scroll", scrollTop: el.scrollTop });
			}
		};
	}, [channelId]);

	// User-intent interruption: release sticky-bottom immediately on upward wheel / touch drag so ResizeObserver never fights user scroll.
	useEffect(() => {
		const el = scrollRef.current;
		if (!el) {
			return;
		}
		let touchStartY: number | null = null;

		const onWheel = (event: WheelEvent) => {
			if (event.deltaY < 0) {
				modeRef.current = "free-scroll";
			} else if (modeRef.current === "jump-target" && event.deltaY !== 0) {
				modeRef.current = "free-scroll";
			}
		};

		const onTouchStart = (event: TouchEvent) => {
			touchStartY = event.touches[0]?.clientY ?? null;
			if (modeRef.current === "jump-target") {
				modeRef.current = "free-scroll";
			}
		};

		const onTouchMove = (event: TouchEvent) => {
			if (touchStartY === null) {
				return;
			}
			const currentY = event.touches[0]?.clientY;
			if (currentY !== undefined && currentY - touchStartY > 6) {
				modeRef.current = "free-scroll";
			}
		};

		const onPointerDown = () => {
			if (modeRef.current === "jump-target") {
				modeRef.current = "free-scroll";
			}
		};

		el.addEventListener("wheel", onWheel, { passive: true });
		el.addEventListener("touchstart", onTouchStart, { passive: true });
		el.addEventListener("touchmove", onTouchMove, { passive: true });
		el.addEventListener("pointerdown", onPointerDown, { passive: true });
		return () => {
			el.removeEventListener("wheel", onWheel);
			el.removeEventListener("touchstart", onTouchStart);
			el.removeEventListener("touchmove", onTouchMove);
			el.removeEventListener("pointerdown", onPointerDown);
		};
	}, []);

	// Deterministic scroll positioning before paint on timeline updates.
	useLayoutEffect(() => {
		const el = scrollRef.current;
		if (!el) {
			return;
		}

		const firstKey = items[0]?.key;
		const lastKey = items[items.length - 1]?.key;
		const prevFirstKey = prevFirstKeyRef.current;
		const prevLastKey = prevLastKeyRef.current;
		const prevTotalSize = prevTotalSizeRef.current;

		// 1. Initial channel mount with data: restore saved scroll position or snap to bottom.
		if (!initialRestoredRef.current && items.length > 0) {
			initialRestoredRef.current = true;
			const saved = channelScrollMemory.get(channelId);
			if (saved?.mode === "free-scroll") {
				modeRef.current = "free-scroll";
				el.scrollTop = Math.min(saved.scrollTop, Math.max(0, el.scrollHeight - el.clientHeight));
				prevScrollTopRef.current = el.scrollTop;
				setShowJumpToLatest(true);
			} else {
				modeRef.current = "sticky-bottom";
				setShowJumpToLatest(false);
				setUnreadNewCount(0);
				snapToBottomNow("auto");
			}
			prevFirstKeyRef.current = firstKey;
			prevLastKeyRef.current = lastKey;
			prevTotalSizeRef.current = totalSize;
			return;
		}

		// 2. Prepend anchoring when older messages are added at the top.
		if (
			prevFirstKey !== undefined &&
			firstKey !== undefined &&
			firstKey !== prevFirstKey &&
			lastKey === prevLastKey
		) {
			const delta = totalSize - prevTotalSize;
			if (delta > 0) {
				el.scrollTop += delta;
				prevScrollTopRef.current = el.scrollTop;
			}
			prevFirstKeyRef.current = firstKey;
			prevLastKeyRef.current = lastKey;
			prevTotalSizeRef.current = totalSize;
			return;
		}

		// 3. New message appended or size refined while locked to bottom.
		if (modeRef.current === "sticky-bottom") {
			snapToBottomNow("auto");
		}

		prevFirstKeyRef.current = firstKey;
		prevLastKeyRef.current = lastKey;
		prevTotalSizeRef.current = totalSize;
	}, [channelId, items, totalSize, snapToBottomNow]);

	// Keep bottom pinned when inner content resizes (e.g. async images, code highlight, embeds).
	useEffect(() => {
		const contentEl = contentRef.current;
		if (!contentEl || typeof ResizeObserver === "undefined") {
			return;
		}
		const observer = new ResizeObserver(() => {
			if (modeRef.current === "sticky-bottom") {
				snapToBottomNow("auto");
			}
		});
		observer.observe(contentEl);
		return () => observer.disconnect();
	}, [snapToBottomNow]);

	const handleScroll = useCallback(() => {
		const el = scrollRef.current;
		if (!el) {
			return;
		}

		const currentTop = el.scrollTop;
		const scrollingUp = currentTop < prevScrollTopRef.current - 1;
		prevScrollTopRef.current = currentTop;

		const distanceFromBottom = el.scrollHeight - currentTop - el.clientHeight;
		const atExactBottom = distanceFromBottom <= 8;
		const nearBottomMovingDown = !scrollingUp && distanceFromBottom <= BOTTOM_THRESHOLD_PX;

		if (atExactBottom || nearBottomMovingDown) {
			modeRef.current = "sticky-bottom";
			setShowJumpToLatest((prev) => (prev ? false : prev));
			setUnreadNewCount((prev) => (prev === 0 ? prev : 0));
		} else {
			if (scrollingUp && modeRef.current === "sticky-bottom") {
				modeRef.current = "free-scroll";
			} else if (distanceFromBottom > BOTTOM_THRESHOLD_PX && modeRef.current === "sticky-bottom") {
				modeRef.current = "free-scroll";
			}
			if (distanceFromBottom > JUMP_PILL_THRESHOLD_PX) {
				setShowJumpToLatest((prev) => (prev ? prev : true));
			}
		}

		const {
			hasOlderMessages: canLoadMore,
			isFetchingOlderMessages: busy,
			fetchOlderMessages: loadOlder,
		} = paginationRef.current;
		if (
			currentTop <= TOP_LOAD_THRESHOLD_PX &&
			canLoadMore &&
			!busy &&
			itemsRef.current.length > 0
		) {
			loadOlder();
		}
	}, []);

	const scrollToBottom = useCallback(
		(behavior: ScrollBehavior = "smooth") => {
			modeRef.current = "sticky-bottom";
			setShowJumpToLatest(false);
			setUnreadNewCount(0);
			const currentItems = itemsRef.current;
			if (currentItems.length > 0) {
				virtualizerRef.current.scrollToIndex(currentItems.length - 1, {
					align: "end",
					behavior: behavior === "smooth" ? "smooth" : "auto",
				});
			}
			snapToBottomNow(behavior);
		},
		[snapToBottomNow],
	);

	const prepareForSend = useCallback(() => {
		modeRef.current = "sticky-bottom";
		setShowJumpToLatest(false);
		setUnreadNewCount(0);
	}, []);

	const notifyIncomingMessage = useCallback(
		(isMine: boolean) => {
			if (isMine) {
				modeRef.current = "sticky-bottom";
				setShowJumpToLatest(false);
				setUnreadNewCount(0);
				snapToBottomNow("auto");
				return;
			}
			if (modeRef.current !== "sticky-bottom") {
				setShowJumpToLatest(true);
				setUnreadNewCount((prev) => prev + 1);
			}
		},
		[snapToBottomNow],
	);

	useEffect(() => {
		return () => {
			if (highlightTimerRef.current) {
				clearTimeout(highlightTimerRef.current);
			}
		};
	}, []);

	const scrollToMessage = useCallback((messageId: string): boolean => {
		if (!messageId) {
			return false;
		}
		const index = itemsRef.current.findIndex(
			(item) => item.kind === "message" && item.message.id === messageId,
		);
		if (index < 0) {
			return false;
		}
		modeRef.current = "jump-target";
		virtualizerRef.current.scrollToIndex(index, { align: "center", behavior: "smooth" });
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
		virtualizer,
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
