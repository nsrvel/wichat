import {
	CreateGroupRequestSchema,
	GroupSchema,
	GroupsListSchema,
	OkSchema,
	type CreateGroupRequest,
	type Group,
} from "@wichat/contracts";
import { api } from "@/integrations/api-client";

export async function listMyGroups() {
	return api.get("/groups/mine", GroupsListSchema);
}

export async function listJoinableGroups() {
	return api.get("/groups/joinable", GroupsListSchema);
}

export async function createGroup(payload: CreateGroupRequest) {
	const body = CreateGroupRequestSchema.parse(payload);
	return api.post("/groups", GroupSchema, body);
}

export async function joinGroup(groupId: string) {
	return api.post(`/groups/${encodeURIComponent(groupId)}/join`, OkSchema, {});
}

export type { Group };
