import {
	CreateGroupRequestSchema,
	GroupSchema,
	GroupsListSchema,
	OkSchema,
	ReorderMyGroupsRequestSchema,
	UpdateGroupProfileBodySchema,
	type CreateGroupRequest,
	type Group,
	type UpdateGroupProfileBody,
	isPublicEntityId,
} from "@wichat/contracts";
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

export type { Group, UpdateGroupProfileBody };
