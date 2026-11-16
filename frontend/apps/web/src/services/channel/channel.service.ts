import {
	ChannelsListSchema,
	CreateChannelCategoryRequestSchema,
	CreateChannelCategoryResponseSchema,
	CreateGroupChannelRequestSchema,
	CreateGroupChannelResponseSchema,
	ReorderChannelCategoriesRequestSchema,
	ReorderGroupChannelsRequestSchema,
	type ChannelsList,
	type CreateChannelCategoryRequest,
	type CreateGroupChannelRequest,
} from "@wichat/contracts/chat/channel";
import { OkSchema } from "@wichat/contracts/common";
import { api } from "@/integrations/api-client";

export async function listGroupChannels(groupId: string): Promise<ChannelsList> {
	return api.get(`/groups/${encodeURIComponent(groupId)}/channels`, ChannelsListSchema);
}

export async function createGroupCategory(groupId: string, payload: CreateChannelCategoryRequest) {
	const body = CreateChannelCategoryRequestSchema.parse(payload);
	const data = await api.post(
		`/groups/${encodeURIComponent(groupId)}/categories`,
		CreateChannelCategoryResponseSchema,
		body,
	);
	return data.category;
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

export async function reorderGroupCategories(groupId: string, categoryIds: string[]) {
	const body = ReorderChannelCategoriesRequestSchema.parse({ category_ids: categoryIds });
	return api.put(`/groups/${encodeURIComponent(groupId)}/categories/order`, OkSchema, body);
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
