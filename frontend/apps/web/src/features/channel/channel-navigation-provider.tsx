"use client";

import { useQueryClient } from "@tanstack/react-query";
import { usePathname, useRouter } from "next/navigation";
import {
	createContext,
	useCallback,
	useContext,
	useEffect,
	useMemo,
	useState,
	useTransition,
	type ReactNode,
} from "react";
import { parseChannelRoute } from "@/features/shell/lib/parse-channel-route";
import { channelPath, quickViewPath, type QuickViewId } from "@/lib/app-routes";
import { prefetchChannelMessages } from "@/services/message/prefetch-channel-messages";

type PendingTarget =
	| { kind: "channel"; groupId: string; channelId: string }
	| { kind: "quickView"; groupId: string; quickView: QuickViewId };

type PendingEntry = {
	target: PendingTarget;
	fromPathname: string;
};

type ChannelNavigationContextValue = {
	groupId: string | undefined;
	channelId: string | undefined;
	quickView: QuickViewId | undefined;
	isNavigating: boolean;
	navigateToChannel: (groupId: string, channelId: string) => void;
	navigateToQuickView: (groupId: string, quickView: QuickViewId) => void;
};

const ChannelNavigationContext = createContext<ChannelNavigationContextValue | null>(null);

function matchesPending(pathname: string, pending: PendingTarget): boolean {
	const parsed = parseChannelRoute(pathname);
	if (pending.kind === "channel") {
		return parsed.groupId === pending.groupId && parsed.channelId === pending.channelId;
	}
	return parsed.groupId === pending.groupId && parsed.quickView === pending.quickView;
}

function targetToRoute(target: PendingTarget) {
	if (target.kind === "channel") {
		return {
			groupId: target.groupId,
			channelId: target.channelId,
			quickView: undefined as QuickViewId | undefined,
		};
	}
	return {
		groupId: target.groupId,
		channelId: undefined as string | undefined,
		quickView: target.quickView,
	};
}

function resolveEffectiveRoute(pathname: string, entry: PendingEntry | null) {
	const parsed = parseChannelRoute(pathname);
	if (!entry) {
		return {
			groupId: parsed.groupId,
			channelId: parsed.channelId,
			quickView: parsed.quickView,
		};
	}

	if (matchesPending(pathname, entry.target)) {
		return {
			groupId: parsed.groupId,
			channelId: parsed.channelId,
			quickView: parsed.quickView,
		};
	}

	if (pathname !== entry.fromPathname) {
		return {
			groupId: parsed.groupId,
			channelId: parsed.channelId,
			quickView: parsed.quickView,
		};
	}

	return targetToRoute(entry.target);
}

export function ChannelNavigationProvider({ children }: { children: ReactNode }) {
	const router = useRouter();
	const pathname = usePathname();
	const queryClient = useQueryClient();
	const [pending, setPending] = useState<PendingEntry | null>(null);
	const [isNavigating, startTransition] = useTransition();

	const parsed = useMemo(() => parseChannelRoute(pathname), [pathname]);

	const activePending = useMemo(() => {
		if (!pending) {
			return null;
		}
		if (matchesPending(pathname, pending.target)) {
			return null;
		}
		if (pathname !== pending.fromPathname) {
			return null;
		}
		return pending;
	}, [pathname, pending]);

	useEffect(() => {
		if (!pending) {
			return;
		}
		if (matchesPending(pathname, pending.target) || pathname !== pending.fromPathname) {
			// eslint-disable-next-line react-hooks/set-state-in-effect -- drop stale optimistic nav after URL sync or back/forward.
			setPending(null);
		}
	}, [pathname, pending]);

	const navigateToChannel = useCallback(
		(groupId: string, channelId: string) => {
			const href = channelPath(groupId, channelId);
			if (parsed.groupId === groupId && parsed.channelId === channelId && !parsed.quickView) {
				return;
			}
			setPending({
				target: { kind: "channel", groupId, channelId },
				fromPathname: pathname,
			});
			prefetchChannelMessages(queryClient, groupId, channelId, { priority: true });
			startTransition(() => {
				router.push(href);
			});
		},
		[parsed.channelId, parsed.groupId, parsed.quickView, pathname, queryClient, router],
	);

	const navigateToQuickView = useCallback(
		(groupId: string, quickView: QuickViewId) => {
			const href = quickViewPath(groupId, quickView);
			if (parsed.groupId === groupId && parsed.quickView === quickView) {
				return;
			}
			setPending({
				target: { kind: "quickView", groupId, quickView },
				fromPathname: pathname,
			});
			startTransition(() => {
				router.push(href);
			});
		},
		[parsed.groupId, parsed.quickView, pathname, router],
	);

	const { groupId, channelId, quickView } = useMemo(
		() => resolveEffectiveRoute(pathname, activePending),
		[activePending, pathname],
	);

	const value = useMemo(
		() => ({
			groupId,
			channelId,
			quickView,
			isNavigating,
			navigateToChannel,
			navigateToQuickView,
		}),
		[channelId, groupId, isNavigating, navigateToChannel, navigateToQuickView, quickView],
	);

	return (
		<ChannelNavigationContext.Provider value={value}>{children}</ChannelNavigationContext.Provider>
	);
}

export function useChannelNavigation(): ChannelNavigationContextValue {
	const ctx = useContext(ChannelNavigationContext);
	if (!ctx) {
		throw new Error("useChannelNavigation must be used within ChannelNavigationProvider");
	}
	return ctx;
}

export function useChannelNavigationOptional(): ChannelNavigationContextValue | null {
	return useContext(ChannelNavigationContext);
}
