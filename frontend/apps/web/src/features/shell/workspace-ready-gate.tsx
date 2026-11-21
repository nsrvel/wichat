"use client";

import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { usePathname } from "next/navigation";
import { useEffect, useMemo, useRef, useState, type ReactNode } from "react";
import { WorkspaceLoader } from "@/components/layout/workspace-loader";
import { parseChannelRoute } from "@/features/shell/lib/parse-channel-route";
import { useMyGroupsQuery } from "@/hooks/use-my-groups-query";
import * as channelQuery from "@/services/channel/channel.query";
import { groupMembersQueryOpts } from "@/services/invite/invite.query";
import { channelMessagesQueryOpts } from "@/services/message/message.query";

const IMAGE_SETTLE_TIMEOUT_MS = 2500;

let workspaceShellReadyPersisted = false;

/** Reset on logout so the next session gets the initial bootstrap overlay. */
export function resetWorkspaceShellReady(): void {
	workspaceShellReadyPersisted = false;
}

function waitForImageElement(img: HTMLImageElement): Promise<void> {
	if (img.complete) {
		if (img.naturalWidth > 0 && typeof img.decode === "function") {
			return img.decode().catch(() => undefined);
		}
		return Promise.resolve();
	}
	return new Promise((resolve) => {
		const done = () => {
			img.removeEventListener("load", onLoad);
			img.removeEventListener("error", done);
			resolve();
		};
		const onLoad = () => {
			if (typeof img.decode === "function") {
				void img
					.decode()
					.catch(() => undefined)
					.finally(done);
				return;
			}
			done();
		};
		img.addEventListener("load", onLoad, { once: true });
		img.addEventListener("error", done, { once: true });
	});
}

/**
 * Mounts the authenticated workspace shell underneath a full-screen WorkspaceLoader overlay
 * and only dismisses the overlay once active queries, DOM critical images (rail logos),
 * and the initial paint frame have settled.
 */
export function WorkspaceReadyGate({ children }: { children: ReactNode }) {
	const pathname = usePathname();
	const [ready, setReady] = useState(workspaceShellReadyPersisted);
	const containerRef = useRef<HTMLDivElement>(null);

	const target = useMemo(() => parseChannelRoute(pathname), [pathname]);
	const isRootRedirecting = pathname === "/";
	const groupId = target.groupId ?? "";
	const channelId = target.channelId ?? "";

	const myGroups = useMyGroupsQuery();
	const channelsQuery = useQuery({
		...channelQuery.groupChannelsQueryOpts(groupId),
		enabled: !ready && Boolean(groupId),
	});
	const membersQuery = useQuery({
		...groupMembersQueryOpts(groupId, ""),
		enabled: !ready && Boolean(groupId),
	});
	const messagesQuery = useInfiniteQuery({
		...channelMessagesQueryOpts(groupId, channelId),
		enabled: !ready && Boolean(groupId && channelId),
	});

	const queriesReady =
		!isRootRedirecting &&
		!myGroups.isPending &&
		(!groupId || (!channelsQuery.isPending && !membersQuery.isPending)) &&
		(!(groupId && channelId) || !messagesQuery.isPending);

	useEffect(() => {
		if (workspaceShellReadyPersisted) {
			return;
		}
		if (ready || !queriesReady) {
			return;
		}

		let cancelled = false;
		let timeoutId: ReturnType<typeof setTimeout> | undefined;
		let raf1 = 0;
		let raf2 = 0;

		const settle = async () => {
			const root = containerRef.current;
			const imgs = root
				? Array.from(root.querySelectorAll<HTMLImageElement>("img[data-critical-img], aside img"))
				: [];

			if (imgs.length > 0) {
				const imagesDone = Promise.allSettled(imgs.map(waitForImageElement));
				const timeout = new Promise<void>((resolve) => {
					timeoutId = setTimeout(resolve, IMAGE_SETTLE_TIMEOUT_MS);
				});
				await Promise.race([imagesDone, timeout]);
				if (timeoutId !== undefined) {
					clearTimeout(timeoutId);
				}
			}

			if (cancelled) {
				return;
			}

			// Double rAF guarantees the mounted shell & decoded images have painted underneath the overlay
			raf1 = requestAnimationFrame(() => {
				raf2 = requestAnimationFrame(() => {
					if (!cancelled) {
						workspaceShellReadyPersisted = true;
						setReady(true);
					}
				});
			});
		};

		void settle();

		return () => {
			cancelled = true;
			if (timeoutId !== undefined) {
				clearTimeout(timeoutId);
			}
			if (raf1) {
				cancelAnimationFrame(raf1);
			}
			if (raf2) {
				cancelAnimationFrame(raf2);
			}
		};
	}, [queriesReady, ready]);

	return (
		<>
			<div ref={containerRef} className="flex min-h-0 min-w-0 flex-1 flex-col">
				{children}
			</div>
			{!ready ? <WorkspaceLoader className="fixed inset-0 z-[100]" /> : null}
		</>
	);
}
