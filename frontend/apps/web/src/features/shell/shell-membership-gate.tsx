"use client";

import { useRouter } from "next/navigation";
import { useEffect, type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { useMyGroupsQuery } from "@/hooks/use-my-groups-query";
import { APP_HOME_PATH } from "@/lib/app-routes";

type Props = {
	groupId: string;
	children: ReactNode;
};

export function ShellMembershipGate({ groupId, children }: Props) {
	const { t } = useTranslation();
	const router = useRouter();
	const myGroups = useMyGroupsQuery();

	useEffect(() => {
		if (myGroups.isPending && !myGroups.data) {
			return;
		}
		const groups = myGroups.isError ? [] : (myGroups.data ?? []);
		const allowed = groups.some((g) => g.id === groupId);
		if (!allowed) {
			router.replace(APP_HOME_PATH);
		}
	}, [groupId, myGroups.data, myGroups.isError, myGroups.isPending, router]);

	const groups = myGroups.isError ? [] : (myGroups.data ?? []);

	if (myGroups.isPending && groups.length === 0) {
		return (
			<div
				className="flex flex-1 items-center justify-center bg-background text-sm text-muted-foreground"
				suppressHydrationWarning
			>
				{t("shell.loading")}
			</div>
		);
	}

	if (!groups.some((g) => g.id === groupId)) {
		return (
			<div
				className="flex flex-1 items-center justify-center bg-background text-sm text-muted-foreground"
				suppressHydrationWarning
			>
				{t("shell.loading")}
			</div>
		);
	}

	return children;
}
