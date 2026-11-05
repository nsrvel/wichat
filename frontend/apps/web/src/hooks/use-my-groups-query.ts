"use client";

import { useQuery } from "@tanstack/react-query";
import { useUiSampleChatRoute } from "@/hooks/use-ui-sample-chat-route";
import * as groupQuery from "@/services/group/group.query";

function useMyGroupsQueryOpts() {
	const uiSample = useUiSampleChatRoute();
	return uiSample ? groupQuery.myGroupsUiSampleQueryOpts() : groupQuery.myGroupsQueryOpts();
}

export function useMyGroupsQuery() {
	return useQuery(useMyGroupsQueryOpts());
}
