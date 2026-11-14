"use client";

import { useRouter } from "next/navigation";
import { useEffect, useRef } from "react";
import { WorkspaceLoader } from "@/components/layout/workspace-loader";
import { readLastChannel } from "@/lib/last-channel-storage";
import { resolveAppDestination } from "@/lib/resolve-app-destination";
import { useMyGroupsQuery } from "@/hooks/use-my-groups-query";

/** `/` — no UI; send signed-in users to `/home` or their last channel. */
export function RootRedirectPage() {
	const router = useRouter();
	const myGroups = useMyGroupsQuery();
	const redirectedRef = useRef(false);

	useEffect(() => {
		if (redirectedRef.current || myGroups.isPending) {
			return;
		}
		redirectedRef.current = true;
		const groups = myGroups.isError ? [] : (myGroups.data ?? []);
		const dest = resolveAppDestination(groups, readLastChannel());
		router.replace(dest);
	}, [myGroups.data, myGroups.isError, myGroups.isPending, router]);

	return <WorkspaceLoader labelKey="group.home.loading" />;
}
