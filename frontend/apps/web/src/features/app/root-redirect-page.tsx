"use client";

import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { useTranslation } from "react-i18next";
import { readLastChannel } from "@/lib/last-channel-storage";
import { resolveAppDestination } from "@/lib/resolve-app-destination";
import { useMyGroupsQuery } from "@/hooks/use-my-groups-query";

/** `/` — no UI; send signed-in users to `/home` or their last channel. */
export function RootRedirectPage() {
	const { t } = useTranslation();
	const router = useRouter();
	const myGroups = useMyGroupsQuery();

	useEffect(() => {
		if (myGroups.isPending) {
			return;
		}
		const groups = myGroups.isError ? [] : (myGroups.data ?? []);
		const dest = resolveAppDestination(groups, readLastChannel());
		router.replace(dest);
	}, [myGroups.data, myGroups.isError, myGroups.isPending, router]);

	return (
		<div
			className="flex flex-1 items-center justify-center bg-background text-sm text-muted-foreground"
			suppressHydrationWarning
		>
			{t("group.home.loading")}
		</div>
	);
}
