import {
	ChannelsListSchema,
	CreateGroupChannelRequestSchema,
	CreateGroupChannelResponseSchema,
	EnsureGroupDmRequestSchema,
	EnsureGroupDmResponseSchema,
	ReorderGroupChannelsRequestSchema,
	UpdateGroupChannelRequestSchema,
	UpdateGroupChannelResponseSchema,
	type ChannelsList,
	type CreateGroupChannelRequest,
	type EnsureGroupDmResponse,
	type UpdateGroupChannelRequest,
} from "@wichat/contracts/chat/channel";
import {
	CategoryPermissionsListSchema,
	UpsertCategoryPermissionOverrideRequestSchema,
	UpsertCategoryPermissionOverrideResponseSchema,
	type CategoryPermissionTargetType,
	type CategoryPermissionsList,
	type UpsertCategoryPermissionOverrideRequest,
} from "@wichat/contracts/chat/channel-category";
import { OkSchema } from "@wichat/contracts/common";
import { api } from "@/integrations/api-client";

export async function listGroupChannels(groupId: string): Promise<ChannelsList> {
	return api.get(`/groups/${encodeURIComponent(groupId)}/channels`, ChannelsListSchema);
}

export async function ensureGroupDmChannel(
	groupId: string,
	userId: string,
): Promise<EnsureGroupDmResponse> {
	const body = EnsureGroupDmRequestSchema.parse({ user_id: userId });
	return api.post(`/groups/${encodeURIComponent(groupId)}/dm`, EnsureGroupDmResponseSchema, body);
}

export async function createGroupChannel(groupId: string, payload: CreateGroupChannelRequest) {
	const body = CreateGroupChannelRequestSchema.parse(payload);
	const data = await api.post(
		`/groups/${encodeURIComponent(groupId)}/channels`,
		CreateGroupChannelResponseSchema,
		body,
	);
	return data.channel;
}

export async function updateGroupChannel(
	groupId: string,
	channelId: string,
	payload: UpdateGroupChannelRequest,
) {
	const body = UpdateGroupChannelRequestSchema.parse(payload);
	const data = await api.patch(
		`/groups/${encodeURIComponent(groupId)}/channels/${encodeURIComponent(channelId)}`,
		UpdateGroupChannelResponseSchema,
		body,
	);
	return data.channel;
}

export async function deleteGroupChannel(groupId: string, channelId: string) {
	return api.delete(
		`/groups/${encodeURIComponent(groupId)}/channels/${encodeURIComponent(channelId)}`,
		OkSchema,
	);
}

export async function listChannelPermissions(
	groupId: string,
	channelId: string,
): Promise<CategoryPermissionsList> {
	return api.get(
		`/groups/${encodeURIComponent(groupId)}/channels/${encodeURIComponent(channelId)}/permissions`,
		CategoryPermissionsListSchema,
	);
}

export async function upsertChannelPermission(
	groupId: string,
	channelId: string,
	payload: UpsertCategoryPermissionOverrideRequest,
) {
	const body = UpsertCategoryPermissionOverrideRequestSchema.parse(payload);
	const data = await api.put(
		`/groups/${encodeURIComponent(groupId)}/channels/${encodeURIComponent(channelId)}/permissions`,
		UpsertCategoryPermissionOverrideResponseSchema,
		body,
	);
	return data.override;
}

export async function deleteChannelPermission(
	groupId: string,
	channelId: string,
	targetType: CategoryPermissionTargetType,
	targetId: string,
) {
	return api.delete(
		`/groups/${encodeURIComponent(groupId)}/channels/${encodeURIComponent(channelId)}/permissions/${encodeURIComponent(targetType)}/${encodeURIComponent(targetId)}`,
		OkSchema,
	);
}

export async function syncChannelPermissions(groupId: string, channelId: string) {
	const data = await api.post(
		`/groups/${encodeURIComponent(groupId)}/channels/${encodeURIComponent(channelId)}/permissions/sync`,
		UpdateGroupChannelResponseSchema,
		{},
	);
	return data.channel;
}

export async function reorderGroupChannels(
	groupId: string,
	payload: { channelIds: string[]; categoryId?: string; sidebarSection?: "favorites" | "channels" },
) {
	const body = ReorderGroupChannelsRequestSchema.parse({
		channel_ids: payload.channelIds,
		category_id: payload.categoryId || undefined,
		sidebar_section: payload.sidebarSection,
	});
	return api.put(`/groups/${encodeURIComponent(groupId)}/channels/order`, OkSchema, body);
}
