"use client";

import { isPublicEntityId } from "@wichat/contracts";
import { useRouter } from "next/navigation";
import { useEffect, type ReactNode } from "react";
import { useMyGroupsQuery } from "@/hooks/use-my-groups-query";
import { APP_HOME_PATH } from "@/lib/app-routes";

type Props = {
	groupId: string;
	children: ReactNode;
};

/** Redirect-only guard for group-scoped workspace routes (quick views) — no channel validation. */
export function GroupRouteGuard({ groupId, children }: Props) {
	const router = useRouter();
	const myGroups = useMyGroupsQuery();

	useEffect(() => {
		if (!isPublicEntityId(groupId)) {
			router.replace(APP_HOME_PATH);
			return;
		}
		if (myGroups.isPending) {
			return;
		}
		const member = myGroups.data?.some((g) => g.id === groupId);
		if (!member) {
			router.replace(APP_HOME_PATH);
		}
	}, [groupId, myGroups.data, myGroups.isPending, router]);

	return children;
}
