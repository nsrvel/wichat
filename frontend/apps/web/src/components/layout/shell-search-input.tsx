"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { usePathname, useRouter } from "next/navigation";
import {
	forwardRef,
	useCallback,
	useEffect,
	useId,
	useImperativeHandle,
	useLayoutEffect,
	useMemo,
	useRef,
	useState,
	type CSSProperties,
	type KeyboardEvent as ReactKeyboardEvent,
} from "react";
import { createPortal } from "react-dom";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { useChannelNavigationOptional } from "@/features/channel/channel-navigation-provider";
import type { ChannelMessagesInfiniteData } from "@/features/chat/lib/message-cache-window";
import { ShellSearchField } from "@/features/search/components/shell-search-field";
import { ShellSearchPanel } from "@/features/search/components/shell-search-panel";
import { filterGroupMembersForSearch } from "@/features/search/lib/filter-group-members";
import { dispatchJumpToMessage } from "@/features/search/lib/jump-to-message-event";
import {
	spotlightChannelMessages,
	type ShellMessageSearchHit,
} from "@/features/search/lib/shell-message-search-hit";
import {
	recentPeopleFromChannelCache,
	type RecentPersonHit,
} from "@/features/search/lib/recent-interaction-people";
import {
	clearRecentSearches,
	pushRecentSearch,
	readRecentSearches,
} from "@/features/search/lib/recent-search-storage";
import { filterGroupChannels } from "@/features/search/lib/filter-group-channels";
import {
	DEFAULT_SHELL_SEARCH_FILTER,
	isShellSearchFilterEnabled,
	type ShellSearchFilterId,
} from "@/features/search/lib/shell-search-filters";
import { Q_GROUP_CHANNELS } from "@/services/channel/channel.key";
import { groupChannelsQueryOpts } from "@/services/channel/channel.query";
import { ensureGroupDmChannel } from "@/services/channel/channel.service";
import { parseChannelRoute } from "@/features/shell/lib/parse-channel-route";
import { channelPath } from "@/lib/app-routes";
import { floatingCollisionPadding } from "@/lib/floating-position";
import { positionWithinViewport } from "@/lib/position-within-viewport";
import { cn } from "@/lib/utils";
import { sessionQueryOpts } from "@/services/auth/auth.query";
import type { GroupMember } from "@wichat/contracts/group/member";
import { Q_INVITE_MEMBERS } from "@/services/invite/invite.query";
import { listGroupMembers } from "@/services/invite/invite.service";
import { Q_CHANNEL_MESSAGES } from "@/services/message/message.key";
import { searchGroupFiles, searchGroupMessages } from "@/services/search/shell-search.service";
import { translateError } from "@/integrations/i18n/translate-error";

const OUTSIDE_DISMISS_GRACE_MS = 320;

type ShellSearchInputProps = {
	className?: string;
};

function measureShellSearchPanelStyle(anchor: HTMLElement, panelHeight: number): CSSProperties {
	const rect = anchor.getBoundingClientRect();
	const width = rect.width;
	const height =
		panelHeight > 0
			? panelHeight
			: Math.min(480, typeof window !== "undefined" ? window.innerHeight * 0.55 : 480);
	const { top, left, side } = positionWithinViewport({
		anchor: rect,
		width,
		height,
		align: "start",
		side: "bottom",
		gap: 0,
	});

	return {
		position: "fixed",
		top,
		left,
		width,
		zIndex: 90,
		maxHeight: side === "top" ? Math.max(200, rect.top - floatingCollisionPadding) : undefined,
	};
}

export const ShellSearchInput = forwardRef<HTMLInputElement, ShellSearchInputProps>(
	function ShellSearchInput({ className }, forwardedRef) {
		const { t } = useTranslation();
		const router = useRouter();
		const pathname = usePathname();
		const queryClient = useQueryClient();
		const nav = useChannelNavigationOptional();
		const searchInputId = useId();
		const panelId = useId();

		const anchorRef = useRef<HTMLDivElement>(null);
		const panelRef = useRef<HTMLDivElement>(null);
		const inputRef = useRef<HTMLInputElement>(null);
		const outsideGraceUntilRef = useRef(0);

		useImperativeHandle(forwardedRef, () => inputRef.current as HTMLInputElement);

		const route = useMemo(() => parseChannelRoute(pathname), [pathname]);
		const groupId = nav?.groupId ?? route.groupId;
		const channelId = nav?.channelId ?? route.channelId;

		const [mounted, setMounted] = useState(false);
		const [open, setOpen] = useState(false);
		const [panelStyle, setPanelStyle] = useState<CSSProperties>({});
		const [query, setQuery] = useState("");
		const [debouncedQuery, setDebouncedQuery] = useState("");
		const [filter, setFilter] = useState<ShellSearchFilterId>(DEFAULT_SHELL_SEARCH_FILTER);
		const [recentQueries, setRecentQueries] = useState<string[]>([]);

		const { data: sessionUser } = useQuery(sessionQueryOpts());
		const channelsQuery = useQuery({
			...groupChannelsQueryOpts(groupId ?? ""),
			enabled: Boolean(groupId) && open,
		});

		const membersQuery = useQuery({
			queryKey: [Q_INVITE_MEMBERS, groupId, "shell-search"],
			queryFn: () => listGroupMembers(groupId ?? "", { limit: 200 }),
			enabled: Boolean(groupId) && open,
			staleTime: 30_000,
		});

		const messageCache = useMemo(() => {
			if (!groupId || !channelId) {
				return undefined;
			}
			return queryClient.getQueryData<ChannelMessagesInfiniteData>([
				Q_CHANNEL_MESSAGES,
				groupId,
				channelId,
			]);
		}, [channelId, groupId, queryClient]);

		const memberById = useMemo(() => {
			const map = new Map<string, GroupMember>();
			for (const member of membersQuery.data?.members ?? []) {
				map.set(member.user_id, member);
			}
			return map;
		}, [membersQuery.data?.members]);

		const recentPeople = useMemo(
			() => recentPeopleFromChannelCache(messageCache, memberById, sessionUser?.id ?? undefined),
			[memberById, messageCache, sessionUser?.id],
		);

		const messageSearchQuery = debouncedQuery.trim();
		const messageSearch = useQuery({
			queryKey: ["shell-search", "messages", groupId ?? "", messageSearchQuery],
			queryFn: () => searchGroupMessages(groupId ?? "", messageSearchQuery, 8),
			enabled: open && filter === "messages" && Boolean(groupId) && messageSearchQuery.length >= 2,
			staleTime: 15_000,
			retry: 0,
		});

		const messageHits = useMemo((): ShellMessageSearchHit[] => {
			if (filter !== "messages" || !messageSearchQuery) {
				return [];
			}
			if (messageSearch.data?.hits?.length) {
				return messageSearch.data.hits.map((h) => ({
					groupId: h.group_id,
					channelId: h.channel_id,
					messageId: h.message_id,
					bodyPreview: h.body_preview,
					authorDisplayName: h.author_display_name,
					createdAt: h.created_at,
				}));
			}
			if (messageSearch.isError && groupId && channelId) {
				return spotlightChannelMessages(messageCache, messageSearchQuery, groupId, channelId);
			}
			return [];
		}, [
			channelId,
			filter,
			groupId,
			messageCache,
			messageSearch.data?.hits,
			messageSearch.isError,
			messageSearchQuery,
		]);
		const messageLoading =
			filter === "messages" && messageSearchQuery.length >= 2 && messageSearch.isFetching;

		const fileSearchQuery = debouncedQuery.trim();
		const fileSearch = useQuery({
			queryKey: ["shell-search", "files", groupId ?? "", filter, fileSearchQuery],
			queryFn: () =>
				searchGroupFiles(groupId ?? "", fileSearchQuery, {
					limit: 8,
					imagesOnly: filter === "images",
				}),
			enabled:
				open &&
				(filter === "files" || filter === "images") &&
				Boolean(groupId) &&
				fileSearchQuery.length >= 2,
			staleTime: 15_000,
			retry: 0,
		});
		const fileHits = filter === "files" || filter === "images" ? (fileSearch.data?.hits ?? []) : [];
		const fileLoading =
			(filter === "files" || filter === "images") &&
			fileSearchQuery.length >= 2 &&
			fileSearch.isFetching;

		const channelHits = useMemo(
			() =>
				filter === "channels" && groupId
					? filterGroupChannels(channelsQuery.data?.channels, groupId, debouncedQuery)
					: [],
			[channelsQuery.data?.channels, debouncedQuery, filter, groupId],
		);

		const peopleSearchQuery = debouncedQuery.trim();
		const peopleHits = useMemo(
			() =>
				filterGroupMembersForSearch(membersQuery.data?.members, peopleSearchQuery, sessionUser?.id),
			[membersQuery.data?.members, peopleSearchQuery, sessionUser?.id],
		);
		const peopleLoading = membersQuery.isFetching && peopleSearchQuery.length >= 2;

		useEffect(() => {
			setMounted(true);
		}, []);

		useEffect(() => {
			const timer = window.setTimeout(() => setDebouncedQuery(query), 280);
			return () => window.clearTimeout(timer);
		}, [query]);

		const closeSearch = useCallback((recordQuery = true) => {
			setQuery((current) => {
				if (recordQuery) {
					const trimmed = current.trim();
					if (trimmed) {
						setRecentQueries(pushRecentSearch(trimmed));
					}
				}
				return "";
			});
			setDebouncedQuery("");
			setOpen(false);
		}, []);

		const openSearch = useCallback(() => {
			if (!groupId) {
				toast.info(t("shell.search.needs_group"));
				return;
			}
			outsideGraceUntilRef.current = performance.now() + OUTSIDE_DISMISS_GRACE_MS;
			setRecentQueries(readRecentSearches());
			setOpen(true);
		}, [groupId, t]);

		const syncPanelPosition = useCallback(() => {
			const anchor = anchorRef.current;
			if (!anchor) {
				return;
			}
			const panelHeight = panelRef.current?.offsetHeight ?? 0;
			setPanelStyle(measureShellSearchPanelStyle(anchor, panelHeight));
		}, []);

		useLayoutEffect(() => {
			if (!open) {
				return;
			}
			syncPanelPosition();
			inputRef.current?.focus({ preventScroll: true });
		}, [open, syncPanelPosition, query, filter, messageHits.length, peopleHits.length]);

		useEffect(() => {
			if (!open) {
				return;
			}
			syncPanelPosition();
			window.addEventListener("resize", syncPanelPosition);
			window.addEventListener("scroll", syncPanelPosition, true);
			return () => {
				window.removeEventListener("resize", syncPanelPosition);
				window.removeEventListener("scroll", syncPanelPosition, true);
			};
		}, [open, syncPanelPosition]);

		useEffect(() => {
			if (!open) {
				return;
			}
			const onPointerDown = (event: PointerEvent) => {
				if (performance.now() < outsideGraceUntilRef.current) {
					return;
				}
				const target = event.target as Node;
				if (anchorRef.current?.contains(target)) {
					return;
				}
				if (panelRef.current?.contains(target)) {
					return;
				}
				closeSearch(false);
			};
			document.addEventListener("pointerdown", onPointerDown, true);
			return () => document.removeEventListener("pointerdown", onPointerDown, true);
		}, [closeSearch, open]);

		useEffect(() => {
			const onKeyDown = (event: globalThis.KeyboardEvent) => {
				if (!(event.metaKey || event.ctrlKey) || !event.altKey) {
					return;
				}
				if (event.key.toLowerCase() !== "e") {
					return;
				}
				event.preventDefault();
				if (!groupId) {
					toast.info(t("shell.search.needs_group"));
					return;
				}
				openSearch();
			};
			window.addEventListener("keydown", onKeyDown);
			return () => window.removeEventListener("keydown", onKeyDown);
		}, [groupId, openSearch, t]);

		const handleFilterChange = useCallback(
			(id: ShellSearchFilterId) => {
				if (!isShellSearchFilterEnabled(id)) {
					toast.info(t("shell.search.filter_soon", { filter: t(`shell.search.filter_${id}`) }));
					return;
				}
				setFilter(id);
			},
			[t],
		);

		const navigateToChannel = useCallback(
			(gid: string, cid: string) => {
				if (nav) {
					nav.navigateToChannel(gid, cid);
					return;
				}
				router.push(channelPath(gid, cid));
			},
			[nav, router],
		);

		const handleSelectChannel = useCallback(
			(hit: { groupId: string; channelId: string }) => {
				closeSearch();
				navigateToChannel(hit.groupId, hit.channelId);
			},
			[closeSearch, navigateToChannel],
		);

		const handleSelectMessage = useCallback(
			(hit: ShellMessageSearchHit) => {
				closeSearch();
				navigateToChannel(hit.groupId, hit.channelId);
				window.setTimeout(() => dispatchJumpToMessage(hit.messageId), 120);
			},
			[closeSearch, navigateToChannel],
		);

		const ensureDm = useMutation({
			mutationFn: (userId: string) => ensureGroupDmChannel(groupId ?? "", userId),
			onSuccess: (data) => {
				if (groupId) {
					void queryClient.invalidateQueries({ queryKey: [Q_GROUP_CHANNELS, groupId] });
				}
				closeSearch();
				navigateToChannel(groupId ?? "", data.channel.id);
			},
			onError: (err) => toast.error(translateError(t, err)),
		});

		const handleSelectPerson = useCallback(
			(person: RecentPersonHit) => {
				if (!groupId || !person.userId) {
					return;
				}
				ensureDm.mutate(person.userId);
			},
			[ensureDm, groupId],
		);

		const handleSelectFile = useCallback(
			(hit: { group_id: string; channel_id: string; message_id: string }) => {
				closeSearch();
				navigateToChannel(hit.group_id, hit.channel_id);
				if (hit.message_id) {
					window.setTimeout(() => dispatchJumpToMessage(hit.message_id), 120);
				}
			},
			[closeSearch, navigateToChannel],
		);

		const handleRecentSelect = useCallback((item: string) => {
			setQuery(item);
			setDebouncedQuery(item);
			inputRef.current?.focus();
		}, []);

		const handleClearRecent = useCallback(() => {
			clearRecentSearches();
			setRecentQueries([]);
		}, []);

		const handleSubmit = useCallback(() => {
			const q = query.trim();
			if (!q) {
				return;
			}
			setRecentQueries(pushRecentSearch(q));
		}, [query]);

		const handleQueryKeyDown = (event: ReactKeyboardEvent<HTMLInputElement>) => {
			if (event.key === "Escape") {
				event.preventDefault();
				closeSearch(false);
				return;
			}
			if (event.key === "Enter") {
				event.preventDefault();
				handleSubmit();
			}
		};

		const panel =
			open && mounted ? (
				<div
					ref={panelRef}
					style={panelStyle}
					className="overflow-hidden rounded-b-md rounded-t-none border-0 bg-card text-card-foreground shadow-elevated outline-none"
				>
					<ShellSearchPanel
						panelId={panelId}
						query={query}
						filter={filter}
						onFilterChange={handleFilterChange}
						recentPeople={recentPeople}
						recentQueries={recentQueries}
						messageHits={messageHits}
						messageLoading={messageLoading}
						channelHits={channelHits}
						fileHits={fileHits}
						fileLoading={fileLoading}
						peopleHits={peopleHits}
						peopleLoading={peopleLoading}
						onSelectMessage={handleSelectMessage}
						onSelectChannel={handleSelectChannel}
						onSelectFile={handleSelectFile}
						onSelectPerson={handleSelectPerson}
						onRecentSelect={handleRecentSelect}
						onClearRecent={handleClearRecent}
					/>
				</div>
			) : null;

		return (
			<>
				<div ref={anchorRef} className={cn("relative w-full", open && "z-[91]", className)}>
					{!open ? (
						<button
							type="button"
							className="w-full rounded-md border-0 bg-transparent p-0 text-left shadow-none outline-none transition-colors hover:[&_span_span]:bg-[#ececed] focus-visible:ring-2 focus-visible:ring-ring/40 dark:hover:[&_span_span]:bg-[#2a2a2a]"
							aria-expanded={false}
							aria-controls={panelId}
							aria-haspopup="dialog"
							aria-label={t("shell.search.open_label")}
							onClick={openSearch}
						>
							<ShellSearchField mode="trigger" />
						</button>
					) : (
						<ShellSearchField
							ref={inputRef}
							mode="active"
							id={searchInputId}
							value={query}
							onValueChange={setQuery}
							onKeyDown={handleQueryKeyDown}
							aria-expanded
							aria-controls={panelId}
						/>
					)}
				</div>
				{panel && mounted ? createPortal(panel, document.body) : null}
			</>
		);
	},
);
