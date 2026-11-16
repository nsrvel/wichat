"use client";

import { useQuery } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect, useSyncExternalStore, type ReactNode } from "react";
import { WorkspaceLoader } from "@/components/layout/workspace-loader";
import { useMyGroupsQuery } from "@/hooks/use-my-groups-query";
import {
	getGatewayStatus,
	subscribeGatewayStatus,
} from "@/integrations/realtime/gateway-connection";
import { APP_HOME_PATH } from "@/lib/app-routes";
import * as channelQuery from "@/services/channel/channel.query";
import { groupMembersQueryOpts } from "@/services/invite/invite.query";

const getServerSnapshot = () => "connected" as const;

type Props = {
	groupId: string;
	children: ReactNode;
};

export function ShellMembershipGate({ groupId, children }: Props) {
	const router = useRouter();
	const myGroups = useMyGroupsQuery();
	const channelsQuery = useQuery(channelQuery.groupChannelsQueryOpts(groupId));
	const membersQuery = useQuery({
		...groupMembersQueryOpts(groupId, ""),
		enabled: Boolean(groupId),
	});
	const gatewayStatus = useSyncExternalStore(
		subscribeGatewayStatus,
		getGatewayStatus,
		getServerSnapshot,
	);

	useEffect(() => {
		if (!myGroups.isSuccess || gatewayStatus === "reconnecting") {
			return;
		}
		const allowed = myGroups.data.some((g) => g.id === groupId);
		if (!allowed) {
			router.replace(APP_HOME_PATH);
		}
	}, [gatewayStatus, groupId, myGroups.data, myGroups.isSuccess, router]);

	const groups = myGroups.data ?? [];

	if ((myGroups.isPending || myGroups.isError) && groups.length === 0) {
		return <WorkspaceLoader />;
	}

	if (!groups.some((g) => g.id === groupId)) {
		return <WorkspaceLoader />;
	}

	const showLoadingOverlay =
		(channelsQuery.isPending && !channelsQuery.data) ||
		(membersQuery.isPending && !membersQuery.data);

	return (
		<>
			{children}
			{showLoadingOverlay ? <WorkspaceLoader className="fixed inset-0 z-[100]" /> : null}
		</>
	);
}
