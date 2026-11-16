import { GroupActivityLogsPageSchema } from "@wichat/contracts/group/audit-log";
import {
	BanGroupMembersRequestSchema,
	BanGroupMembersResponseSchema,
	GroupBansPageSchema,
	TimeoutGroupMemberRequestSchema,
	TimeoutGroupMemberResponseSchema,
	type BanGroupMembersRequest,
	type TimeoutGroupMemberRequest,
} from "@wichat/contracts/group/member";
import {
	CreateGroupRequestSchema,
	GroupSchema,
	GroupsListSchema,
	ReorderMyGroupsRequestSchema,
	UpdateGroupProfileBodySchema,
	type CreateGroupRequest,
	type Group,
	type UpdateGroupProfileBody,
} from "@wichat/contracts/group";
import { OkSchema, isPublicEntityId } from "@wichat/contracts/common";
import { api } from "@/integrations/api-client";

export type MyGroup = Group;

function normalizeMembershipRole(role: string | undefined): "owner" | "member" | undefined {
	if (role === "owner") {
		return "owner";
	}
	if (role === "member" || role === "admin" || role === "guest") {
		return "member";
	}
	return undefined;
}

function normalizeMyGroup(raw: Group): MyGroup | null {
	if (!isPublicEntityId(raw.id)) {
		return null;
	}
	const role = normalizeMembershipRole(raw.role);
	if (!role) {
		return null;
	}
	return { ...raw, role };
}

export async function reorderMyGroups(groupIds: string[]) {
	const body = ReorderMyGroupsRequestSchema.parse({ group_ids: groupIds });
	return api.put("/groups/mine/order", OkSchema, body);
}

export async function listMyGroups() {
	const data = await api.get("/groups/mine", GroupsListSchema);
	const groups = data.groups.map(normalizeMyGroup).filter((g): g is MyGroup => g !== null);
	return { groups };
}

export async function createGroup(payload: CreateGroupRequest) {
	const body = CreateGroupRequestSchema.parse(payload);
	return api.post("/groups", GroupSchema, body);
}

export async function joinGroup(groupId: string) {
	return api.post(`/groups/${encodeURIComponent(groupId)}/join`, OkSchema, {});
}

export async function leaveGroup(groupId: string) {
	return api.post(`/groups/${encodeURIComponent(groupId)}/leave`, OkSchema, {});
}

export async function deleteGroup(groupId: string) {
	const id = encodeURIComponent(groupId);
	return api.post(`/groups/${id}/delete`, OkSchema, {});
}

export async function uploadGroupIcon(groupId: string, file: File) {
	const form = new FormData();
	form.append("file", file);
	return api.postForm(`/groups/${encodeURIComponent(groupId)}/icon`, GroupSchema, form);
}

export async function uploadGroupBanner(groupId: string, file: File) {
	const form = new FormData();
	form.append("file", file);
	return api.postForm(`/groups/${encodeURIComponent(groupId)}/banner`, GroupSchema, form);
}

export async function updateGroupProfile(groupId: string, body: UpdateGroupProfileBody) {
	const payload = UpdateGroupProfileBodySchema.parse(body);
	return api.patch(`/groups/${encodeURIComponent(groupId)}/profile`, GroupSchema, payload);
}

export type SaveGroupProfileInput = UpdateGroupProfileBody & {
	iconFile?: File | null;
	bannerFile?: File | null;
};

export async function saveGroupProfile(groupId: string, input: SaveGroupProfileInput) {
	const { iconFile, bannerFile, ...body } = input;
	let group = await updateGroupProfile(groupId, body);
	if (iconFile) {
		group = await uploadGroupIcon(groupId, iconFile);
	}
	if (bannerFile) {
		group = await uploadGroupBanner(groupId, bannerFile);
	}
	return group;
}

export async function kickGroupMember(groupId: string, userId: string, reason = "") {
	return api.post(
		`/groups/${encodeURIComponent(groupId)}/members/${encodeURIComponent(userId)}/kick`,
		OkSchema,
		{ reason },
	);
}

export async function banGroupMembers(groupId: string, input: BanGroupMembersRequest) {
	const payload = BanGroupMembersRequestSchema.parse(input);
	return api.post(
		`/groups/${encodeURIComponent(groupId)}/bans`,
		BanGroupMembersResponseSchema,
		payload,
	);
}

export async function timeoutGroupMember(
	groupId: string,
	userId: string,
	input: TimeoutGroupMemberRequest,
) {
	const payload = TimeoutGroupMemberRequestSchema.parse(input);
	return api.put(
		`/groups/${encodeURIComponent(groupId)}/members/${encodeURIComponent(userId)}/timeout`,
		TimeoutGroupMemberResponseSchema,
		payload,
	);
}

export async function listGroupBans(
	groupId: string,
	params?: { search?: string; limit?: number; cursor?: string },
) {
	const qs = new URLSearchParams();
	if (params?.search?.trim()) {
		qs.set("search", params.search.trim());
	}
	if (params?.limit) {
		qs.set("limit", String(params.limit));
	}
	if (params?.cursor) {
		qs.set("cursor", params.cursor);
	}
	const query = qs.toString();
	return api.get(
		`/groups/${encodeURIComponent(groupId)}/bans${query ? `?${query}` : ""}`,
		GroupBansPageSchema,
	);
}

export async function unbanGroupMember(groupId: string, userId: string) {
	return api.delete(
		`/groups/${encodeURIComponent(groupId)}/bans/${encodeURIComponent(userId)}`,
		OkSchema,
	);
}

export async function listGroupAuditLogs(
	groupId: string,
	params?: {
		actorId?: string;
		category?: string;
		action?: string;
		limit?: number;
		cursor?: string;
	},
) {
	const qs = new URLSearchParams();
	if (params?.actorId?.trim()) {
		qs.set("actor_id", params.actorId.trim());
	}
	if (params?.category?.trim()) {
		qs.set("category", params.category.trim());
	}
	if (params?.action?.trim()) {
		qs.set("action", params.action.trim());
	}
	if (params?.limit) {
		qs.set("limit", String(params.limit));
	}
	if (params?.cursor) {
		qs.set("cursor", params.cursor);
	}
	const query = qs.toString();
	return api.get(
		`/groups/${encodeURIComponent(groupId)}/audit-logs${query ? `?${query}` : ""}`,
		GroupActivityLogsPageSchema,
	);
}

export type { BanGroupMembersRequest, Group, TimeoutGroupMemberRequest, UpdateGroupProfileBody };
